package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	MemberID  int       `json:"member_id"`
	Name      string    `json:"name"`
	Tier      string    `json:"tier"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var ErrInvalidUser = errors.New("invalid user")

func NewUser(id string, memberID int, name, tier string, now time.Time) User {
	return User{ID: strings.TrimSpace(id), MemberID: memberID, Name: strings.TrimSpace(name), Tier: strings.ToLower(strings.TrimSpace(tier)), Active: true, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
}

func (u User) Validate() error {
	if u.ID == "" || u.MemberID <= 0 || u.Name == "" {
		return fmt.Errorf("%w: id, member and name are required", ErrInvalidUser)
	}
	if !ValidTier(u.Tier) {
		return fmt.Errorf("%w: invalid tier %s", ErrInvalidUser, u.Tier)
	}
	if u.CreatedAt.IsZero() {
		return fmt.Errorf("%w: creation time is required", ErrInvalidUser)
	}
	return nil
}

func ValidTier(tier string) bool {
	switch tier {
	case "standard", "silver", "gold", "platinum":
		return true
	default:
		return false
	}
}

func (u User) Deactivate(now time.Time) User {
	u.Active = false
	u.UpdatedAt = now.UTC()
	return u
}

func (u User) Activate(now time.Time) User {
	u.Active = true
	u.UpdatedAt = now.UTC()
	return u
}

func (u User) DisplayName() string {
	if u.Name == "" {
		return fmt.Sprintf("会员%d", u.MemberID)
	}
	return u.Name
}

func (u User) EligibleForBenefit(code string) bool {
	if !u.Active || strings.TrimSpace(code) == "" {
		return false
	}
	if u.Tier == "platinum" || u.Tier == "gold" {
		return true
	}
	return strings.HasPrefix(strings.ToUpper(code), "BASIC-")
}
