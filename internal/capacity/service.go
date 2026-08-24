package capacity

import (
	"errors"
	"math"
	"sort"
	"time"

	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
)

type ApproachInput struct {
	ID             string  `json:"id"`
	LaneCount      int     `json:"laneCount"`
	SaturationFlow float64 `json:"saturationFlow"`
	Green          int     `json:"green"`
	Cycle          int     `json:"cycle"`
	Arrival        int     `json:"arrival"`
	Speed          float64 `json:"speed"`
	Queue          float64 `json:"queue"`
}
type ApproachResult struct {
	ID       string  `json:"id"`
	Capacity float64 `json:"capacity"`
	Demand   float64 `json:"demand"`
	VCRatio  float64 `json:"vcRatio"`
	Delay    float64 `json:"delay"`
	Level    string  `json:"level"`
	Reserve  float64 `json:"reserve"`
}
type Assessment struct {
	IntersectionID   string           `json:"intersectionId"`
	Approaches       []ApproachResult `json:"approaches"`
	TotalCapacity    float64          `json:"totalCapacity"`
	TotalDemand      float64          `json:"totalDemand"`
	AverageVC        float64          `json:"averageVc"`
	CriticalApproach string           `json:"criticalApproach"`
	AssessedAt       time.Time        `json:"assessedAt"`
}
type Service struct {
	DefaultSaturationFlow float64
	DefaultCycle          int
}

func NewService() *Service { return &Service{DefaultSaturationFlow: 1800, DefaultCycle: 120} }
func (s *Service) Validate(input ApproachInput) error {
	if input.ID == "" {
		return errors.New("approach id is required")
	}
	if input.LaneCount <= 0 || input.LaneCount > 20 {
		return errors.New("lane count is outside range")
	}
	if input.SaturationFlow <= 0 {
		input.SaturationFlow = s.DefaultSaturationFlow
	}
	if input.Green <= 0 || input.Cycle <= input.Green {
		return errors.New("green and cycle values are invalid")
	}
	if input.Arrival < 0 {
		return errors.New("arrival demand cannot be negative")
	}
	if input.Speed < 0 || input.Speed > 180 {
		return errors.New("speed is outside range")
	}
	return nil
}
func (s *Service) Evaluate(input ApproachInput) ApproachResult {
	flow := input.SaturationFlow
	if flow <= 0 {
		flow = s.DefaultSaturationFlow
	}
	cycle := input.Cycle
	if cycle <= 0 {
		cycle = s.DefaultCycle
	}
	capacity := flow * float64(input.LaneCount) * float64(input.Green) / float64(cycle)
	demand := float64(input.Arrival)
	ratio := 0.
	if capacity > 0 {
		ratio = demand / capacity
	}
	delay := s.delay(ratio, input.Cycle, input.Green)
	return ApproachResult{ID: input.ID, Capacity: capacity, Demand: demand, VCRatio: ratio, Delay: delay, Level: level(ratio), Reserve: capacity - demand}
}
func (s *Service) delay(ratio float64, cycle, green int) float64 {
	if ratio <= 0 {
		return 0
	}
	if cycle <= 0 {
		cycle = s.DefaultCycle
	}
	if green <= 0 {
		green = cycle / 2
	}
	uniform := float64(cycle) * math.Pow(1-float64(green)/float64(cycle), 2)
	if ratio < 1 {
		uniform += float64(cycle) * ratio * ratio / (2 * (1 - ratio))
	}
	if ratio >= 1 {
		uniform += float64(cycle) * (ratio - 1) * 2
	}
	return uniform
}
func level(ratio float64) string {
	switch {
	case ratio < .6:
		return "A"
	case ratio < .75:
		return "B"
	case ratio < .9:
		return "C"
	case ratio < 1:
		return "D"
	case ratio < 1.1:
		return "E"
	default:
		return "F"
	}
}
func (s *Service) Assess(intersectionID string, inputs []ApproachInput) Assessment {
	result := Assessment{IntersectionID: intersectionID, Approaches: []ApproachResult{}, AssessedAt: time.Now().UTC()}
	for _, input := range inputs {
		item := s.Evaluate(input)
		result.Approaches = append(result.Approaches, item)
		result.TotalCapacity += item.Capacity
		result.TotalDemand += item.Demand
	}
	if len(result.Approaches) > 0 {
		result.AverageVC = result.TotalDemand / result.TotalCapacity
		sort.Slice(result.Approaches, func(i, j int) bool { return result.Approaches[i].VCRatio > result.Approaches[j].VCRatio })
		result.CriticalApproach = result.Approaches[0].ID
	}
	return result
}
func (s *Service) RecommendGreen(input ApproachInput, target float64) int {
	if target <= 0 || target >= 1 {
		target = .85
	}
	flow := input.SaturationFlow
	if flow <= 0 {
		flow = s.DefaultSaturationFlow
	}
	cycle := input.Cycle
	if cycle <= 0 {
		cycle = s.DefaultCycle
	}
	required := float64(input.Arrival) / (flow * float64(input.LaneCount) * target) * float64(cycle)
	green := int(math.Ceil(required))
	if green < 10 {
		green = 10
	}
	if green > cycle-5 {
		green = cycle - 5
	}
	return green
}
func (s *Service) Sensitivity(input ApproachInput, steps []int) []ApproachResult {
	result := make([]ApproachResult, 0, len(steps))
	for _, green := range steps {
		copyInput := input
		copyInput.Green = green
		result = append(result, s.Evaluate(copyInput))
	}
	return result
}
func (s *Service) Bottleneck(assessment Assessment) []ApproachResult {
	items := append([]ApproachResult{}, assessment.Approaches...)
	sort.Slice(items, func(i, j int) bool { return items[i].VCRatio > items[j].VCRatio })
	if len(items) > 3 {
		items = items[:3]
	}
	return items
}
func (s *Service) FromIntersection(item domain.Intersection, approaches []domain.Approach, plans []domain.TimingPlan) Assessment {
	inputs := make([]ApproachInput, 0, len(approaches))
	cycle := s.DefaultCycle
	green := cycle / 2
	if len(plans) > 0 && plans[0].Cycle > 0 {
		cycle = plans[0].Cycle
	}
	if len(plans) > 0 && len(plans[0].Phases) > 0 {
		green = plans[0].Phases[0].Green
	}
	for _, approach := range approaches {
		inputs = append(inputs, ApproachInput{ID: approach.ID, LaneCount: approach.LaneCount, SaturationFlow: s.DefaultSaturationFlow, Green: green, Cycle: cycle, Speed: item.AverageSpeed})
	}
	return s.Assess(item.ID, inputs)
}
