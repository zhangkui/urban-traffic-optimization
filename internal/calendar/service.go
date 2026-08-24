package calendar

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"sort"
	"strings"
	"time"
)

type Window struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Weekdays    []time.Weekday `json:"weekdays"`
	StartMinute int            `json:"startMinute"`
	EndMinute   int            `json:"endMinute"`
	PlanID      string         `json:"planId"`
	Priority    int            `json:"priority"`
	Enabled     bool           `json:"enabled"`
}
type Calendar struct {
	ID             string   `json:"id"`
	IntersectionID string   `json:"intersectionId"`
	Timezone       string   `json:"timezone"`
	Windows        []Window `json:"windows"`
}
type Service struct{ calendars map[string]Calendar }

func NewService() *Service { return &Service{calendars: map[string]Calendar{}} }
func (s *Service) Create(c Calendar) error {
	if c.ID == "" || c.IntersectionID == "" {
		return errors.New("calendar identity is incomplete")
	}
	if c.Timezone == "" {
		c.Timezone = "Asia/Shanghai"
	}
	s.calendars[c.ID] = c
	return nil
}
func (s *Service) AddWindow(calendarID string, w Window) error {
	c, ok := s.calendars[calendarID]
	if !ok {
		return errors.New("calendar not found")
	}
	if w.ID == "" || w.PlanID == "" || w.StartMinute < 0 || w.EndMinute > 1440 || w.StartMinute >= w.EndMinute {
		return errors.New("invalid schedule window")
	}
	for _, existing := range c.Windows {
		if existing.Enabled && w.Enabled && overlap(existing, w) && existing.Priority == w.Priority {
			return errors.New("schedule windows overlap")
		}
	}
	c.Windows = append(c.Windows, w)
	sort.Slice(c.Windows, func(i, j int) bool { return c.Windows[i].Priority > c.Windows[j].Priority })
	s.calendars[calendarID] = c
	return nil
}
func overlap(a, b Window) bool {
	for _, day := range a.Weekdays {
		for _, other := range b.Weekdays {
			if day == other && a.StartMinute < b.EndMinute && b.StartMinute < a.EndMinute {
				return true
			}
		}
	}
	return false
}
func (s *Service) PlanAt(calendarID string, at time.Time) (string, error) {
	c, ok := s.calendars[calendarID]
	if !ok {
		return "", errors.New("calendar not found")
	}
	minute := at.Hour()*60 + at.Minute()
	for _, w := range c.Windows {
		if !w.Enabled || !containsDay(w.Weekdays, at.Weekday()) || minute < w.StartMinute || minute >= w.EndMinute {
			continue
		}
		return w.PlanID, nil
	}
	return "", errors.New("no timing plan for current time")
}
func containsDay(days []time.Weekday, day time.Weekday) bool {
	for _, candidate := range days {
		if candidate == day {
			return true
		}
	}
	return false
}
func (s *Service) ActivePlans(at time.Time) map[string]string {
	result := map[string]string{}
	for id, c := range s.calendars {
		if plan, err := s.PlanAt(id, at); err == nil {
			result[c.IntersectionID] = plan
		}
	}
	return result
}
func (s *Service) Validate(c Calendar) []string {
	issues := []string{}
	if strings.TrimSpace(c.ID) == "" {
		issues = append(issues, "calendar id is required")
	}
	if len(c.Windows) == 0 {
		issues = append(issues, "at least one schedule window is required")
	}
	for i, w := range c.Windows {
		if w.StartMinute >= w.EndMinute {
			issues = append(issues, "window "+w.ID+" has invalid time range")
		}
		for j, next := range c.Windows {
			if i < j && overlap(w, next) && w.Priority == next.Priority {
				issues = append(issues, "windows overlap at same priority")
			}
		}
	}
	return issues
}
func (s *Service) Calendar(id string) (Calendar, error) {
	c, ok := s.calendars[id]
	if !ok {
		return Calendar{}, errors.New("calendar not found")
	}
	return c, nil
}
func (s *Service) List() []Calendar {
	result := []Calendar{}
	for _, c := range s.calendars {
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].IntersectionID < result[j].IntersectionID })
	return result
}
func ScheduleForPlan(plan domain.TimingPlan, w Window) bool { return plan.ID == w.PlanID && w.Enabled }
