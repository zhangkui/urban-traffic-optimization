package priority

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"sort"
	"sync"
	"time"
)

type Request struct {
	ID             string    `json:"id"`
	IntersectionID string    `json:"intersectionId"`
	Direction      string    `json:"direction"`
	Vehicle        string    `json:"vehicle"`
	Priority       string    `json:"priority"`
	RequestedAt    time.Time `json:"requestedAt"`
	ExpiresAt      time.Time `json:"expiresAt"`
	Status         string    `json:"status"`
}
type Decision struct {
	RequestID   string    `json:"requestId"`
	Granted     bool      `json:"granted"`
	HoldSeconds int       `json:"holdSeconds"`
	Reason      string    `json:"reason"`
	DecidedAt   time.Time `json:"decidedAt"`
}
type Service struct {
	mu       sync.Mutex
	requests map[string]Request
	maxHold  int
}

func NewService() *Service { return &Service{requests: map[string]Request{}, maxHold: 30} }
func (s *Service) Submit(request Request) error {
	if request.ID == "" || request.IntersectionID == "" || request.Direction == "" {
		return errors.New("priority request identity is incomplete")
	}
	if request.Priority == "" {
		request.Priority = "normal"
	}
	if request.RequestedAt.IsZero() {
		request.RequestedAt = time.Now().UTC()
	}
	if request.ExpiresAt.IsZero() {
		request.ExpiresAt = request.RequestedAt.Add(2 * time.Minute)
	}
	if !request.ExpiresAt.After(request.RequestedAt) {
		return errors.New("priority request expiry is invalid")
	}
	request.Status = "pending"
	s.mu.Lock()
	s.requests[request.ID] = request
	s.mu.Unlock()
	return nil
}
func (s *Service) Decide(id string, now time.Time) Decision {
	s.mu.Lock()
	defer s.mu.Unlock()
	request, ok := s.requests[id]
	decision := Decision{RequestID: id, DecidedAt: now}
	if !ok {
		decision.Reason = "request not found"
		return decision
	}
	if request.Status != "pending" {
		decision.Reason = "request is not pending"
		return decision
	}
	if now.After(request.ExpiresAt) {
		request.Status = "expired"
		s.requests[id] = request
		decision.Reason = "request expired"
		return decision
	}
	decision.Granted = true
	decision.HoldSeconds = s.maxHold
	decision.Reason = "priority granted"
	if request.Priority == "emergency" {
		decision.HoldSeconds = s.maxHold * 2
	}
	request.Status = "granted"
	s.requests[id] = request
	return decision
}
func (s *Service) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	request, ok := s.requests[id]
	if !ok {
		return errors.New("priority request not found")
	}
	if request.Status != "pending" {
		return errors.New("only pending request can be cancelled")
	}
	request.Status = "cancelled"
	s.requests[id] = request
	return nil
}
func (s *Service) Pending(now time.Time) []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := []Request{}
	for id, request := range s.requests {
		if request.Status == "pending" && now.After(request.ExpiresAt) {
			request.Status = "expired"
			s.requests[id] = request
			continue
		}
		if request.Status == "pending" {
			items = append(items, request)
		}
	}
	sort.Slice(items, func(i, j int) bool { return priorityValue(items[i].Priority) > priorityValue(items[j].Priority) })
	return items
}
func priorityValue(value string) int {
	switch value {
	case "emergency":
		return 4
	case "public_transport":
		return 3
	case "rescue":
		return 2
	default:
		return 1
	}
}
func (s *Service) SetMaxHold(seconds int) error {
	if seconds < 5 || seconds > 120 {
		return errors.New("max hold must be between 5 and 120 seconds")
	}
	s.mu.Lock()
	s.maxHold = seconds
	s.mu.Unlock()
	return nil
}
func (s *Service) Get(id string) (Request, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.requests[id]
	return item, ok
}
func Emergency(event domain.TrafficEvent) Request {
	return Request{ID: "event-" + event.ID, IntersectionID: event.IntersectionID, Direction: "all", Vehicle: "emergency", Priority: "emergency", RequestedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(time.Minute)}
}
