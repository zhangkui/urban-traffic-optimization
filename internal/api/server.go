package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhangkui/urban-traffic-optimization/internal/adaptive"
	"github.com/zhangkui/urban-traffic-optimization/internal/calendar"
	"github.com/zhangkui/urban-traffic-optimization/internal/corridor"
	"github.com/zhangkui/urban-traffic-optimization/internal/geofence"
	"github.com/zhangkui/urban-traffic-optimization/internal/maintenance"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zhangkui/urban-traffic-optimization/internal/dashboard"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/event"
	"github.com/zhangkui/urban-traffic-optimization/internal/optimization"
	"github.com/zhangkui/urban-traffic-optimization/internal/realtime"
	"github.com/zhangkui/urban-traffic-optimization/internal/road"
	"github.com/zhangkui/urban-traffic-optimization/internal/signal"
	"github.com/zhangkui/urban-traffic-optimization/internal/simulation"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"github.com/zhangkui/urban-traffic-optimization/internal/traffic"
)

type Server struct {
	store         *store.Store
	hub           *realtime.Hub
	roads         *road.Service
	signals       *signal.Service
	traffic       *traffic.Service
	events        *event.Service
	simulations   *simulation.Service
	optimizations *optimization.Service
	dashboard     *dashboard.Service
	corridors     *corridor.Service
	adaptive      *adaptive.Service
	maintenance   *maintenance.Service
	zones         *geofence.Service
	calendars     *calendar.Service
	ids           atomic.Uint64
}

func New(s *store.Store, h *realtime.Hub) *Server {
	return &Server{store: s, hub: h, roads: road.NewService(s), signals: signal.NewService(s), traffic: traffic.NewService(s), events: event.NewService(s), simulations: simulation.NewService(s), optimizations: optimization.NewService(s), dashboard: dashboard.NewService(s), corridors: corridor.NewService(), adaptive: adaptive.NewService(), maintenance: maintenance.NewService(), zones: geofence.NewService(), calendars: calendar.NewService()}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/api/v1/auth/login", s.login)
	mux.HandleFunc("/api/v1/dashboard/summary", s.summary)
	mux.HandleFunc("/api/v1/roads", s.roadsHandler)
	mux.HandleFunc("/api/v1/intersections", s.genericCollection("intersections"))
	mux.HandleFunc("/api/v1/timing-plans", s.timingHandler)
	mux.HandleFunc("/api/v1/traffic/readings", s.trafficHandler)
	mux.HandleFunc("/api/v1/events", s.eventsHandler)
	mux.HandleFunc("/api/v1/simulations", s.simulationHandler)
	mux.HandleFunc("/api/v1/optimization/candidates", s.optimizationHandler)
	mux.HandleFunc("/api/v1/corridors", s.corridorHandler)
	mux.HandleFunc("/api/v1/geofences", s.geofenceHandler)
	mux.HandleFunc("/api/v1/maintenance/inspections", s.inspectionHandler)
	mux.HandleFunc("/api/v1/signal/adaptive/decisions", s.adaptiveHandler)
	mux.Handle("/api/v1/stream", s.hub)
	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": value})
}
func fail(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": status, "message": err.Error()}})
}
func decode(r *http.Request, value any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	if err := json.NewDecoder(r.Body).Decode(value); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, map[string]string{"status": "ok", "service": "urban-traffic-optimization"})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, errors.New("method not allowed"))
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decode(r, &input); err != nil || input.Username == "" || input.Password == "" {
		fail(w, 400, errors.New("username and password are required"))
		return
	}
	write(w, 200, map[string]any{"token": "local-" + input.Username, "expiresIn": 7200, "user": domain.User{ID: "u-1", Username: input.Username, DisplayName: input.Username, RoleIDs: []string{"traffic-admin"}, Enabled: true}})
}
func (s *Server) summary(w http.ResponseWriter, _ *http.Request) {
	write(w, 200, s.dashboard.Summary())
}
func pageRequest(r *http.Request) domain.PageRequest {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	return domain.PageRequest{Page: p, Size: size, Keyword: r.URL.Query().Get("keyword"), Status: r.URL.Query().Get("status")}
}
func (s *Server) roadsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		result, err := s.roads.List(pageRequest(r))
		if err != nil {
			fail(w, 500, err)
			return
		}
		write(w, 200, result)
	case http.MethodPost:
		var item domain.Road
		if err := decode(r, &item); err != nil {
			fail(w, 400, err)
			return
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("road-%06d", s.ids.Add(1))
		}
		if err := s.roads.Create(item); err != nil {
			fail(w, 400, err)
			return
		}
		s.hub.Publish("road.created", item)
		write(w, 201, item)
	default:
		fail(w, 405, errors.New("method not allowed"))
	}
}
func (s *Server) timingHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		write(w, 200, s.signals.List(r.URL.Query().Get("intersectionId")))
	case http.MethodPost:
		var item domain.TimingPlan
		if err := decode(r, &item); err != nil {
			fail(w, 400, err)
			return
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("plan-%06d", s.ids.Add(1))
		}
		if err := s.signals.Create(item); err != nil {
			fail(w, 400, err)
			return
		}
		s.hub.Publish("timing-plan.created", item)
		write(w, 201, item)
	case http.MethodPatch:
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/timing-plans/")
		action := r.URL.Query().Get("action")
		var err error
		switch action {
		case "submit":
			err = s.signals.SubmitReview(id)
		case "approve":
			err = s.signals.Approve(id)
		case "publish":
			err = s.signals.Publish(id)
		default:
			err = errors.New("unsupported timing plan action")
		}
		if err != nil {
			fail(w, 400, err)
			return
		}
		write(w, 200, map[string]string{"id": id, "action": action})
	default:
		fail(w, 405, errors.New("method not allowed"))
	}
}
func (s *Server) trafficHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("intersectionId")
		since := time.Now().Add(-time.Hour)
		write(w, 200, s.traffic.Recent(id, since))
	case http.MethodPost:
		var item domain.TrafficReading
		if err := decode(r, &item); err != nil {
			fail(w, 400, err)
			return
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("reading-%06d", s.ids.Add(1))
		}
		if err := s.traffic.Record(item); err != nil {
			fail(w, 400, err)
			return
		}
		s.hub.Publish("traffic.reading", item)
		write(w, 201, item)
	default:
		fail(w, 405, errors.New("method not allowed"))
	}
}
func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		write(w, 200, s.events.List(r.URL.Query().Get("status")))
	case http.MethodPost:
		var item domain.TrafficEvent
		if err := decode(r, &item); err != nil {
			fail(w, 400, err)
			return
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("event-%06d", s.ids.Add(1))
		}
		if err := s.events.Open(item); err != nil {
			fail(w, 400, err)
			return
		}
		s.hub.Publish("event.created", item)
		write(w, 201, item)
	case http.MethodPatch:
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/events/")
		var input struct {
			Status string `json:"status"`
		}
		if err := decode(r, &input); err != nil {
			fail(w, 400, err)
			return
		}
		if err := s.events.Transition(id, input.Status); err != nil {
			fail(w, 400, err)
			return
		}
		write(w, 200, map[string]string{"id": id, "status": input.Status})
	default:
		fail(w, 405, errors.New("method not allowed"))
	}
}
func (s *Server) simulationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/simulations/")
		if id != "" && id != r.URL.Path {
			item, err := s.simulations.Get(id)
			if err != nil {
				fail(w, 404, err)
				return
			}
			write(w, 200, item)
			return
		}
		write(w, 200, []domain.SimulationTask{})
		return
	}
	if r.Method != http.MethodPost {
		fail(w, 405, errors.New("method not allowed"))
		return
	}
	var item domain.SimulationTask
	if err := decode(r, &item); err != nil {
		fail(w, 400, err)
		return
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("sim-%06d", s.ids.Add(1))
	}
	if err := s.simulations.Create(item); err != nil {
		fail(w, 400, err)
		return
	}
	go func(id string) {
		_, _ = s.simulations.Run(context.Background(), id, 1000, 1200)
		s.hub.Publish("simulation.completed", map[string]string{"id": id})
	}(item.ID)
	write(w, 202, item)
}
func (s *Server) optimizationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, errors.New("method not allowed"))
		return
	}
	var input struct {
		TaskID     string `json:"taskId"`
		BaseCycle  int    `json:"baseCycle"`
		BaseOffset int    `json:"baseOffset"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, 400, err)
		return
	}
	candidates := s.optimizations.Generate(input.TaskID, input.BaseCycle, input.BaseOffset)
	write(w, 200, map[string]any{"items": s.optimizations.Rank(candidates), "total": len(candidates)})
}
func (s *Server) genericCollection(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var values []json.RawMessage
			if err := s.store.List(kind, func(raw []byte) error { values = append(values, json.RawMessage(raw)); return nil }); err != nil {
				fail(w, 500, err)
				return
			}
			write(w, 200, map[string]any{"items": values, "total": len(values)})
			return
		}
		if r.Method != http.MethodPost {
			fail(w, 405, errors.New("method not allowed"))
			return
		}
		var value map[string]any
		if err := decode(r, &value); err != nil {
			fail(w, 400, err)
			return
		}
		id, _ := value["id"].(string)
		if id == "" {
			id = fmt.Sprintf("%s-%06d", strings.TrimSuffix(kind, "s"), s.ids.Add(1))
			value["id"] = id
		}
		value["updatedAt"] = time.Now().UTC()
		if err := s.store.Save(kind, id, value); err != nil {
			fail(w, 500, err)
			return
		}
		s.hub.Publish(kind+".updated", value)
		write(w, 201, value)
	}
}

func (s *Server) corridorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		write(w, http.StatusOK, s.corridors.List(r.URL.Query().Get("status")))
		return
	}
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var item corridor.Corridor
	if err := decode(r, &item); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("corridor-%06d", s.ids.Add(1))
	}
	if err := s.corridors.Create(item); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusCreated, item)
}

func (s *Server) geofenceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		write(w, http.StatusOK, map[string]any{"zones": s.zones.Zones(), "rules": s.zones.Rules(r.URL.Query().Get("zoneId"))})
		return
	}
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var item geofence.Zone
	if err := decode(r, &item); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("zone-%06d", s.ids.Add(1))
	}
	if err := s.zones.AddZone(item); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusCreated, item)
}

func (s *Server) inspectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		write(w, http.StatusOK, s.maintenance.List(r.URL.Query().Get("deviceId"), r.URL.Query().Get("status")))
		return
	}
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var input struct {
		Device    domain.Device `json:"device"`
		Inspector string        `json:"inspector"`
		At        time.Time     `json:"at"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusCreated, s.maintenance.Schedule(input.Device, input.Inspector, input.At))
}

func (s *Server) adaptiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var input struct {
		Controller adaptive.Controller     `json:"controller"`
		Phases     []domain.SignalPhase    `json:"phases"`
		Readings   []domain.TrafficReading `json:"readings"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := s.adaptive.Register(input.Controller); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	controller, _ := s.adaptive.Get(input.Controller.ID)
	write(w, http.StatusOK, s.adaptive.Decide(controller, input.Phases, input.Readings))
}
