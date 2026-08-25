package shared

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"net/http"
	"strings"
	"time"
)

type IDGenerator struct {
	Prefix string
	Next   func() uint64
}

func NewID(prefix string, next func() uint64) IDGenerator {
	return IDGenerator{Prefix: prefix, Next: next}
}
func (g IDGenerator) Generate() string { return fmt.Sprintf("%s-%06d", g.Prefix, g.Next()) }
func Require(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
func OneOf(value, name string, allowed ...string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %s", name, strings.Join(allowed, ","))
}
func RangeInt(value int, name string, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return nil
}
func RangeFloat(value float64, name string, min, max float64) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %.2f and %.2f", name, min, max)
	}
	return nil
}
func DecodeJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func SaveEvent(s *store.Store, subject, action string) error { return s.Event(subject, action) }
func ContextDeadline(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}
	return time.Until(deadline)
}
func UniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
