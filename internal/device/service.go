package device

import (
	"encoding/json"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"sort"
	"time"
)

type Service struct {
	store        *store.Store
	offlineAfter time.Duration
}

func NewService(s *store.Store) *Service { return &Service{store: s, offlineAfter: 5 * time.Minute} }
func (s *Service) SetOfflineAfter(value time.Duration) {
	if value >= time.Minute {
		s.offlineAfter = value
	}
}
func (s *Service) Validate(item domain.Device) error {
	if err := shared.Require(item.ID, "device id"); err != nil {
		return err
	}
	if err := shared.Require(item.Code, "device code"); err != nil {
		return err
	}
	if err := shared.Require(item.Name, "device name"); err != nil {
		return err
	}
	if err := shared.Require(item.IntersectionID, "intersection id"); err != nil {
		return err
	}
	if err := shared.OneOf(item.Type, "device type", "signal", "controller", "detector", "camera"); err != nil {
		return err
	}
	return shared.OneOf(item.Status, "device status", "online", "offline", "warning", "maintenance", "disabled")
}
func (s *Service) Register(item domain.Device) error {
	if item.Status == "" {
		item.Status = "offline"
	}
	return s.save(item)
}
func (s *Service) save(item domain.Device) error {
	if err := s.Validate(item); err != nil {
		return err
	}
	return s.store.Save("devices", item.ID, item)
}
func (s *Service) Get(id string) (domain.Device, error) {
	var item domain.Device
	err := s.store.Load("devices", id, &item)
	return item, err
}
func (s *Service) Heartbeat(id string, at time.Time) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	item.LastHeartbeat = at
	item.Status = "online"
	item.Message = "heartbeat received"
	return s.save(item)
}
func (s *Service) Evaluate(item domain.Device, now time.Time) string {
	if item.Status == "disabled" || item.Status == "maintenance" {
		return item.Status
	}
	if item.LastHeartbeat.IsZero() || now.Sub(item.LastHeartbeat) > s.offlineAfter {
		return "offline"
	}
	return "online"
}
func (s *Service) RefreshStatuses(now time.Time) error {
	var first error
	_ = s.store.List("devices", func(raw []byte) error {
		var item domain.Device
		if err := json.Unmarshal(raw, &item); err != nil {
			return err
		}
		status := s.Evaluate(item, now)
		if item.Status != status {
			item.Status = status
			if err := s.store.Save("devices", item.ID, item); err != nil && first == nil {
				first = err
			}
		}
		return nil
	})
	return first
}
func (s *Service) List(status, deviceType string) []domain.Device {
	items := []domain.Device{}
	_ = s.store.List("devices", func(raw []byte) error {
		var item domain.Device
		if json.Unmarshal(raw, &item) == nil && (status == "" || item.Status == status) && (deviceType == "" || item.Type == deviceType) {
			items = append(items, item)
		}
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}
func (s *Service) OnlineRate() float64 {
	items := s.List("", "")
	if len(items) == 0 {
		return 0
	}
	online := 0
	for _, item := range items {
		if item.Status == "online" {
			online++
		}
	}
	return float64(online) / float64(len(items))
}
func (s *Service) Disable(id, reason string) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	item.Status = "disabled"
	item.Message = reason
	return s.save(item)
}
func (s *Service) Enable(id string) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	item.Status = "offline"
	item.Message = "enabled"
	return s.save(item)
}
