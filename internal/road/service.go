package road

import (
	"encoding/json"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"sort"
	"strings"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Validate(road domain.Road) error {
	if err := shared.Require(road.ID, "road id"); err != nil {
		return err
	}
	if err := shared.Require(road.Name, "road name"); err != nil {
		return err
	}
	if err := shared.Require(road.Code, "road code"); err != nil {
		return err
	}
	if err := shared.OneOf(road.Level, "road level", "expressway", "arterial", "secondary", "branch"); err != nil {
		return err
	}
	if err := shared.RangeInt(road.Lanes, "lanes", 1, 20); err != nil {
		return err
	}
	if err := shared.RangeInt(road.SpeedLimit, "speed limit", 10, 130); err != nil {
		return err
	}
	if err := shared.RangeFloat(road.LengthMeters, "length meters", 1, 100000); err != nil {
		return err
	}
	return shared.OneOf(road.Status, "road status", "active", "inactive", "maintenance")
}
func (s *Service) Create(road domain.Road) error {
	if road.Status == "" {
		road.Status = "active"
	}
	if err := s.Validate(road); err != nil {
		return err
	}
	return s.store.Save("roads", road.ID, road)
}
func (s *Service) Update(road domain.Road) error {
	var old domain.Road
	if err := s.store.Load("roads", road.ID, &old); err != nil {
		return err
	}
	if err := s.Validate(road); err != nil {
		return err
	}
	return s.store.Save("roads", road.ID, road)
}
func (s *Service) Get(id string) (domain.Road, error) {
	var road domain.Road
	err := s.store.Load("roads", id, &road)
	return road, err
}
func (s *Service) Delete(id string) error { return s.store.Delete("roads", id) }
func (s *Service) List(request domain.PageRequest) (domain.Page[domain.Road], error) {
	request = request.Normalize()
	roads := make([]domain.Road, 0)
	err := s.store.List("roads", func(raw []byte) error {
		var road domain.Road
		if err := json.Unmarshal(raw, &road); err != nil {
			return err
		}
		if request.Keyword != "" && !strings.Contains(strings.ToLower(road.Name), strings.ToLower(request.Keyword)) {
			return nil
		}
		if request.Status != "" && road.Status != request.Status {
			return nil
		}
		roads = append(roads, road)
		return nil
	})
	if err != nil {
		return domain.Page[domain.Road]{}, err
	}
	sort.Slice(roads, func(i, j int) bool { return roads[i].Name < roads[j].Name })
	return domain.MakePage(roads, len(roads), request), nil
}
func (s *Service) AddLane(lane domain.RoadLane) error {
	if err := shared.Require(lane.ID, "lane id"); err != nil {
		return err
	}
	if err := shared.Require(lane.RoadID, "road id"); err != nil {
		return err
	}
	if err := shared.RangeInt(lane.Number, "lane number", 1, 20); err != nil {
		return err
	}
	if err := shared.OneOf(lane.Direction, "direction", "north", "south", "east", "west"); err != nil {
		return err
	}
	lane.Allowed = true
	return s.store.Save("road_lanes", lane.ID, lane)
}
func (s *Service) BuildTopology() map[string][]string {
	graph := map[string][]string{}
	_ = s.store.List("roads", func(raw []byte) error {
		var road domain.Road
		if json.Unmarshal(raw, &road) == nil {
			graph[road.ID] = shared.UniqueStrings(road.Downstream)
		}
		return nil
	})
	return graph
}
func (s *Service) Reachable(start string) []string {
	graph := s.BuildTopology()
	visited := map[string]bool{}
	queue := []string{start}
	result := []string{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		if current != start {
			result = append(result, current)
		}
		queue = append(queue, graph[current]...)
	}
	return result
}
