package notification

import (
	"encoding/json"
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"time"
)

type Service struct {
	store   *store.Store
	publish func(string, any)
}

func NewService(s *store.Store, publish func(string, any)) *Service {
	return &Service{store: s, publish: publish}
}
func (s *Service) Create(item domain.Notification) error {
	if item.ID == "" || item.UserID == "" || item.Title == "" {
		return errors.New("notification identity and title are required")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	if err := s.store.Save("notifications", item.ID, item); err != nil {
		return err
	}
	if s.publish != nil {
		s.publish("notification.created", item)
	}
	return nil
}
func (s *Service) MarkRead(id string) error {
	var item domain.Notification
	if err := s.store.Load("notifications", id, &item); err != nil {
		return err
	}
	item.Read = true
	return s.store.Save("notifications", id, item)
}
func (s *Service) Unread(userID string) []domain.Notification {
	items := []domain.Notification{}
	_ = s.store.List("notifications", func(raw []byte) error {
		var item domain.Notification
		if json.Unmarshal(raw, &item) == nil && item.UserID == userID && !item.Read {
			items = append(items, item)
		}
		return nil
	})
	return items
}
