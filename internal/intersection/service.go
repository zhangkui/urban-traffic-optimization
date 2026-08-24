package intersection

import (
	"encoding/json"
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"math"
	"sort"
)

type Service struct{ store *store.Store }

func NewService(s *store.Store) *Service { return &Service{store: s} }
func (s *Service) Validate(item domain.Intersection) error {
	if err := shared.Require(item.ID, "intersection id"); err != nil {
		return err
	}
	if err := shared.Require(item.Name, "intersection name"); err != nil {
		return err
	}
	if err := shared.Require(item.Code, "intersection code"); err != nil {
		return err
	}
	if err := shared.RangeFloat(item.Latitude, "latitude", -90, 90); err != nil {
		return err
	}
	if err := shared.RangeFloat(item.Longitude, "longitude", -180, 180); err != nil {
		return err
	}
	if err := shared.OneOf(item.Level, "intersection level", "一级", "二级", "三级"); err != nil {
		return err
	}
	return shared.OneOf(item.Status, "intersection status", "active", "inactive", "maintenance")
}
func (s *Service) Create(item domain.Intersection) error {
	if item.Status == "" {
		item.Status = "active"
	}
	return s.Validate(item)
}
func (s *Service) Save(item domain.Intersection) error {
	if err := s.Validate(item); err != nil {
		return err
	}
	return s.store.Save("intersections", item.ID, item)
}
func (s *Service) Get(id string) (domain.Intersection, error) {
	var item domain.Intersection
	err := s.store.Load("intersections", id, &item)
	return item, err
}
func (s *Service) List(request domain.PageRequest) (domain.Page[domain.Intersection], error) {
	request = request.Normalize()
	items := []domain.Intersection{}
	err := s.store.List("intersections", func(raw []byte) error {
		var item domain.Intersection
		if json.Unmarshal(raw, &item) != nil {
			return errors.New("invalid intersection record")
		}
		if request.Status != "" && item.Status != request.Status {
			return nil
		}
		items = append(items, item)
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].Congestion > items[j].Congestion })
	return domain.MakePage(items, len(items), request), err
}
func (s *Service) Distance(a, b domain.Intersection) float64 {
	lat1, lat2 := a.Latitude*math.Pi/180, b.Latitude*math.Pi/180
	deltaLat := (b.Latitude - a.Latitude) * math.Pi / 180
	deltaLon := (b.Longitude - a.Longitude) * math.Pi / 180
	h := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	return 6371000 * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}
func (s *Service) UpdateMetrics(id string, speed, delay, congestion float64) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if err := shared.RangeFloat(speed, "average speed", 0, 180); err != nil {
		return err
	}
	if err := shared.RangeFloat(delay, "average delay", 0, 600); err != nil {
		return err
	}
	if err := shared.RangeFloat(congestion, "congestion", 0, 10); err != nil {
		return err
	}
	item.AverageSpeed = speed
	item.AverageDelay = delay
	item.Congestion = congestion
	return s.Save(item)
}
