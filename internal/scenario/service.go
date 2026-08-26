package scenario

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
)

type Scenario struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IntersectionIDs []string  `json:"intersectionIds"`
	RoadIDs         []string  `json:"roadIds"`
	PlanIDs         []string  `json:"planIds"`
	FlowMultiplier  float64   `json:"flowMultiplier"`
	SpeedMultiplier float64   `json:"speedMultiplier"`
	Weather         string    `json:"weather"`
	Visibility      int       `json:"visibility"`
	Status          string    `json:"status"`
	CreatedBy       string    `json:"createdBy"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Result struct {
	ScenarioID     string    `json:"scenarioId"`
	IntersectionID string    `json:"intersectionId"`
	BaselineDelay  float64   `json:"baselineDelay"`
	ScenarioDelay  float64   `json:"scenarioDelay"`
	BaselineQueue  float64   `json:"baselineQueue"`
	ScenarioQueue  float64   `json:"scenarioQueue"`
	Throughput     float64   `json:"throughput"`
	Improvement    float64   `json:"improvement"`
	CalculatedAt   time.Time `json:"calculatedAt"`
}

type Service struct {
	mu        sync.RWMutex
	scenarios map[string]Scenario
	results   map[string][]Result
}

func NewService() *Service {
	return &Service{scenarios: map[string]Scenario{}, results: map[string][]Result{}}
}

func (s *Service) Validate(item Scenario) error {
	if strings.TrimSpace(item.ID) == "" {
		return errors.New("scenario id is required")
	}
	if strings.TrimSpace(item.Name) == "" {
		return errors.New("scenario name is required")
	}
	if len(item.IntersectionIDs) == 0 {
		return errors.New("at least one intersection is required")
	}
	if item.FlowMultiplier <= 0 || item.FlowMultiplier > 5 {
		return errors.New("flow multiplier must be between 0 and 5")
	}
	if item.SpeedMultiplier <= 0 || item.SpeedMultiplier > 2 {
		return errors.New("speed multiplier must be between 0 and 2")
	}
	if item.Visibility < 0 || item.Visibility > 10000 {
		return errors.New("visibility is outside supported range")
	}
	return nil
}

func (s *Service) Create(item Scenario) error {
	if item.FlowMultiplier == 0 {
		item.FlowMultiplier = 1
	}
	if item.SpeedMultiplier == 0 {
		item.SpeedMultiplier = 1
	}
	if item.Status == "" {
		item.Status = "draft"
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	item.UpdatedAt = time.Now().UTC()
	if err := s.Validate(item); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.scenarios[item.ID]; exists {
		return errors.New("scenario already exists")
	}
	s.scenarios[item.ID] = item
	return nil
}

func (s *Service) Update(item Scenario) error {
	if err := s.Validate(item); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.scenarios[item.ID]; !exists {
		return errors.New("scenario not found")
	}
	item.UpdatedAt = time.Now().UTC()
	s.scenarios[item.ID] = item
	return nil
}

func (s *Service) Get(id string) (Scenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.scenarios[id]
	if !ok {
		return Scenario{}, errors.New("scenario not found")
	}
	return item, nil
}

func (s *Service) List(status string) []Scenario {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Scenario, 0, len(s.scenarios))
	for _, item := range s.scenarios {
		if status == "" || item.Status == status {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	return items
}

func (s *Service) Activate(id string) error { return s.transition(id, "active") }
func (s *Service) Archive(id string) error  { return s.transition(id, "archived") }
func (s *Service) Complete(id string) error { return s.transition(id, "completed") }

func (s *Service) transition(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.scenarios[id]
	if !ok {
		return errors.New("scenario not found")
	}
	allowed := map[string]map[string]bool{
		"draft":     {"active": true, "archived": true},
		"active":    {"completed": true, "archived": true},
		"completed": {"archived": true},
	}
	if !allowed[item.Status][status] {
		return errors.New("invalid scenario status transition")
	}
	item.Status = status
	item.UpdatedAt = time.Now().UTC()
	s.scenarios[id] = item
	return nil
}

func (s *Service) Calculate(id string, readings []domain.TrafficReading) ([]Result, error) {
	item, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	byIntersection := map[string][]domain.TrafficReading{}
	for _, reading := range readings {
		byIntersection[reading.IntersectionID] = append(byIntersection[reading.IntersectionID], reading)
	}
	results := make([]Result, 0, len(item.IntersectionIDs))
	for _, intersectionID := range item.IntersectionIDs {
		values := byIntersection[intersectionID]
		if len(values) == 0 {
			continue
		}
		var delay, queue, volume float64
		for _, value := range values {
			volume += float64(value.Volume)
			queue += value.Queue
			delay += delayForSpeed(value.Speed, item.SpeedMultiplier)
		}
		baselineDelay := delay / float64(len(values))
		baselineQueue := queue / float64(len(values))
		scenarioDelay := baselineDelay / item.SpeedMultiplier
		scenarioQueue := baselineQueue * item.FlowMultiplier
		improvement := percentChange(baselineDelay, scenarioDelay)
		results = append(results, Result{ScenarioID: id, IntersectionID: intersectionID, BaselineDelay: baselineDelay, ScenarioDelay: scenarioDelay, BaselineQueue: baselineQueue, ScenarioQueue: scenarioQueue, Throughput: volume * item.FlowMultiplier, Improvement: improvement, CalculatedAt: time.Now().UTC()})
	}
	s.mu.Lock()
	s.results[id] = append([]Result{}, results...)
	s.mu.Unlock()
	return results, nil
}

func delayForSpeed(speed, multiplier float64) float64 {
	if speed <= 0 {
		return 120
	}
	value := 60 * (1 - speed/60)
	if value < 0 {
		value = 0
	}
	return value / multiplier
}

func percentChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		return 0
	}
	return (oldValue - newValue) / oldValue * 100
}

func (s *Service) Results(id string) []Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Result{}, s.results[id]...)
}

func (s *Service) BestResult(id string) (Result, error) {
	items := s.Results(id)
	if len(items) == 0 {
		return Result{}, errors.New("scenario result not found")
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Improvement > items[j].Improvement })
	return items[0], nil
}

func (s *Service) Compare(left, right string) map[string]float64 {
	leftItems := s.Results(left)
	rightItems := s.Results(right)
	leftMap := map[string]float64{}
	for _, item := range leftItems {
		leftMap[item.IntersectionID] = item.ScenarioDelay
	}
	result := map[string]float64{}
	for _, item := range rightItems {
		if value, ok := leftMap[item.IntersectionID]; ok {
			result[item.IntersectionID] = value - item.ScenarioDelay
		}
	}
	return result
}
