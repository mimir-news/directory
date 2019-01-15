package domain

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/mimir-news/pkg/httputil/auth"
	"github.com/mimir-news/pkg/id"
	"github.com/mimir-news/pkg/schema/user"
	"golang.org/x/crypto/sha3"
)

// FullUser user with credentials.
type FullUser struct {
	User        user.User
	Credentials StoredCredentials
}

// NewUser creates a new full users.
func NewUser(credentials StoredCredentials, watchlists []user.Watchlist) FullUser {
	return FullUser{
		User:        user.New(credentials.Email, auth.UserRole, watchlists),
		Credentials: credentials,
	}
}

// StoredCredentials user credentials in hashed and encrypted from.
type StoredCredentials struct {
	Email    string
	Password string
	Salt     string
}

// Session describes a users session.
type Session struct {
	ID           string
	UserID       string
	RefreshToken string
	Active       bool
	CreatedAt    time.Time
}

// NewSession creates a new session.
func NewSession(userID string) Session {
	return Session{
		ID:           id.New(),
		UserID:       userID,
		RefreshToken: generateRefreshToken(),
		Active:       true,
		CreatedAt:    time.Now().UTC(),
	}
}

// generateRefreshToken generates a random refresh token.
func generateRefreshToken() string {
	c1 := sha256.Sum256([]byte(id.New()))
	c2 := sha3.Sum256([]byte(id.New()))

	return fmt.Sprintf("%x-%x", c1, c2)
}
