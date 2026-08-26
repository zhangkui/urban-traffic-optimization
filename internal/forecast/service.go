package forecast

import (
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"time"
)

type Model struct {
	Window   time.Duration
	Alpha    float64
	Beta     float64
	Capacity int
}
type Point struct {
	At    time.Time `json:"at"`
	Value float64   `json:"value"`
	Lower float64   `json:"lower"`
	Upper float64   `json:"upper"`
}
type Forecast struct {
	IntersectionID string  `json:"intersectionId"`
	Metric         string  `json:"metric"`
	Points         []Point `json:"points"`
	Capacity       int     `json:"capacity"`
	Saturation     float64 `json:"saturation"`
}
type Service struct{ model Model }

func NewService(model Model) *Service {
	if model.Window <= 0 {
		model.Window = 15 * time.Minute
	}
	if model.Alpha <= 0 || model.Alpha > 1 {
		model.Alpha = .35
	}
	if model.Beta <= 0 || model.Beta > 1 {
		model.Beta = .2
	}
	if model.Capacity <= 0 {
		model.Capacity = 1800
	}
	return &Service{model: model}
}
func (s *Service) Model() Model { return s.model }
func (s *Service) MovingAverage(readings []domain.TrafficReading) float64 {
	if len(readings) == 0 {
		return 0
	}
	values := make([]float64, 0, len(readings))
	for _, r := range readings {
		values = append(values, float64(r.Volume))
	}
	sort.Float64s(values)
	if len(values) > 2 {
		values = values[1 : len(values)-1]
	}
	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}
func (s *Service) Exponential(readings []domain.TrafficReading) float64 {
	if len(readings) == 0 {
		return 0
	}
	estimate := float64(readings[0].Volume)
	for _, r := range readings[1:] {
		estimate = s.model.Alpha*float64(r.Volume) + (1-s.model.Alpha)*estimate
	}
	return estimate
}
func (s *Service) Linear(readings []domain.TrafficReading, horizon int) []Point {
	if horizon <= 0 || len(readings) == 0 {
		return nil
	}
	var sx, sy, sxx, sxy float64
	for i, r := range readings {
		x := float64(i)
		y := float64(r.Volume)
		sx += x
		sy += y
		sxx += x * x
		sxy += x * y
	}
	n := float64(len(readings))
	denom := n*sxx - sx*sx
	slope := 0.
	if denom != 0 {
		slope = (n*sxy - sx*sy) / denom
	}
	intercept := (sy - slope*sx) / n
	result := make([]Point, 0, horizon)
	base := readings[len(readings)-1].RecordedAt
	for i := 1; i <= horizon; i++ {
		value := intercept + slope*float64(len(readings)-1+i)
		if value < 0 {
			value = 0
		}
		spread := math.Sqrt(value+1) * 1.96
		result = append(result, Point{At: base.Add(time.Duration(i) * s.model.Window), Value: value, Lower: math.Max(0, value-spread), Upper: value + spread})
	}
	return result
}
func (s *Service) Saturation(volume float64) float64 { return volume / float64(s.model.Capacity) }
func (s *Service) NeedIntervention(f Forecast) bool {
	if len(f.Points) == 0 {
		return false
	}
	return f.Points[len(f.Points)-1].Upper >= float64(f.Capacity)*.9
}
func (s *Service) Build(intersectionID string, readings []domain.TrafficReading, horizon int) Forecast {
	points := s.Linear(readings, horizon)
	baseline := s.Exponential(readings)
	if len(points) > 0 {
		points[0].Value = (points[0].Value + baseline) / 2
	}
	var volume float64
	for _, r := range readings {
		volume += float64(r.Volume)
	}
	return Forecast{IntersectionID: intersectionID, Metric: "volume", Points: points, Capacity: s.model.Capacity, Saturation: s.Saturation(volume)}
}
