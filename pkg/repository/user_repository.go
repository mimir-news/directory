package repository

import (
	"github.com/mimir-news/directory/pkg/domain"
)

// UserRepo interface for getting and storing users in the database.
type UserRepo interface {
	Find(id string) (domain.FullUser, error)
	FindByEmail(email string) (domain.FullUser, error)
	Save(user domain.FullUser) error
}
