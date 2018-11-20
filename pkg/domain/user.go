package domain

import "github.com/mimir-news/pkg/schema/user"

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
