package maintenance

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"sort"
	"sync"
	"time"
)

type Inspection struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"deviceId"`
	Inspector   string    `json:"inspector"`
	Status      string    `json:"status"`
	Score       int       `json:"score"`
	Findings    []string  `json:"findings"`
	InspectedAt time.Time `json:"inspectedAt"`
	NextDueAt   time.Time `json:"nextDueAt"`
}
type Service struct {
	mu       sync.RWMutex
	items    map[string]Inspection
	interval time.Duration
}

func NewService() *Service {
	return &Service{items: map[string]Inspection{}, interval: 30 * 24 * time.Hour}
}
func (s *Service) Schedule(device domain.Device, inspector string, when time.Time) Inspection {
	if when.IsZero() {
		when = time.Now().UTC()
	}
	item := Inspection{ID: "inspection-" + device.ID + "-" + when.Format("20060102"), DeviceID: device.ID, Inspector: inspector, Status: "scheduled", InspectedAt: when, NextDueAt: when.Add(s.interval)}
	s.mu.Lock()
	s.items[item.ID] = item
	s.mu.Unlock()
	return item
}
func (s *Service) Complete(id string, score int, findings []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return errors.New("inspection not found")
	}
	if score < 0 || score > 100 {
		return errors.New("score must be between 0 and 100")
	}
	item.Score = score
	item.Findings = append([]string{}, findings...)
	item.Status = "completed"
	if score < 60 {
		item.Status = "attention"
	}
	item.NextDueAt = time.Now().UTC().Add(s.interval)
	s.items[id] = item
	return nil
}
func (s *Service) Get(id string) (Inspection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return Inspection{}, errors.New("inspection not found")
	}
	return item, nil
}
func (s *Service) List(deviceID, status string) []Inspection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Inspection{}
	for _, item := range s.items {
		if deviceID != "" && item.DeviceID != deviceID {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].InspectedAt.After(result[j].InspectedAt) })
	return result
}
func (s *Service) Overdue(now time.Time) []Inspection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Inspection{}
	for _, item := range s.items {
		if item.Status != "completed" && item.NextDueAt.Before(now) {
			result = append(result, item)
		}
	}
	return result
}
func (s *Service) SetInterval(interval time.Duration) error {
	if interval < time.Hour {
		return errors.New("inspection interval is too short")
	}
	s.interval = interval
	return nil
}
