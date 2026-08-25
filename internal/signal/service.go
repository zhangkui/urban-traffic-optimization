package signal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"sort"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) ValidatePhase(phase domain.SignalPhase) error {
	if err := shared.Require(phase.ID, "phase id"); err != nil {
		return err
	}
	if err := shared.Require(phase.Name, "phase name"); err != nil {
		return err
	}
	if err := shared.RangeInt(phase.Green, "green time", 5, 180); err != nil {
		return err
	}
	if err := shared.RangeInt(phase.Yellow, "yellow time", 2, 10); err != nil {
		return err
	}
	if err := shared.RangeInt(phase.AllRed, "all red time", 0, 10); err != nil {
		return err
	}
	if len(phase.Movements) == 0 {
		return errors.New("phase movements are required")
	}
	return nil
}
func (s *Service) Conflicts(phases []domain.SignalPhase) []string {
	conflicts := []string{}
	for i := range phases {
		for j := i + 1; j < len(phases); j++ {
			for _, left := range phases[i].Movements {
				for _, right := range phases[j].Movements {
					if left == right {
						conflicts = append(conflicts, fmt.Sprintf("movement %s appears in phases %s and %s", left, phases[i].ID, phases[j].ID))
					}
				}
			}
		}
	}
	return conflicts
}
func (s *Service) ValidatePlan(plan domain.TimingPlan) error {
	if err := shared.Require(plan.ID, "plan id"); err != nil {
		return err
	}
	if err := shared.Require(plan.Name, "plan name"); err != nil {
		return err
	}
	if err := shared.Require(plan.IntersectionID, "intersection id"); err != nil {
		return err
	}
	if err := shared.RangeInt(plan.Cycle, "cycle", 30, 300); err != nil {
		return err
	}
	if plan.Offset < 0 || plan.Offset >= plan.Cycle {
		return errors.New("offset must be within cycle")
	}
	if len(plan.Phases) == 0 {
		return errors.New("at least one phase is required")
	}
	sum := 0
	for _, phase := range plan.Phases {
		if err := s.ValidatePhase(phase); err != nil {
			return err
		}
		sum += phase.Green + phase.Yellow + phase.AllRed
	}
	if sum > plan.Cycle {
		return fmt.Errorf("phase duration %d exceeds cycle %d", sum, plan.Cycle)
	}
	if conflicts := s.Conflicts(plan.Phases); len(conflicts) > 0 {
		return errors.New(conflicts[0])
	}
	return shared.OneOf(plan.Status, "plan status", "draft", "pending_review", "approved", "active", "disabled", "archived")
}
func (s *Service) Create(plan domain.TimingPlan) error {
	if plan.Status == "" {
		plan.Status = "draft"
	}
	if plan.Version < 1 {
		plan.Version = 1
	}
	if err := s.ValidatePlan(plan); err != nil {
		return err
	}
	return s.store.Save("timing_plans", plan.ID, plan)
}
func (s *Service) Get(id string) (domain.TimingPlan, error) {
	var plan domain.TimingPlan
	err := s.store.Load("timing_plans", id, &plan)
	return plan, err
}
func (s *Service) SubmitReview(id string) error { return s.transition(id, "pending_review") }
func (s *Service) Approve(id string) error      { return s.transition(id, "approved") }
func (s *Service) Publish(id string) error {
	plan, err := s.Get(id)
	if err != nil {
		return err
	}
	if plan.Status != "approved" {
		return errors.New("only approved plans can be published")
	}
	plan.Status = "active"
	return s.store.Save("timing_plans", id, plan)
}
func (s *Service) transition(id, status string) error {
	plan, err := s.Get(id)
	if err != nil {
		return err
	}
	allowed := map[string][]string{"draft": {"pending_review"}, "pending_review": {"approved", "draft"}, "approved": {"active", "draft"}, "active": {"disabled", "archived"}, "disabled": {"draft"}, "archived": {}}
	valid := false
	for _, candidate := range allowed[plan.Status] {
		if candidate == status {
			valid = true
		}
	}
	if !valid {
		return fmt.Errorf("cannot transition %s to %s", plan.Status, status)
	}
	plan.Status = status
	return s.store.Save("timing_plans", id, plan)
}
func (s *Service) List(intersectionID string) []domain.TimingPlan {
	plans := []domain.TimingPlan{}
	_ = s.store.List("timing_plans", func(raw []byte) error {
		var plan domain.TimingPlan
		if json.Unmarshal(raw, &plan) == nil && (intersectionID == "" || plan.IntersectionID == intersectionID) {
			plans = append(plans, plan)
		}
		return nil
	})
	sort.Slice(plans, func(i, j int) bool { return plans[i].Version > plans[j].Version })
	return plans
}
func (s *Service) GreenRatio(plan domain.TimingPlan) float64 {
	if plan.Cycle == 0 {
		return 0
	}
	green := 0
	for _, phase := range plan.Phases {
		green += phase.Green
	}
	return float64(green) / float64(plan.Cycle)
}
