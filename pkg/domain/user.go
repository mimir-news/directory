package domain

import "github.com/mimir-news/pkg/schema/user"

// FullUser user with credentials.
type FullUser struct {
	User        user.User
	Credentials StoredCredentials
}

// StoredCredentials user credentials in hashed and encrypted from.
type StoredCredentials struct {
	Email    string
	Password string
	Salt     string
}
