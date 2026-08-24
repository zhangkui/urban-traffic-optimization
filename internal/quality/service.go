package quality

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type Rule struct {
	Name     string
	Min      float64
	Max      float64
	Required bool
}
type Issue struct {
	ReadingID string `json:"readingId"`
	Field     string `json:"field"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
}
type Result struct {
	Accepted  []domain.TrafficReading `json:"accepted"`
	Rejected  []domain.TrafficReading `json:"rejected"`
	Issues    []Issue                 `json:"issues"`
	Missing   int                     `json:"missing"`
	Duplicate int                     `json:"duplicate"`
}
type Service struct {
	mu    sync.RWMutex
	rules map[string]Rule
	last  map[string]time.Time
}

func NewService() *Service {
	return &Service{rules: map[string]Rule{"volume": {Name: "volume", Min: 0, Max: 100000, Required: true}, "speed": {Name: "speed", Min: 0, Max: 180, Required: true}, "queue": {Name: "queue", Min: 0, Max: 5000, Required: true}, "occupancy": {Name: "occupancy", Min: 0, Max: 100, Required: true}}, last: map[string]time.Time{}}
}
func (s *Service) SetRule(rule Rule) error {
	if strings.TrimSpace(rule.Name) == "" || rule.Min > rule.Max {
		return errors.New("invalid quality rule")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.Name] = rule
	return nil
}
func (s *Service) Rule(name string) (Rule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rule, ok := s.rules[name]
	return rule, ok
}
func (s *Service) Validate(r domain.TrafficReading) []Issue {
	s.mu.RLock()
	rules := make(map[string]Rule, len(s.rules))
	for k, v := range s.rules {
		rules[k] = v
	}
	s.mu.RUnlock()
	issues := []Issue{}
	if strings.TrimSpace(r.ID) == "" {
		issues = append(issues, Issue{Field: "id", Code: "required", Message: "reading id is required", Severity: "error"})
	}
	if strings.TrimSpace(r.IntersectionID) == "" {
		issues = append(issues, Issue{ReadingID: r.ID, Field: "intersectionId", Code: "required", Message: "intersection is required", Severity: "error"})
	}
	values := map[string]float64{"volume": float64(r.Volume), "speed": r.Speed, "queue": r.Queue, "occupancy": r.Occupancy}
	for name, rule := range rules {
		value := values[name]
		if value < rule.Min || value > rule.Max {
			issues = append(issues, Issue{ReadingID: r.ID, Field: name, Code: "range", Message: name + " is outside the configured range", Severity: "error"})
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			issues = append(issues, Issue{ReadingID: r.ID, Field: name, Code: "number", Message: name + " must be finite", Severity: "error"})
		}
	}
	if r.RecordedAt.IsZero() {
		issues = append(issues, Issue{ReadingID: r.ID, Field: "recordedAt", Code: "required", Message: "recorded time is required", Severity: "error"})
	}
	return issues
}
func (s *Service) ValidateBatch(readings []domain.TrafficReading) Result {
	result := Result{Accepted: []domain.TrafficReading{}, Rejected: []domain.TrafficReading{}, Issues: []Issue{}}
	seen := map[string]bool{}
	for _, r := range readings {
		if seen[r.ID] {
			result.Duplicate++
			result.Issues = append(result.Issues, Issue{ReadingID: r.ID, Field: "id", Code: "duplicate", Message: "duplicate reading id", Severity: "warning"})
			continue
		}
		seen[r.ID] = true
		issues := s.Validate(r)
		if len(issues) > 0 {
			result.Rejected = append(result.Rejected, r)
			result.Issues = append(result.Issues, issues...)
		} else {
			result.Accepted = append(result.Accepted, r)
		}
	}
	return result
}
func (s *Service) MissingWindows(readings []domain.TrafficReading, start, end time.Time, window time.Duration) int {
	if window <= 0 {
		return 0
	}
	present := map[int64]bool{}
	for _, r := range readings {
		bucket := r.RecordedAt.Truncate(window)
		present[bucket.Unix()] = true
	}
	missing := 0
	for cursor := start.Truncate(window); cursor.Before(end); cursor = cursor.Add(window) {
		if !present[cursor.Unix()] {
			missing++
		}
	}
	return missing
}
func (s *Service) QualityScore(result Result) float64 {
	total := len(result.Accepted) + len(result.Rejected) + result.Duplicate
	if total == 0 {
		return 100
	}
	score := float64(len(result.Accepted)) / float64(total) * 100
	score -= float64(result.Missing) * 2
	if score < 0 {
		return 0
	}
	return score
}
func (s *Service) Outliers(readings []domain.TrafficReading) []domain.TrafficReading {
	if len(readings) < 4 {
		return nil
	}
	values := make([]float64, len(readings))
	for i, r := range readings {
		values[i] = r.Speed
	}
	sort.Float64s(values)
	q1 := values[len(values)/4]
	q3 := values[len(values)*3/4]
	spread := q3 - q1
	low := q1 - 1.5*spread
	high := q3 + 1.5*spread
	result := []domain.TrafficReading{}
	for _, r := range readings {
		if r.Speed < low || r.Speed > high {
			result = append(result, r)
		}
	}
	return result
}
func (s *Service) Normalize(r domain.TrafficReading) domain.TrafficReading {
	if r.Speed < 0 {
		r.Speed = 0
	}
	if r.Queue < 0 {
		r.Queue = 0
	}
	if r.Occupancy < 0 {
		r.Occupancy = 0
	}
	if r.Occupancy > 100 {
		r.Occupancy = 100
	}
	if r.RecordedAt.IsZero() {
		r.RecordedAt = time.Now().UTC()
	}
	return r
}
