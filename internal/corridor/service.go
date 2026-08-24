package corridor

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"time"
)

type Corridor struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	RoadIDs         []string  `json:"roadIds"`
	IntersectionIDs []string  `json:"intersectionIds"`
	Direction       string    `json:"direction"`
	Speed           float64   `json:"speed"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
}
type OffsetPlan struct {
	CorridorID  string         `json:"corridorId"`
	Direction   string         `json:"direction"`
	Offsets     map[string]int `json:"offsets"`
	Cycle       int            `json:"cycle"`
	Progression float64        `json:"progression"`
	GeneratedAt time.Time      `json:"generatedAt"`
}
type Service struct{ corridors map[string]Corridor }

func NewService() *Service { return &Service{corridors: map[string]Corridor{}} }
func (s *Service) Create(c Corridor) error {
	if c.ID == "" || c.Name == "" {
		return errors.New("corridor identity is incomplete")
	}
	if len(c.IntersectionIDs) < 2 {
		return errors.New("corridor requires at least two intersections")
	}
	if c.Speed <= 0 {
		c.Speed = 40
	}
	if c.Status == "" {
		c.Status = "draft"
	}
	c.CreatedAt = time.Now().UTC()
	s.corridors[c.ID] = c
	return nil
}
func (s *Service) Get(id string) (Corridor, error) {
	c, ok := s.corridors[id]
	if !ok {
		return Corridor{}, errors.New("corridor not found")
	}
	return c, nil
}
func (s *Service) List(status string) []Corridor {
	items := []Corridor{}
	for _, c := range s.corridors {
		if status == "" || c.Status == status {
			items = append(items, c)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}
func (s *Service) Activate(id string) error {
	c, err := s.Get(id)
	if err != nil {
		return err
	}
	c.Status = "active"
	s.corridors[id] = c
	return nil
}
func (s *Service) CalculateOffsets(id, direction string, cycle int, distances []float64) (OffsetPlan, error) {
	c, err := s.Get(id)
	if err != nil {
		return OffsetPlan{}, err
	}
	if cycle < 30 || cycle > 300 {
		return OffsetPlan{}, errors.New("cycle must be between 30 and 300 seconds")
	}
	if len(distances) != len(c.IntersectionIDs)-1 {
		return OffsetPlan{}, errors.New("distance count does not match corridor")
	}
	offsets := map[string]int{}
	elapsed := 0.
	for i, intersection := range c.IntersectionIDs {
		if i > 0 {
			elapsed += distances[i-1] / (c.Speed / 3.6)
		}
		offsets[intersection] = int(math.Round(math.Mod(elapsed, float64(cycle))))
	}
	progression := s.Progression(c, offsets, cycle, distances)
	return OffsetPlan{CorridorID: id, Direction: direction, Offsets: offsets, Cycle: cycle, Progression: progression, GeneratedAt: time.Now().UTC()}, nil
}
func (s *Service) Progression(c Corridor, offsets map[string]int, cycle int, distances []float64) float64 {
	if len(c.IntersectionIDs) < 2 {
		return 0
	}
	success := 0.
	for i := 0; i < len(distances); i++ {
		travel := distances[i] / (c.Speed / 3.6)
		delta := math.Abs(float64(offsets[c.IntersectionIDs[i+1]]-offsets[c.IntersectionIDs[i]]) - travel)
		delta = math.Mod(delta, float64(cycle))
		if delta > float64(cycle)/2 {
			delta = float64(cycle) - delta
		}
		success += math.Max(0, 1-delta/(float64(cycle)/2))
	}
	return success / float64(len(distances))
}
func (s *Service) Recommend(c Corridor, plans []OffsetPlan) OffsetPlan {
	if len(plans) == 0 {
		return OffsetPlan{CorridorID: c.ID, Cycle: 120}
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].Progression > plans[j].Progression })
	return plans[0]
}
func Compatible(plan OffsetPlan, readings []domain.TrafficReading) bool {
	if plan.Cycle <= 0 {
		return false
	}
	for _, reading := range readings {
		if reading.Speed <= 0 || reading.Queue > float64(plan.Cycle*2) {
			return false
		}
	}
	return true
}
