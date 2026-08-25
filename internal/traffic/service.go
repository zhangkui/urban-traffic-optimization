package traffic

import (
	"encoding/json"
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"math"
	"sort"
	"sync"
	"time"
)

type Service struct {
	store *store.Store
	mu    sync.RWMutex
	cache map[string]domain.TrafficReading
}

func NewService(s *store.Store) *Service {
	return &Service{store: s, cache: map[string]domain.TrafficReading{}}
}
func (s *Service) Validate(reading domain.TrafficReading) error {
	if err := shared.Require(reading.ID, "reading id"); err != nil {
		return err
	}
	if err := shared.Require(reading.IntersectionID, "intersection id"); err != nil {
		return err
	}
	if err := shared.RangeInt(reading.Volume, "volume", 0, 100000); err != nil {
		return err
	}
	if err := shared.RangeFloat(reading.Speed, "speed", 0, 180); err != nil {
		return err
	}
	if err := shared.RangeFloat(reading.Queue, "queue", 0, 10000); err != nil {
		return err
	}
	if err := shared.RangeFloat(reading.Occupancy, "occupancy", 0, 100); err != nil {
		return err
	}
	if reading.RecordedAt.IsZero() {
		return errors.New("recorded at is required")
	}
	return nil
}
func (s *Service) Record(reading domain.TrafficReading) error {
	if err := s.Validate(reading); err != nil {
		return err
	}
	s.mu.Lock()
	s.cache[reading.ID] = reading
	s.mu.Unlock()
	return s.store.Save("traffic_readings", reading.ID, reading)
}
func (s *Service) RecordBatch(readings []domain.TrafficReading) (int, map[string]string) {
	accepted := 0
	rejected := map[string]string{}
	for _, reading := range readings {
		if err := s.Record(reading); err != nil {
			rejected[reading.ID] = err.Error()
		} else {
			accepted++
		}
	}
	return accepted, rejected
}
func (s *Service) Recent(intersectionID string, since time.Time) []domain.TrafficReading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []domain.TrafficReading{}
	for _, reading := range s.cache {
		if reading.IntersectionID == intersectionID && !reading.RecordedAt.Before(since) {
			result = append(result, reading)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].RecordedAt.Before(result[j].RecordedAt) })
	return result
}
func (s *Service) Aggregate(intersectionID string, start, end time.Time) domain.TrafficAggregate {
	result := domain.TrafficAggregate{IntersectionID: intersectionID, Start: start, End: end}
	for _, reading := range s.Recent(intersectionID, start) {
		if reading.RecordedAt.After(end) {
			continue
		}
		result.Volume += reading.Volume
		result.AverageSpeed += reading.Speed
		result.AverageQueue += reading.Queue
		result.AverageOccupancy += reading.Occupancy
		result.Samples++
	}
	if result.Samples > 0 {
		n := float64(result.Samples)
		result.AverageSpeed /= n
		result.AverageQueue /= n
		result.AverageOccupancy /= n
		result.AverageDelay = math.Max(0, (1-result.AverageSpeed/60)*45)
	}
	return result
}
func (s *Service) LoadCache() error {
	return s.store.List("traffic_readings", func(raw []byte) error {
		var reading domain.TrafficReading
		if err := json.Unmarshal(raw, &reading); err != nil {
			return err
		}
		s.mu.Lock()
		s.cache[reading.ID] = reading
		s.mu.Unlock()
		return nil
	})
}
func CongestionIndex(aggregate domain.TrafficAggregate, freeFlow float64) float64 {
	if aggregate.Samples == 0 || freeFlow <= 0 {
		return 0
	}
	speedFactor := 1 - aggregate.AverageSpeed/freeFlow
	queueFactor := math.Min(aggregate.AverageQueue/100, 1)
	occupancyFactor := aggregate.AverageOccupancy / 100
	index := 1 + speedFactor*2 + queueFactor + occupancyFactor
	return math.Round(math.Max(1, index)*100) / 100
}
