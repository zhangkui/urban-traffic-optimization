package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/shared"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"strings"
	"time"
)

type Service struct {
	store  *store.Store
	secret string
}

func NewService(s *store.Store, secret string) *Service {
	if secret == "" {
		secret = "development-secret"
	}
	return &Service{store: s, secret: secret}
}
func (s *Service) CreateUser(user domain.User, password string) error {
	if err := shared.Require(user.ID, "user id"); err != nil {
		return err
	}
	if err := shared.Require(user.Username, "username"); err != nil {
		return err
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	user.Enabled = true
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	return s.store.Save("users", user.ID, map[string]any{"user": user, "password": s.HashPassword(password)})
}
func (s *Service) HashPassword(password string) string {
	sum := sha256.Sum256([]byte(s.secret + password))
	return hex.EncodeToString(sum[:])
}
func (s *Service) Authenticate(username, password string) (domain.User, error) {
	var found domain.User
	var matched bool
	err := s.store.List("users", func(raw []byte) error {
		var item struct {
			User     domain.User `json:"user"`
			Password string      `json:"password"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return err
		}
		if item.User.Username == username {
			found = item.User
			matched = item.Password == s.HashPassword(password)
		}
		return nil
	})
	if err != nil {
		return found, err
	}
	if !matched || !found.Enabled {
		return found, errors.New("invalid username or password")
	}
	now := time.Now().UTC()
	found.LastLoginAt = &now
	return found, nil
}
func (s *Service) Token(user domain.User) string {
	return strings.Join([]string{user.ID, user.Username, time.Now().UTC().Format("20060102150405")}, ".")
}
func Authorize(user domain.User, permission string, roles map[string]domain.Role) bool {
	for _, roleID := range user.RoleIDs {
		if role, ok := roles[roleID]; ok {
			for _, candidate := range role.Permissions {
				if candidate == permission || candidate == "*" {
					return true
				}
			}
		}
	}
	return false
}
