package access

import (
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"strings"
	"sync"
)

type Policy struct {
	mu            sync.RWMutex
	roles         map[string]map[string]bool
	regions       map[string]map[string]bool
	intersections map[string]map[string]bool
}

func NewPolicy() *Policy {
	return &Policy{roles: map[string]map[string]bool{}, regions: map[string]map[string]bool{}, intersections: map[string]map[string]bool{}}
}
func (p *Policy) GrantRole(role, permission string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.roles[role] == nil {
		p.roles[role] = map[string]bool{}
	}
	p.roles[role][permission] = true
}
func (p *Policy) RevokeRole(role, permission string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.roles[role], permission)
}
func (p *Policy) Authorize(user domain.User, permission string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !user.Enabled {
		return errors.New("user is disabled")
	}
	for _, role := range user.RoleIDs {
		if p.roles[role][permission] || p.roles[role]["*"] {
			return nil
		}
	}
	return errors.New("permission denied")
}
func (p *Policy) GrantRegion(userID, region string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.regions[userID] == nil {
		p.regions[userID] = map[string]bool{}
	}
	p.regions[userID][region] = true
}
func (p *Policy) InRegion(userID, region string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.regions[userID][region]
}
func (p *Policy) GrantIntersection(userID, id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.intersections[userID] == nil {
		p.intersections[userID] = map[string]bool{}
	}
	p.intersections[userID][id] = true
}
func (p *Policy) CanAccessIntersection(userID, id string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.intersections[userID][id]
}
func NormalizePermissions(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
