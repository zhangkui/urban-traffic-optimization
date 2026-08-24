package compliance

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"sort"
	"strings"
	"time"
)

type Finding struct {
	Code        string `json:"code"`
	Severity    string `json:"severity"`
	Field       string `json:"field"`
	Message     string `json:"message"`
	Remediation string `json:"remediation"`
}
type Review struct {
	PlanID     string    `json:"planId"`
	Passed     bool      `json:"passed"`
	Findings   []Finding `json:"findings"`
	ReviewedAt time.Time `json:"reviewedAt"`
	Reviewer   string    `json:"reviewer"`
}
type Service struct {
	minimumCycle int
	maximumCycle int
}

func NewService() *Service { return &Service{minimumCycle: 30, maximumCycle: 300} }
func (s *Service) Review(plan domain.TimingPlan, reviewer string) Review {
	findings := []Finding{}
	if plan.Name == "" {
		findings = append(findings, Finding{"PLAN_NAME", "error", "name", "timing plan name is required", "provide a descriptive plan name"})
	}
	if plan.IntersectionID == "" {
		findings = append(findings, Finding{"PLAN_INTERSECTION", "error", "intersectionId", "intersection is required", "select a managed intersection"})
	}
	if plan.Cycle < s.minimumCycle || plan.Cycle > s.maximumCycle {
		findings = append(findings, Finding{"PLAN_CYCLE", "error", "cycle", "cycle is outside the permitted range", "use a cycle between 30 and 300 seconds"})
	}
	if plan.Offset < 0 || plan.Offset >= plan.Cycle {
		findings = append(findings, Finding{"PLAN_OFFSET", "error", "offset", "offset must be inside the cycle", "set offset smaller than cycle"})
	}
	if len(plan.Phases) == 0 {
		findings = append(findings, Finding{"PLAN_PHASES", "error", "phases", "at least one phase is required", "configure signal phases"})
	}
	total := 0
	for _, phase := range plan.Phases {
		findings = append(findings, s.checkPhase(phase)...)
		total += phase.Green + phase.Yellow + phase.AllRed
	}
	if len(plan.Phases) > 0 && total != plan.Cycle {
		findings = append(findings, Finding{"PLAN_TOTAL", "error", "phases", "phase durations do not match cycle", "adjust phase durations to cycle"})
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Severity > findings[j].Severity })
	return Review{PlanID: plan.ID, Passed: len(findings) == 0, Findings: findings, ReviewedAt: time.Now().UTC(), Reviewer: reviewer}
}
func (s *Service) checkPhase(phase domain.SignalPhase) []Finding {
	result := []Finding{}
	if strings.TrimSpace(phase.ID) == "" {
		result = append(result, Finding{"PHASE_ID", "error", "phase.id", "phase id is required", "assign a unique phase id"})
	}
	if phase.Green < 5 || phase.Green > 180 {
		result = append(result, Finding{"PHASE_GREEN", "error", "phase.green", "green duration is unsafe", "use a green duration between 5 and 180 seconds"})
	}
	if phase.Yellow < 2 || phase.Yellow > 10 {
		result = append(result, Finding{"PHASE_YELLOW", "warning", "phase.yellow", "yellow duration needs review", "use a yellow duration between 2 and 10 seconds"})
	}
	if phase.AllRed < 0 || phase.AllRed > 10 {
		result = append(result, Finding{"PHASE_ALL_RED", "warning", "phase.allRed", "all-red duration needs review", "use a value between 0 and 10 seconds"})
	}
	if len(phase.Movements) == 0 {
		result = append(result, Finding{"PHASE_MOVEMENTS", "error", "phase.movements", "phase has no movements", "configure permitted movements"})
	}
	return result
}
func (s *Service) ConfigureCycle(minimum, maximum int) error {
	if minimum < 10 || maximum <= minimum {
		return errors.New("invalid cycle policy")
	}
	s.minimumCycle = minimum
	s.maximumCycle = maximum
	return nil
}
func (s *Service) Passed(review Review) bool { return review.Passed }
func (s *Service) Blocking(findings []Finding) []Finding {
	result := []Finding{}
	for _, finding := range findings {
		if finding.Severity == "error" || finding.Severity == "critical" {
			result = append(result, finding)
		}
	}
	return result
}
func (s *Service) CanPublish(plan domain.TimingPlan, review Review) error {
	if plan.Status != "approved" {
		return errors.New("only approved plan can be published")
	}
	if !review.Passed {
		return errors.New("compliance review has blocking findings")
	}
	return nil
}
