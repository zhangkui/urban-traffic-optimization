package analysis

import (
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/traffic"
	"math"
	"sort"
	"time"
)

type Service struct{ traffic *traffic.Service }

func NewService(t *traffic.Service) *Service { return &Service{traffic: t} }

type IntersectionMetric struct {
	ID         string  `json:"id"`
	Volume     int     `json:"volume"`
	Speed      float64 `json:"speed"`
	Delay      float64 `json:"delay"`
	Queue      float64 `json:"queue"`
	Congestion float64 `json:"congestion"`
	Rank       int     `json:"rank"`
}
type TrendPoint struct {
	At     time.Time `json:"at"`
	Volume int       `json:"volume"`
	Speed  float64   `json:"speed"`
	Queue  float64   `json:"queue"`
	Delay  float64   `json:"delay"`
}
type Comparison struct {
	IntersectionID string                  `json:"intersectionId"`
	Current        domain.TrafficAggregate `json:"current"`
	Previous       domain.TrafficAggregate `json:"previous"`
	SpeedChange    float64                 `json:"speedChange"`
	DelayChange    float64                 `json:"delayChange"`
	VolumeChange   float64                 `json:"volumeChange"`
}

func (s *Service) Metric(id string, start, end time.Time) IntersectionMetric {
	a := s.traffic.Aggregate(id, start, end)
	congestion := traffic.CongestionIndex(a, 50)
	return IntersectionMetric{ID: id, Volume: a.Volume, Speed: a.AverageSpeed, Delay: a.AverageDelay, Queue: a.AverageQueue, Congestion: congestion}
}
func (s *Service) Trend(id string, start, end time.Time, step time.Duration) []TrendPoint {
	if step <= 0 {
		step = time.Hour
	}
	result := []TrendPoint{}
	for cursor := start; cursor.Before(end); cursor = cursor.Add(step) {
		next := cursor.Add(step)
		if next.After(end) {
			next = end
		}
		a := s.traffic.Aggregate(id, cursor, next)
		result = append(result, TrendPoint{At: cursor, Volume: a.Volume, Speed: a.AverageSpeed, Queue: a.AverageQueue, Delay: a.AverageDelay})
	}
	return result
}
func (s *Service) Compare(id string, currentStart, currentEnd time.Time) Comparison {
	duration := currentEnd.Sub(currentStart)
	previousEnd := currentStart
	previousStart := previousEnd.Add(-duration)
	current := s.traffic.Aggregate(id, currentStart, currentEnd)
	previous := s.traffic.Aggregate(id, previousStart, previousEnd)
	return Comparison{IntersectionID: id, Current: current, Previous: previous, SpeedChange: percent(current.AverageSpeed, previous.AverageSpeed), DelayChange: percent(current.AverageDelay, previous.AverageDelay), VolumeChange: percent(float64(current.Volume), float64(previous.Volume))}
}
func percent(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return (current - previous) / math.Abs(previous) * 100
}
func (s *Service) Rank(ids []string, start, end time.Time) []IntersectionMetric {
	items := make([]IntersectionMetric, 0, len(ids))
	for _, id := range ids {
		items = append(items, s.Metric(id, start, end))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Congestion > items[j].Congestion })
	for i := range items {
		items[i].Rank = i + 1
	}
	return items
}
func (s *Service) DetectAnomaly(points []TrendPoint) []TrendPoint {
	if len(points) < 3 {
		return nil
	}
	var total float64
	for _, p := range points {
		total += float64(p.Volume)
	}
	mean := total / float64(len(points))
	var variance float64
	for _, p := range points {
		d := float64(p.Volume) - mean
		variance += d * d
	}
	deviation := math.Sqrt(variance / float64(len(points)))
	result := []TrendPoint{}
	for _, p := range points {
		if deviation > 0 && math.Abs(float64(p.Volume)-mean) > 2*deviation {
			result = append(result, p)
		}
	}
	return result
}
func (s *Service) Heatmap(ids []string, start, end time.Time) map[string]float64 {
	result := map[string]float64{}
	for _, id := range ids {
		result[id] = s.Metric(id, start, end).Congestion
	}
	return result
}
func (s *Service) DirectionBalance(readings []domain.TrafficReading) float64 {
	if len(readings) == 0 {
		return 0
	}
	totals := map[string]int{}
	for _, r := range readings {
		totals[r.Lane] += r.Volume
	}
	var max, min int
	for _, v := range totals {
		if max == 0 || v > max {
			max = v
		}
		if min == 0 || v < min {
			min = v
		}
	}
	if max == 0 {
		return 1
	}
	return float64(min) / float64(max)
}
