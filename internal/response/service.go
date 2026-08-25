package response

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"sort"
	"sync"
	"time"
)

type Action struct {
	ID          string     `json:"id"`
	EventID     string     `json:"eventId"`
	Name        string     `json:"name"`
	Owner       string     `json:"owner"`
	Priority    int        `json:"priority"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	Notes       string     `json:"notes"`
}
type Plan struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	EventTypes []string `json:"eventTypes"`
	Level      string   `json:"level"`
	Actions    []Action `json:"actions"`
	Enabled    bool     `json:"enabled"`
}
type Service struct {
	mu      sync.RWMutex
	plans   map[string]Plan
	actions map[string]Action
}

func NewService() *Service { return &Service{plans: map[string]Plan{}, actions: map[string]Action{}} }
func (s *Service) AddPlan(p Plan) error {
	if p.ID == "" || p.Name == "" || len(p.EventTypes) == 0 {
		return errors.New("response plan identity is incomplete")
	}
	p.Enabled = true
	s.mu.Lock()
	s.plans[p.ID] = p
	s.mu.Unlock()
	return nil
}
func (s *Service) PlanFor(event domain.TrafficEvent) []Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Plan{}
	for _, plan := range s.plans {
		if !plan.Enabled || plan.Level != "" && plan.Level != event.Level {
			continue
		}
		for _, kind := range plan.EventTypes {
			if kind == event.Type {
				result = append(result, plan)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return len(result[i].Actions) > len(result[j].Actions) })
	return result
}
func (s *Service) Instantiate(planID, eventID string) []Action {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.plans[planID]
	if !ok {
		return nil
	}
	result := []Action{}
	for _, action := range plan.Actions {
		action.EventID = eventID
		action.Status = "pending"
		s.actions[action.ID] = action
		result = append(result, action)
	}
	return result
}
func (s *Service) Start(id string) error { return s.transition(id, "running") }
func (s *Service) Complete(id, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	action, ok := s.actions[id]
	if !ok {
		return errors.New("response action not found")
	}
	if action.Status != "running" {
		return errors.New("action is not running")
	}
	now := time.Now().UTC()
	action.Status = "completed"
	action.CompletedAt = &now
	action.Notes = notes
	s.actions[id] = action
	return nil
}
func (s *Service) transition(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	action, ok := s.actions[id]
	if !ok {
		return errors.New("response action not found")
	}
	if status == "running" && action.Status != "pending" {
		return errors.New("action cannot be started")
	}
	now := time.Now().UTC()
	action.Status = status
	action.StartedAt = &now
	s.actions[id] = action
	return nil
}
func (s *Service) Actions(eventID, status string) []Action {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Action{}
	for _, action := range s.actions {
		if eventID != "" && action.EventID != eventID {
			continue
		}
		if status != "" && action.Status != status {
			continue
		}
		result = append(result, action)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Priority > result[j].Priority })
	return result
}
func (s *Service) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	action, ok := s.actions[id]
	if !ok {
		return errors.New("response action not found")
	}
	if action.Status == "completed" {
		return errors.New("completed action cannot be cancelled")
	}
	action.Status = "cancelled"
	s.actions[id] = action
	return nil
}

func (s *Service) ApplyPlan(planID, eventID string) error {
	actions := s.Actions(eventID, "")
	if len(actions) == 0 {
		actions = s.Instantiate(planID, eventID)
	}
	for _, action := range actions {
		if err := s.Start(action.ID); err != nil {
			return err
		}
	}
	return nil
}
