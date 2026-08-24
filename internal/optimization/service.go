package optimization

import (
	"errors"
	"fmt"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"sort"
	"time"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Create(task domain.OptimizationTask) error {
	if err := shared.Require(task.ID, "optimization id"); err != nil {
		return err
	}
	if err := shared.Require(task.IntersectionID, "intersection id"); err != nil {
		return err
	}
	if err := shared.OneOf(task.Objective, "optimization objective", "minimize_delay", "minimize_queue", "maximize_throughput", "reduce_stops", "balance_direction"); err != nil {
		return err
	}
	if task.Status == "" {
		task.Status = "created"
	}
	task.CreatedAt = time.Now().UTC()
	return s.store.Save("optimization_tasks", task.ID, task)
}
func (s *Service) Generate(taskID string, baseCycle, baseOffset int) []domain.OptimizationCandidate {
	candidates := []domain.OptimizationCandidate{}
	cycles := []int{baseCycle - 10, baseCycle, baseCycle + 10}
	offsets := []int{baseOffset - 5, baseOffset, baseOffset + 5}
	index := 0
	for _, cycle := range cycles {
		for _, offset := range offsets {
			if cycle < 30 || cycle > 300 || offset < 0 || offset >= cycle {
				continue
			}
			index++
			green := 0.35 + float64(index%4)*0.08
			candidates = append(candidates, domain.OptimizationCandidate{ID: fmt.Sprintf("candidate-%s-%02d", taskID, index), TaskID: taskID, Cycle: cycle, Offset: offset, GreenRatio: green, Delay: float64(cycle) / green / 20, Queue: float64(cycle) * (1 - green) / 10, Stops: float64(index%5) + 1, Throughput: 1000 * green})
		}
	}
	return s.Rank(candidates)
}
func (s *Service) Rank(candidates []domain.OptimizationCandidate) []domain.OptimizationCandidate {
	result := append([]domain.OptimizationCandidate(nil), candidates...)
	for i := range result {
		result[i].Score = result[i].Throughput - result[i].Delay*2 - result[i].Queue - result[i].Stops
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Score > result[j].Score })
	return result
}
func (s *Service) Recommend(taskID string, candidates []domain.OptimizationCandidate) (domain.OptimizationCandidate, error) {
	ranked := s.Rank(candidates)
	if len(ranked) == 0 {
		return domain.OptimizationCandidate{}, errors.New("no optimization candidates")
	}
	best := ranked[0]
	best.TaskID = taskID
	if err := s.store.Save("optimization_candidates", best.ID, best); err != nil {
		return best, err
	}
	return best, nil
}
