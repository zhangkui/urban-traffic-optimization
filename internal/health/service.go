package health

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type Sample struct {
	At        time.Time `json:"at"`
	LatencyMS int       `json:"latencyMs"`
	ErrorRate float64   `json:"errorRate"`
	Online    bool      `json:"online"`
	Version   string    `json:"version"`
}
type Score struct {
	DeviceID        string    `json:"deviceId"`
	Status          string    `json:"status"`
	Score           float64   `json:"score"`
	Availability    float64   `json:"availability"`
	Reliability     float64   `json:"reliability"`
	LatencyScore    float64   `json:"latencyScore"`
	LastSeen        time.Time `json:"lastSeen"`
	Recommendations []string  `json:"recommendations"`
}
type Service struct {
	mu      sync.RWMutex
	samples map[string][]Sample
	window  time.Duration
}

func NewService() *Service { return &Service{samples: map[string][]Sample{}, window: 24 * time.Hour} }
func (s *Service) Record(deviceID string, sample Sample) error {
	if strings.TrimSpace(deviceID) == "" {
		return errors.New("device id is required")
	}
	if sample.LatencyMS < 0 || sample.LatencyMS > 60000 {
		return errors.New("latency is outside range")
	}
	if sample.ErrorRate < 0 || sample.ErrorRate > 1 {
		return errors.New("error rate is outside range")
	}
	if sample.At.IsZero() {
		sample.At = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples[deviceID] = append(s.samples[deviceID], sample)
	s.trimLocked(deviceID, sample.At)
	return nil
}
func (s *Service) trimLocked(id string, now time.Time) {
	cutoff := now.Add(-s.window)
	items := s.samples[id]
	kept := items[:0]
	for _, item := range items {
		if item.At.After(cutoff) {
			kept = append(kept, item)
		}
	}
	s.samples[id] = kept
}
func (s *Service) SetWindow(window time.Duration) error {
	if window < time.Minute {
		return errors.New("health window is too short")
	}
	s.mu.Lock()
	s.window = window
	s.mu.Unlock()
	return nil
}
func (s *Service) Score(deviceID string, now time.Time) (Score, error) {
	s.mu.RLock()
	items := append([]Sample{}, s.samples[deviceID]...)
	s.mu.RUnlock()
	if len(items) == 0 {
		return Score{}, errors.New("no health samples")
	}
	var online, errorRate, latency float64
	latest := items[0]
	for _, item := range items {
		if item.At.After(latest.At) {
			latest = item
		}
		if item.Online {
			online++
		}
		errorRate += item.ErrorRate
		latency += float64(item.LatencyMS)
	}
	availability := online / float64(len(items))
	errorRate /= float64(len(items))
	latency /= float64(len(items))
	latencyScore := math.Max(0, 1-latency/5000)
	score := (availability*.5 + latencyScore*.25 + (1-errorRate)*.25) * 100
	status := "healthy"
	recommendations := []string{}
	if score < 80 {
		status = "warning"
	}
	if score < 60 {
		status = "critical"
	}
	if availability < .95 {
		recommendations = append(recommendations, "check network connectivity and controller heartbeat")
	}
	if latency > 1000 {
		recommendations = append(recommendations, "inspect upstream sensor and controller latency")
	}
	if errorRate > .05 {
		recommendations = append(recommendations, "review device protocol errors")
	}
	if latest.Version == "" {
		recommendations = append(recommendations, "record device firmware version")
	}
	return Score{DeviceID: deviceID, Status: status, Score: score, Availability: availability, Reliability: 1 - errorRate, LatencyScore: latencyScore, LastSeen: latest.At, Recommendations: recommendations}, nil
}
func (s *Service) Rank(now time.Time) []Score {
	s.mu.RLock()
	ids := make([]string, 0, len(s.samples))
	for id := range s.samples {
		ids = append(ids, id)
	}
	s.mu.RUnlock()
	result := []Score{}
	for _, id := range ids {
		if item, err := s.Score(id, now); err == nil {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Score < result[j].Score })
	return result
}
func (s *Service) Offline(now time.Time, after time.Duration) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []string{}
	for id, items := range s.samples {
		if len(items) == 0 || now.Sub(items[len(items)-1].At) > after {
			result = append(result, id)
		}
	}
	sort.Strings(result)
	return result
}
func (s *Service) Clear(deviceID string) { s.mu.Lock(); delete(s.samples, deviceID); s.mu.Unlock() }
func Healthy(item Score) bool            { return item.Status == "healthy" && item.Score >= 80 }
func DeviceTypeWeight(device domain.Device) float64 {
	switch device.Type {
	case "controller":
		return 1.2
	case "camera":
		return 1.
	case "detector":
		return .9
	default:
		return .8
	}
}
