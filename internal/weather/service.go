package weather

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"strings"
	"time"
)

type Observation struct {
	ID          string    `json:"id"`
	RegionCode  string    `json:"regionCode"`
	Condition   string    `json:"condition"`
	Temperature float64   `json:"temperature"`
	RainMM      float64   `json:"rainMm"`
	Visibility  int       `json:"visibility"`
	WindSpeed   float64   `json:"windSpeed"`
	ObservedAt  time.Time `json:"observedAt"`
}
type Impact struct {
	Condition          string  `json:"condition"`
	FlowMultiplier     float64 `json:"flowMultiplier"`
	SpeedMultiplier    float64 `json:"speedMultiplier"`
	CapacityMultiplier float64 `json:"capacityMultiplier"`
	RecommendedGap     int     `json:"recommendedGap"`
	AlertLevel         string  `json:"alertLevel"`
}
type Service struct{ observations map[string]Observation }

func NewService() *Service { return &Service{observations: map[string]Observation{}} }
func (s *Service) Validate(o Observation) error {
	if o.ID == "" || o.RegionCode == "" {
		return errors.New("weather observation identity is incomplete")
	}
	if o.RainMM < 0 || o.RainMM > 500 {
		return errors.New("rainfall is outside range")
	}
	if o.Visibility < 0 || o.Visibility > 100000 {
		return errors.New("visibility is outside range")
	}
	if o.WindSpeed < 0 || o.WindSpeed > 300 {
		return errors.New("wind speed is outside range")
	}
	return nil
}
func (s *Service) Record(o Observation) error {
	if o.ObservedAt.IsZero() {
		o.ObservedAt = time.Now().UTC()
	}
	if err := s.Validate(o); err != nil {
		return err
	}
	s.observations[o.RegionCode] = o
	return nil
}
func (s *Service) Latest(region string) (Observation, bool) {
	o, ok := s.observations[region]
	return o, ok
}
func (s *Service) Impact(o Observation) Impact {
	condition := strings.ToLower(o.Condition)
	result := Impact{Condition: condition, FlowMultiplier: 1, SpeedMultiplier: 1, CapacityMultiplier: 1, RecommendedGap: 3, AlertLevel: "normal"}
	if o.RainMM > 2 || strings.Contains(condition, "rain") {
		result.FlowMultiplier -= math.Min(.3, o.RainMM/100)
		result.SpeedMultiplier -= .15
		result.CapacityMultiplier -= .1
		result.RecommendedGap = 4
	}
	if o.RainMM > 25 || strings.Contains(condition, "storm") {
		result.FlowMultiplier = .55
		result.SpeedMultiplier = .6
		result.CapacityMultiplier = .5
		result.RecommendedGap = 6
		result.AlertLevel = "major"
	}
	if o.Visibility > 0 && o.Visibility < 500 {
		result.SpeedMultiplier *= .7
		result.CapacityMultiplier *= .75
		result.RecommendedGap += 2
		result.AlertLevel = "major"
	}
	if o.WindSpeed > 60 {
		result.CapacityMultiplier *= .8
		result.AlertLevel = "warning"
	}
	return result
}
func (s *Service) AdjustReading(o domain.TrafficReading, impact Impact) domain.TrafficReading {
	o.Volume = int(float64(o.Volume) * impact.FlowMultiplier)
	o.Speed *= impact.SpeedMultiplier
	o.Queue /= math.Max(.1, impact.CapacityMultiplier)
	return o
}
func (s *Service) Expired(o Observation, now time.Time) bool {
	return now.Sub(o.ObservedAt) > time.Hour
}
