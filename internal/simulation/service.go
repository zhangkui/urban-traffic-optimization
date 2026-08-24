package simulation

import (
	"context"
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"math"
	"time"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Validate(task domain.SimulationTask) error {
	if err := shared.Require(task.ID, "simulation id"); err != nil {
		return err
	}
	if err := shared.Require(task.Name, "simulation name"); err != nil {
		return err
	}
	if len(task.IntersectionIDs) == 0 {
		return errors.New("at least one intersection is required")
	}
	if task.EndAt.Before(task.StartAt) {
		return errors.New("simulation end must not precede start")
	}
	if task.FlowMultiplier < 0.1 || task.FlowMultiplier > 5 {
		return errors.New("flow multiplier must be between 0.1 and 5")
	}
	return nil
}
func (s *Service) Create(task domain.SimulationTask) error {
	if task.Status == "" {
		task.Status = "created"
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}
	if err := s.Validate(task); err != nil {
		return err
	}
	return s.store.Save("simulation_tasks", task.ID, task)
}
func (s *Service) Get(id string) (domain.SimulationTask, error) {
	var task domain.SimulationTask
	err := s.store.Load("simulation_tasks", id, &task)
	return task, err
}
func (s *Service) Queue(id string) error {
	task, err := s.Get(id)
	if err != nil {
		return err
	}
	if task.Status != "created" && task.Status != "failed" {
		return errors.New("only created or failed tasks can be queued")
	}
	task.Status = "queued"
	task.Progress = 0
	return s.store.Save("simulation_tasks", id, task)
}
func (s *Service) Run(ctx context.Context, id string, volume, capacity int) (domain.SimulationTask, error) {
	task, err := s.Get(id)
	if err != nil {
		return task, err
	}
	if task.Status != "queued" {
		return task, errors.New("simulation must be queued before running")
	}
	task.Status = "running"
	task.Progress = 10
	_ = s.store.Save("simulation_tasks", id, task)
	for step := 1; step <= 5; step++ {
		select {
		case <-ctx.Done():
			task.Status = "cancelled"
			_ = s.store.Save("simulation_tasks", id, task)
			return task, ctx.Err()
		case <-time.After(5 * time.Millisecond):
			task.Progress = step * 20
			_ = s.store.Save("simulation_tasks", id, task)
		}
	}
	overload := math.Max(0, float64(volume-capacity))
	task.Result = &domain.SimulationResult{TaskID: id, AverageDelay: overload / 10, AverageQueue: overload / 5, Throughput: math.Min(float64(volume), float64(capacity)), Stops: overload / 20, TravelTime: float64(len(task.IntersectionIDs))*3 + overload/50, CreatedAt: time.Now().UTC()}
	task.Status = "completed"
	return task, s.store.Save("simulation_tasks", id, task)
}
func (s *Service) Cancel(id string) error {
	task, err := s.Get(id)
	if err != nil {
		return err
	}
	if task.Status != "queued" && task.Status != "running" {
		return errors.New("only queued or running tasks can be cancelled")
	}
	task.Status = "cancelled"
	return s.store.Save("simulation_tasks", id, task)
}
