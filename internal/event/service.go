package event

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Validate(item domain.TrafficEvent) error {
	if err := shared.Require(item.ID, "event id"); err != nil {
		return err
	}
	if err := shared.Require(item.Title, "event title"); err != nil {
		return err
	}
	if err := shared.OneOf(item.Type, "event type", "congestion", "accident", "construction", "device_fault", "signal_fault"); err != nil {
		return err
	}
	return shared.OneOf(item.Level, "event level", "low", "medium", "high", "critical")
}
func (s *Service) Open(item domain.TrafficEvent) error {
	if item.Status == "" {
		item.Status = "detected"
	}
	if item.Source == "" {
		item.Source = "manual"
	}
	if item.StartedAt.IsZero() {
		item.StartedAt = time.Now().UTC()
	}
	if err := s.Validate(item); err != nil {
		return err
	}
	return s.store.Save("traffic_events", item.ID, item)
}
func (s *Service) Get(id string) (domain.TrafficEvent, error) {
	var item domain.TrafficEvent
	err := s.store.Load("traffic_events", id, &item)
	return item, err
}
func (s *Service) Transition(id, status string) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	allowed := map[string][]string{"detected": {"confirmed", "ignored"}, "confirmed": {"processing", "ignored"}, "processing": {"resolved"}, "resolved": {"closed"}, "closed": {}, "ignored": {}}
	valid := false
	for _, next := range allowed[item.Status] {
		if next == status {
			valid = true
		}
	}
	if !valid {
		return fmt.Errorf("cannot transition event %s from %s to %s", id, item.Status, status)
	}
	item.Status = status
	if status == "resolved" || status == "closed" {
		now := time.Now().UTC()
		item.ResolvedAt = &now
	}
	return s.store.Save("traffic_events", id, item)
}
func (s *Service) List(status string) []domain.TrafficEvent {
	items := []domain.TrafficEvent{}
	_ = s.store.List("traffic_events", func(raw []byte) error {
		var item domain.TrafficEvent
		if json.Unmarshal(raw, &item) == nil && (status == "" || item.Status == status) {
			items = append(items, item)
		}
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].StartedAt.After(items[j].StartedAt) })
	return items
}
func (s *Service) DetectFromIndex(id, intersectionID string, index float64) error {
	if index < 3 {
		return nil
	}
	level := "medium"
	if index >= 4 {
		level = "high"
	}
	if index >= 5 {
		level = "critical"
	}
	return s.Open(domain.TrafficEvent{ID: id, Type: "congestion", Level: level, Title: "拥堵指数超过阈值", Status: "detected", IntersectionID: intersectionID, Description: fmt.Sprintf("congestion index %.2f", index), Source: "threshold"})
}
