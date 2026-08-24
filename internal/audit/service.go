package audit

import (
	"encoding/json"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"sort"
	"time"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Record(item domain.AuditLog) error {
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	if item.ID == "" {
		item.ID = item.CreatedAt.Format("20060102150405.000000000")
	}
	return s.store.Save("audit_logs", item.ID, item)
}
func (s *Service) Change(actor, action, resource, id string, before, after any) error {
	detail, _ := json.Marshal(map[string]any{"before": before, "after": after})
	return s.Record(domain.AuditLog{ActorID: actor, Action: action, Resource: resource, ResourceID: id, Detail: string(detail)})
}
func (s *Service) List(actor, resource string) []domain.AuditLog {
	items := []domain.AuditLog{}
	_ = s.store.List("audit_logs", func(raw []byte) error {
		var item domain.AuditLog
		if json.Unmarshal(raw, &item) == nil && (actor == "" || item.ActorID == actor) && (resource == "" || item.Resource == resource) {
			items = append(items, item)
		}
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items
}
