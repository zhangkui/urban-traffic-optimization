package jobs

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Status string

const (
	Queued    Status = "queued"
	Running   Status = "running"
	Completed Status = "completed"
	Failed    Status = "failed"
	Cancelled Status = "cancelled"
)

type Job struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     Status     `json:"status"`
	Progress   int        `json:"progress"`
	Error      string     `json:"error,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}
type Handler func(context.Context, func(int)) error
type Queue struct {
	mu       sync.RWMutex
	jobs     map[string]Job
	handlers map[string]Handler
	cancel   map[string]context.CancelFunc
}

func NewQueue() *Queue {
	return &Queue{jobs: map[string]Job{}, handlers: map[string]Handler{}, cancel: map[string]context.CancelFunc{}}
}
func (q *Queue) Register(name string, handler Handler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[name] = handler
}
func (q *Queue) Enqueue(id, name string) error {
	q.mu.Lock()
	if _, ok := q.jobs[id]; ok {
		q.mu.Unlock()
		return errors.New("job already exists")
	}
	if q.handlers[name] == nil {
		q.mu.Unlock()
		return errors.New("job handler is not registered")
	}
	q.jobs[id] = Job{ID: id, Name: name, Status: Queued, CreatedAt: time.Now().UTC()}
	q.mu.Unlock()
	go q.execute(id, name)
	return nil
}
func (q *Queue) execute(id, name string) {
	ctx, cancel := context.WithCancel(context.Background())
	q.mu.Lock()
	q.cancel[id] = cancel
	job := q.jobs[id]
	now := time.Now().UTC()
	job.Status = Running
	job.StartedAt = &now
	q.jobs[id] = job
	handler := q.handlers[name]
	q.mu.Unlock()
	err := handler(ctx, func(progress int) {
		q.mu.Lock()
		item := q.jobs[id]
		item.Progress = progress
		q.jobs[id] = item
		q.mu.Unlock()
	})
	q.mu.Lock()
	defer q.mu.Unlock()
	job = q.jobs[id]
	if errors.Is(err, context.Canceled) {
		job.Status = Cancelled
	} else if err != nil {
		job.Status = Failed
		job.Error = err.Error()
	} else {
		job.Status = Completed
		job.Progress = 100
	}
	finished := time.Now().UTC()
	job.FinishedAt = &finished
	q.jobs[id] = job
	delete(q.cancel, id)
}
func (q *Queue) Cancel(id string) error {
	q.mu.RLock()
	cancel := q.cancel[id]
	_, exists := q.jobs[id]
	q.mu.RUnlock()
	if !exists {
		return errors.New("job not found")
	}
	if cancel == nil {
		return errors.New("job is not running")
	}
	cancel()
	return nil
}
func (q *Queue) Get(id string) (Job, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	job, ok := q.jobs[id]
	return job, ok
}
func (q *Queue) List(status Status) []Job {
	q.mu.RLock()
	defer q.mu.RUnlock()
	items := []Job{}
	for _, job := range q.jobs {
		if status == "" || job.Status == status {
			items = append(items, job)
		}
	}
	return items
}
