package sensor

import (
	"encoding/json"
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"sort"
	"time"
)

type Sensor struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	IntersectionID string    `json:"intersectionId"`
	Lane           string    `json:"lane"`
	Enabled        bool      `json:"enabled"`
	LastValue      float64   `json:"lastValue"`
	LastAt         time.Time `json:"lastAt"`
}
type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Register(item Sensor) error {
	if item.ID == "" || item.Code == "" || item.IntersectionID == "" {
		return errors.New("sensor identity is incomplete")
	}
	item.Enabled = true
	return s.store.Save("sensors", item.ID, item)
}
func (s *Service) Get(id string) (Sensor, error) {
	var item Sensor
	err := s.store.Load("sensors", id, &item)
	return item, err
}
func (s *Service) Accept(id string, value float64, at time.Time) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if !item.Enabled {
		return errors.New("sensor is disabled")
	}
	if value < 0 || value > 100000 {
		return errors.New("sensor value is outside bounds")
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	item.LastValue = value
	item.LastAt = at
	return s.store.Save("sensors", id, item)
}
func (s *Service) Disable(id string) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	item.Enabled = false
	return s.store.Save("sensors", id, item)
}
func (s *Service) List(intersectionID string) []Sensor {
	items := []Sensor{}
	_ = s.store.List("sensors", func(raw []byte) error {
		var item Sensor
		if json.Unmarshal(raw, &item) == nil && (intersectionID == "" || item.IntersectionID == intersectionID) {
			items = append(items, item)
		}
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	return items
}
func (s *Service) Stale(now time.Time, limit time.Duration) []Sensor {
	result := []Sensor{}
	for _, item := range s.List("") {
		if item.LastAt.IsZero() || now.Sub(item.LastAt) > limit {
			result = append(result, item)
		}
	}
	return result
}
func ValidateReading(value float64) error {
	return shared.RangeFloat(value, "sensor reading", 0, 100000)
}
func EncodeBatch(values []Sensor) []byte { raw, _ := json.Marshal(values); return raw }
