package demand

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"math"
	"sort"
	"sync"
	"time"
)

type Pair struct {
	Origin       string    `json:"origin"`
	Destination  string    `json:"destination"`
	Volume       int       `json:"volume"`
	AverageSpeed float64   `json:"averageSpeed"`
	TravelTime   float64   `json:"travelTime"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
type Matrix struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Period      string    `json:"period"`
	Pairs       []Pair    `json:"pairs"`
	TotalVolume int       `json:"totalVolume"`
	CreatedAt   time.Time `json:"createdAt"`
}
type Forecast struct {
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	Current     int     `json:"current"`
	Predicted   int     `json:"predicted"`
	Growth      float64 `json:"growth"`
	Confidence  float64 `json:"confidence"`
}
type Service struct {
	mu       sync.RWMutex
	matrices map[string]Matrix
}

func NewService() *Service { return &Service{matrices: map[string]Matrix{}} }
func (s *Service) ValidatePair(item Pair) error {
	if item.Origin == "" || item.Destination == "" {
		return errors.New("origin and destination are required")
	}
	if item.Origin == item.Destination {
		return errors.New("origin and destination must differ")
	}
	if item.Volume < 0 {
		return errors.New("volume cannot be negative")
	}
	if item.AverageSpeed < 0 || item.AverageSpeed > 180 {
		return errors.New("speed is outside range")
	}
	return nil
}
func (s *Service) Create(matrix Matrix) error {
	if matrix.ID == "" || matrix.Name == "" {
		return errors.New("matrix identity is incomplete")
	}
	if len(matrix.Pairs) == 0 {
		return errors.New("matrix requires at least one OD pair")
	}
	total := 0
	for _, pair := range matrix.Pairs {
		if err := s.ValidatePair(pair); err != nil {
			return err
		}
		total += pair.Volume
	}
	matrix.TotalVolume = total
	if matrix.CreatedAt.IsZero() {
		matrix.CreatedAt = time.Now().UTC()
	}
	s.mu.Lock()
	s.matrices[matrix.ID] = matrix
	s.mu.Unlock()
	return nil
}
func (s *Service) AddPair(id string, pair Pair) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ValidatePair(pair); err != nil {
		return err
	}
	matrix, ok := s.matrices[id]
	if !ok {
		return errors.New("matrix not found")
	}
	matrix.Pairs = append(matrix.Pairs, pair)
	matrix.TotalVolume += pair.Volume
	s.matrices[id] = matrix
	return nil
}
func (s *Service) Get(id string) (Matrix, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	matrix, ok := s.matrices[id]
	if !ok {
		return Matrix{}, errors.New("matrix not found")
	}
	return matrix, nil
}
func (s *Service) List() []Matrix {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Matrix{}
	for _, matrix := range s.matrices {
		result = append(result, matrix)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result
}
func (s *Service) TopOrigins(matrix Matrix, limit int) map[string]int {
	totals := map[string]int{}
	for _, pair := range matrix.Pairs {
		totals[pair.Origin] += pair.Volume
	}
	return top(totals, limit)
}
func (s *Service) TopDestinations(matrix Matrix, limit int) map[string]int {
	totals := map[string]int{}
	for _, pair := range matrix.Pairs {
		totals[pair.Destination] += pair.Volume
	}
	return top(totals, limit)
}
func top(values map[string]int, limit int) map[string]int {
	type entry struct {
		id    string
		value int
	}
	items := []entry{}
	for id, value := range values {
		items = append(items, entry{id, value})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].value > items[j].value })
	if limit <= 0 {
		limit = 10
	}
	result := map[string]int{}
	for i, item := range items {
		if i >= limit {
			break
		}
		result[item.id] = item.value
	}
	return result
}
func (s *Service) Balance(matrix Matrix) float64 {
	origins := s.TopOrigins(matrix, len(matrix.Pairs))
	destinations := s.TopDestinations(matrix, len(matrix.Pairs))
	if len(origins) == 0 || len(destinations) == 0 {
		return 0
	}
	var out, in int
	for _, v := range origins {
		out += v
	}
	for _, v := range destinations {
		in += v
	}
	if out+in == 0 {
		return 0
	}
	return 1 - math.Abs(float64(out-in))/float64(out+in)
}
func (s *Service) Forecast(matrix Matrix, growth float64) []Forecast {
	if growth < -0.9 {
		growth = -.9
	}
	result := make([]Forecast, 0, len(matrix.Pairs))
	for _, pair := range matrix.Pairs {
		predicted := int(math.Round(float64(pair.Volume) * (1 + growth)))
		confidence := 1 - math.Abs(growth)*.5
		if confidence < .2 {
			confidence = .2
		}
		result = append(result, Forecast{Origin: pair.Origin, Destination: pair.Destination, Current: pair.Volume, Predicted: predicted, Growth: growth, Confidence: confidence})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Predicted > result[j].Predicted })
	return result
}
func (s *Service) AggregateReadings(readings []domain.TrafficReading, origin, destination string) Pair {
	var volume int
	var speed, queue float64
	count := 0
	for _, reading := range readings {
		if origin != "" && reading.IntersectionID != origin {
			continue
		}
		volume += reading.Volume
		speed += reading.Speed
		queue += reading.Queue
		count++
	}
	if count > 0 {
		speed /= float64(count)
		queue /= float64(count)
	}
	travel := 0.
	if speed > 0 {
		travel = queue / speed * 3600
	}
	return Pair{Origin: origin, Destination: destination, Volume: volume, AverageSpeed: speed, TravelTime: travel, UpdatedAt: time.Now().UTC()}
}
func (s *Service) RemovePair(id, origin, destination string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	matrix, ok := s.matrices[id]
	if !ok {
		return errors.New("matrix not found")
	}
	filtered := matrix.Pairs[:0]
	removed := 0
	for _, pair := range matrix.Pairs {
		if pair.Origin == origin && pair.Destination == destination {
			removed += pair.Volume
			continue
		}
		filtered = append(filtered, pair)
	}
	if removed == 0 {
		return errors.New("OD pair not found")
	}
	matrix.Pairs = filtered
	matrix.TotalVolume -= removed
	s.matrices[id] = matrix
	return nil
}

func (s *Service) ValidateMatrix(matrix Matrix) []string {
	issues := []string{}
	if matrix.ID == "" {
		issues = append(issues, "matrix id is required")
	}
	if matrix.Name == "" {
		issues = append(issues, "matrix name is required")
	}
	if len(matrix.Pairs) == 0 {
		issues = append(issues, "matrix is empty")
	}
	seen := map[string]bool{}
	for _, pair := range matrix.Pairs {
		key := pair.Origin + "->" + pair.Destination
		if seen[key] {
			issues = append(issues, "duplicate OD pair: "+key)
		}
		seen[key] = true
		if err := s.ValidatePair(pair); err != nil {
			issues = append(issues, key+": "+err.Error())
		}
	}
	return issues
}

func (s *Service) ZoneTotals(matrix Matrix, zones map[string]string) map[string]int {
	result := map[string]int{}
	for _, pair := range matrix.Pairs {
		origin := zones[pair.Origin]
		destination := zones[pair.Destination]
		if origin == "" {
			origin = pair.Origin
		}
		if destination == "" {
			destination = pair.Destination
		}
		result[origin+"->"+destination] += pair.Volume
	}
	return result
}

func (s *Service) AverageTravelTime(matrix Matrix) float64 {
	if len(matrix.Pairs) == 0 {
		return 0
	}
	var weighted float64
	var volume int
	for _, pair := range matrix.Pairs {
		weighted += pair.TravelTime * float64(pair.Volume)
		volume += pair.Volume
	}
	if volume == 0 {
		return 0
	}
	return weighted / float64(volume)
}

func (s *Service) Impedance(pair Pair, valueOfTime float64) float64 {
	if valueOfTime < 0 {
		valueOfTime = 0
	}
	congestion := 1.0
	if pair.AverageSpeed > 0 {
		congestion = 40 / pair.AverageSpeed
	}
	if congestion < 1 {
		congestion = 1
	}
	return pair.TravelTime*congestion + float64(pair.Volume)*valueOfTime/1000
}

func (s *Service) Assign(matrix Matrix, origins map[string]int, destinations map[string]int) []Pair {
	result := []Pair{}
	for origin, supply := range origins {
		remaining := supply
		for destination, demand := range destinations {
			if remaining <= 0 {
				break
			}
			if demand <= 0 {
				continue
			}
			amount := remaining
			if amount > demand {
				amount = demand
			}
			result = append(result, Pair{Origin: origin, Destination: destination, Volume: amount, UpdatedAt: time.Now().UTC()})
			remaining -= amount
			destinations[destination] -= amount
		}
	}
	return result
}

func (s *Service) Growth(current, previous Matrix) float64 {
	if previous.TotalVolume == 0 {
		return 0
	}
	return float64(current.TotalVolume-previous.TotalVolume) / float64(previous.TotalVolume)
}

func (s *Service) Filter(matrix Matrix, origin, destination string, minimum int) Matrix {
	filtered := matrix
	filtered.Pairs = []Pair{}
	filtered.TotalVolume = 0
	for _, pair := range matrix.Pairs {
		if origin != "" && pair.Origin != origin {
			continue
		}
		if destination != "" && pair.Destination != destination {
			continue
		}
		if pair.Volume < minimum {
			continue
		}
		filtered.Pairs = append(filtered.Pairs, pair)
		filtered.TotalVolume += pair.Volume
	}
	return filtered
}
