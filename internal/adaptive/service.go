package adaptive

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"time"
)

type Controller struct {
	ID             string    `json:"id"`
	IntersectionID string    `json:"intersectionId"`
	MinGreen       int       `json:"minGreen"`
	MaxGreen       int       `json:"maxGreen"`
	Step           int       `json:"step"`
	Enabled        bool      `json:"enabled"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
type Decision struct {
	PhaseID    string  `json:"phaseId"`
	OldGreen   int     `json:"oldGreen"`
	NewGreen   int     `json:"newGreen"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}
type Service struct{ controllers map[string]Controller }

func NewService() *Service { return &Service{controllers: map[string]Controller{}} }
func (s *Service) Register(c Controller) error {
	if c.ID == "" || c.IntersectionID == "" {
		return errors.New("controller identity is incomplete")
	}
	if c.MinGreen < 5 {
		c.MinGreen = 5
	}
	if c.MaxGreen < c.MinGreen {
		c.MaxGreen = 90
	}
	if c.Step <= 0 {
		c.Step = 5
	}
	c.UpdatedAt = time.Now().UTC()
	s.controllers[c.ID] = c
	return nil
}
func (s *Service) Configure(id string, min, max, step int) error {
	c, ok := s.controllers[id]
	if !ok {
		return errors.New("adaptive controller not found")
	}
	if min < 5 || max < min || step <= 0 {
		return errors.New("invalid adaptive limits")
	}
	c.MinGreen = min
	c.MaxGreen = max
	c.Step = step
	c.UpdatedAt = time.Now().UTC()
	s.controllers[id] = c
	return nil
}
func (s *Service) Get(id string) (Controller, error) {
	c, ok := s.controllers[id]
	if !ok {
		return Controller{}, errors.New("adaptive controller not found")
	}
	return c, nil
}
func (s *Service) Decide(c Controller, phases []domain.SignalPhase, readings []domain.TrafficReading) []Decision {
	if !c.Enabled {
		return nil
	}
	pressure := map[string]float64{}
	for _, r := range readings {
		pressure[r.Lane] += float64(r.Volume) + r.Queue*2
	}
	decisions := make([]Decision, 0, len(phases))
	for _, phase := range phases {
		total := 0.
		for _, movement := range phase.Movements {
			total += pressure[movement]
		}
		ratio := 0.
		if len(readings) > 0 {
			ratio = total / float64(len(readings))
		}
		newGreen := phase.Green
		if ratio > 80 {
			newGreen += c.Step
		} else if ratio < 20 {
			newGreen -= c.Step
		}
		if newGreen < c.MinGreen {
			newGreen = c.MinGreen
		}
		if newGreen > c.MaxGreen {
			newGreen = c.MaxGreen
		}
		reason := "traffic pressure is stable"
		if newGreen > phase.Green {
			reason = "increase green time for high demand"
		}
		if newGreen < phase.Green {
			reason = "release green time from low demand"
		}
		confidence := math.Min(1, math.Abs(float64(newGreen-phase.Green))/float64(c.MaxGreen-c.MinGreen)+.5)
		decisions = append(decisions, Decision{PhaseID: phase.ID, OldGreen: phase.Green, NewGreen: newGreen, Reason: reason, Confidence: confidence})
	}
	sort.Slice(decisions, func(i, j int) bool { return decisions[i].Confidence > decisions[j].Confidence })
	return decisions
}
func (s *Service) Apply(plan domain.TimingPlan, decisions []Decision) domain.TimingPlan {
	byID := map[string]Decision{}
	for _, d := range decisions {
		byID[d.PhaseID] = d
	}
	for i := range plan.Phases {
		if d, ok := byID[plan.Phases[i].ID]; ok {
			plan.Phases[i].Green = d.NewGreen
		}
	}
	plan.Version++
	plan.Status = "draft"
	return plan
}
