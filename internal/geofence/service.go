package geofence

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"time"
)

type Zone struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CenterLat    float64   `json:"centerLat"`
	CenterLon    float64   `json:"centerLon"`
	RadiusMeters float64   `json:"radiusMeters"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
}
type Rule struct {
	ID        string        `json:"id"`
	ZoneID    string        `json:"zoneId"`
	Metric    string        `json:"metric"`
	Threshold float64       `json:"threshold"`
	Duration  time.Duration `json:"duration"`
	Severity  string        `json:"severity"`
	Enabled   bool          `json:"enabled"`
}
type Alert struct {
	RuleID      string    `json:"ruleId"`
	ZoneID      string    `json:"zoneId"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Severity    string    `json:"severity"`
	TriggeredAt time.Time `json:"triggeredAt"`
}
type Service struct {
	zones map[string]Zone
	rules map[string]Rule
}

func NewService() *Service { return &Service{zones: map[string]Zone{}, rules: map[string]Rule{}} }
func (s *Service) AddZone(z Zone) error {
	if z.ID == "" || z.Name == "" || z.RadiusMeters <= 0 {
		return errors.New("invalid geofence zone")
	}
	if z.CenterLat < -90 || z.CenterLat > 90 || z.CenterLon < -180 || z.CenterLon > 180 {
		return errors.New("invalid zone coordinates")
	}
	z.CreatedAt = time.Now().UTC()
	z.Enabled = true
	s.zones[z.ID] = z
	return nil
}
func (s *Service) AddRule(r Rule) error {
	if r.ID == "" || r.ZoneID == "" || r.Metric == "" || r.Threshold < 0 {
		return errors.New("invalid geofence rule")
	}
	if _, ok := s.zones[r.ZoneID]; !ok {
		return errors.New("zone not found")
	}
	if r.Duration <= 0 {
		r.Duration = 5 * time.Minute
	}
	r.Enabled = true
	s.rules[r.ID] = r
	return nil
}
func (s *Service) Contains(zoneID string, lat, lon float64) bool {
	z, ok := s.zones[zoneID]
	if !ok || !z.Enabled {
		return false
	}
	return haversine(z.CenterLat, z.CenterLon, lat, lon) <= z.RadiusMeters
}
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const radius = 6371000.
	p1 := lat1 * math.Pi / 180
	p2 := lat2 * math.Pi / 180
	dp := (lat2 - lat1) * math.Pi / 180
	dl := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	return radius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
func (s *Service) Evaluate(values map[string]float64, now time.Time) []Alert {
	result := []Alert{}
	for _, rule := range s.rules {
		if !rule.Enabled {
			continue
		}
		if value, ok := values[rule.Metric]; ok && value >= rule.Threshold {
			result = append(result, Alert{RuleID: rule.ID, ZoneID: rule.ZoneID, Metric: rule.Metric, Value: value, Severity: rule.Severity, TriggeredAt: now})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Value > result[j].Value })
	return result
}
func (s *Service) Zones() []Zone {
	result := []Zone{}
	for _, z := range s.zones {
		result = append(result, z)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
func (s *Service) Rules(zoneID string) []Rule {
	result := []Rule{}
	for _, r := range s.rules {
		if zoneID == "" || r.ZoneID == zoneID {
			result = append(result, r)
		}
	}
	return result
}
func EventValue(event domain.TrafficEvent) float64 {
	if event.Level == "critical" {
		return 100
	}
	if event.Level == "major" {
		return 75
	}
	return 40
}
