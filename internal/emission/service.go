package emission

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"strings"
	"time"
)

type Factor struct {
	VehicleType    string  `json:"vehicleType"`
	Pollutant      string  `json:"pollutant"`
	GramPerVehicle float64 `json:"gramPerVehicle"`
	SpeedEffect    float64 `json:"speedEffect"`
}
type Reading struct {
	At             time.Time      `json:"at"`
	IntersectionID string         `json:"intersectionId"`
	Volume         int            `json:"volume"`
	AverageSpeed   float64        `json:"averageSpeed"`
	VehicleTypes   map[string]int `json:"vehicleTypes"`
}
type Estimate struct {
	IntersectionID string    `json:"intersectionId"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
	Pollutant      string    `json:"pollutant"`
	Grams          float64   `json:"grams"`
	Vehicles       int       `json:"vehicles"`
	Intensity      float64   `json:"intensity"`
	GeneratedAt    time.Time `json:"generatedAt"`
}
type Service struct{ factors map[string]Factor }

func NewService() *Service {
	return &Service{factors: map[string]Factor{"car_co": {VehicleType: "car", Pollutant: "co", GramPerVehicle: 12, SpeedEffect: .4}, "car_nox": {VehicleType: "car", Pollutant: "nox", GramPerVehicle: 1.2, SpeedEffect: .2}, "bus_co": {VehicleType: "bus", Pollutant: "co", GramPerVehicle: 45, SpeedEffect: .6}, "truck_nox": {VehicleType: "truck", Pollutant: "nox", GramPerVehicle: 8, SpeedEffect: .5}}}
}
func (s *Service) SetFactor(key string, factor Factor) error {
	if strings.TrimSpace(key) == "" || factor.VehicleType == "" || factor.Pollutant == "" || factor.GramPerVehicle < 0 {
		return errors.New("invalid emission factor")
	}
	s.factors[key] = factor
	return nil
}
func (s *Service) Factor(key string) (Factor, bool) { factor, ok := s.factors[key]; return factor, ok }
func (s *Service) Estimate(reading Reading, pollutant string) Estimate {
	result := Estimate{IntersectionID: reading.IntersectionID, Start: reading.At, End: reading.At, Pollutant: pollutant, GeneratedAt: time.Now().UTC()}
	for vehicle, count := range reading.VehicleTypes {
		for _, factor := range s.factors {
			if factor.VehicleType != vehicle || factor.Pollutant != pollutant {
				continue
			}
			speedEffect := 1.
			if reading.AverageSpeed > 0 {
				speedEffect = 1 + factor.SpeedEffect*math.Abs(40-reading.AverageSpeed)/40
			}
			result.Grams += float64(count) * factor.GramPerVehicle * speedEffect
			result.Vehicles += count
		}
	}
	if result.Vehicles > 0 {
		result.Intensity = result.Grams / float64(result.Vehicles)
	}
	return result
}
func (s *Service) Aggregate(readings []Reading, pollutant string) []Estimate {
	groups := map[string][]Reading{}
	for _, reading := range readings {
		groups[reading.IntersectionID] = append(groups[reading.IntersectionID], reading)
	}
	result := []Estimate{}
	for id, items := range groups {
		var total Estimate
		total.IntersectionID = id
		total.Pollutant = pollutant
		for _, item := range items {
			estimate := s.Estimate(item, pollutant)
			total.Grams += estimate.Grams
			total.Vehicles += estimate.Vehicles
			if total.Start.IsZero() || item.At.Before(total.Start) {
				total.Start = item.At
			}
			if item.At.After(total.End) {
				total.End = item.At
			}
		}
		if total.Vehicles > 0 {
			total.Intensity = total.Grams / float64(total.Vehicles)
		}
		total.GeneratedAt = time.Now().UTC()
		result = append(result, total)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Grams > result[j].Grams })
	return result
}
func (s *Service) Hotspots(estimates []Estimate, limit int) []Estimate {
	if limit <= 0 {
		limit = 10
	}
	items := append([]Estimate{}, estimates...)
	sort.Slice(items, func(i, j int) bool { return items[i].Intensity > items[j].Intensity })
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}
func (s *Service) Compare(before, after Estimate) float64 {
	if before.Grams == 0 {
		return 0
	}
	return (after.Grams - before.Grams) / before.Grams * 100
}
func (s *Service) TargetReduction(current, target float64) float64 {
	if current <= 0 || target < 0 || target >= current {
		return 0
	}
	return (current - target) / current * 100
}
func (s *Service) ValidateReading(reading Reading) error {
	if reading.IntersectionID == "" {
		return errors.New("intersection id is required")
	}
	if reading.Volume < 0 {
		return errors.New("volume cannot be negative")
	}
	if reading.AverageSpeed < 0 || reading.AverageSpeed > 180 {
		return errors.New("speed is outside range")
	}
	return nil
}
func FromTraffic(item domain.TrafficReading) Reading {
	return Reading{At: item.RecordedAt, IntersectionID: item.IntersectionID, Volume: item.Volume, AverageSpeed: item.Speed, VehicleTypes: item.VehicleTypes}
}

func (s *Service) EstimateStrict(reading Reading, pollutant string) (Estimate, error) {
	return s.Estimate(reading, pollutant), nil
}
