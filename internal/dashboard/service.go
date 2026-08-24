package dashboard

import (
	"encoding/json"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"time"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Summary() domain.DashboardSummary {
	summary := domain.DashboardSummary{}
	speeds := []float64{}
	delays := []float64{}
	volume := 0
	_ = s.store.List("intersections", func(raw []byte) error {
		var item domain.Intersection
		if json.Unmarshal(raw, &item) == nil {
			summary.ActiveIntersections++
			summary.CongestionIndex += item.Congestion
			speeds = append(speeds, item.AverageSpeed)
			delays = append(delays, item.AverageDelay)
		}
		return nil
	})
	_ = s.store.List("devices", func(raw []byte) error {
		var item domain.Device
		if json.Unmarshal(raw, &item) == nil {
			summary.DeviceTotal++
			if item.Status == "online" {
				summary.OnlineDevices++
			}
		}
		return nil
	})
	_ = s.store.List("traffic_readings", func(raw []byte) error {
		var item domain.TrafficReading
		if json.Unmarshal(raw, &item) == nil && item.RecordedAt.After(time.Now().UTC().Add(-24*time.Hour)) {
			volume += item.Volume
		}
		return nil
	})
	_ = s.store.List("traffic_events", func(raw []byte) error {
		var item domain.TrafficEvent
		if json.Unmarshal(raw, &item) == nil && item.Status != "closed" && item.Status != "ignored" {
			summary.ActiveEvents++
		}
		return nil
	})
	if summary.ActiveIntersections > 0 {
		summary.CongestionIndex /= float64(summary.ActiveIntersections)
	}
	for _, value := range speeds {
		summary.AverageSpeed += value
	}
	for _, value := range delays {
		summary.AverageDelay += value
	}
	if len(speeds) > 0 {
		summary.AverageSpeed /= float64(len(speeds))
	}
	if len(delays) > 0 {
		summary.AverageDelay /= float64(len(delays))
	}
	summary.TodayVolume = volume
	return summary
}
