package network

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"sort"
	"sync"
)

type Graph struct {
	mu    sync.RWMutex
	roads map[string]domain.Road
	edges map[string]map[string]float64
}

func NewGraph() *Graph {
	return &Graph{roads: map[string]domain.Road{}, edges: map[string]map[string]float64{}}
}
func (g *Graph) AddRoad(item domain.Road) error {
	if item.ID == "" || item.Name == "" {
		return errors.New("road identity is incomplete")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.roads[item.ID] = item
	if g.edges[item.ID] == nil {
		g.edges[item.ID] = map[string]float64{}
	}
	for _, next := range item.Downstream {
		g.edges[item.ID][next] = item.LengthMeters
	}
	return nil
}
func (g *Graph) Connect(from, to string, distance float64) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.roads[from].ID == "" || g.roads[to].ID == "" {
		return errors.New("road does not exist")
	}
	if distance <= 0 {
		distance = g.roads[from].LengthMeters
	}
	if g.edges[from] == nil {
		g.edges[from] = map[string]float64{}
	}
	g.edges[from][to] = distance
	return nil
}
func (g *Graph) Neighbors(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := []string{}
	for next := range g.edges[id] {
		result = append(result, next)
	}
	sort.Strings(result)
	return result
}
func (g *Graph) ShortestPath(from, to string) ([]string, float64, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.roads[from].ID == "" || g.roads[to].ID == "" {
		return nil, 0, errors.New("road does not exist")
	}
	distance := map[string]float64{}
	previous := map[string]string{}
	visited := map[string]bool{}
	for id := range g.roads {
		distance[id] = 1e18
	}
	distance[from] = 0
	for range g.roads {
		current := ""
		best := 1e18
		for id, value := range distance {
			if !visited[id] && value < best {
				current = id
				best = value
			}
		}
		if current == "" {
			break
		}
		visited[current] = true
		if current == to {
			break
		}
		for next, weight := range g.edges[current] {
			candidate := best + weight
			if candidate < distance[next] {
				distance[next] = candidate
				previous[next] = current
			}
		}
	}
	if distance[to] >= 1e18 {
		return nil, 0, errors.New("no route")
	}
	path := []string{to}
	for path[len(path)-1] != from {
		path = append(path, previous[path[len(path)-1]])
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, distance[to], nil
}
func (g *Graph) Snapshot() map[string][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := map[string][]string{}
	for id, edges := range g.edges {
		for next := range edges {
			result[id] = append(result[id], next)
		}
		sort.Strings(result[id])
	}
	return result
}
