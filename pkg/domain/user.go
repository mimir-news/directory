package domain

import (
	"time"

	"github.com/mimir-news/pkg/id"

	"github.com/mimir-news/pkg/schema/user"
)

// FullUser user with credentials.
type FullUser struct {
	User        user.User
	Credentials StoredCredentials
}

// NewUser creates a new full users.
func NewUser(credentials StoredCredentials, watchlists []user.Watchlist) FullUser {
	return FullUser{
		User:        user.New(credentials.Email, watchlists),
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
	ID        string
	UserID    string
	CreatedAt time.Time
}

// NewSession creates a new session.
func NewSession(userID string) Session {
	return Session{
		ID:        id.New(),
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}
}
