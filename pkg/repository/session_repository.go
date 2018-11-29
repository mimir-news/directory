package repository

import (
	"database/sql"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/pkg/dbutil"
)

var (
	emptySession = domain.Session{}
)

// SessionRepo interface for storing user session info.
type SessionRepo interface {
	Save(session domain.Session) error
}

// NewSessionRepo creates a new SesssionRepo using the default implementation.
func NewSessionRepo(db *sql.DB) SessionRepo {
	return &pgSessionRepo{
		db: db,
	}
}

type pgSessionRepo struct {
	db *sql.DB
}

const saveSessionQuery = `INSERT INTO session(id, user_id, created_at) VALUES ($1, $2, $3)`

// Save stores a session in the database.
func (sr *pgSessionRepo) Save(session domain.Session) error {
	res, err := sr.db.Exec(saveSessionQuery, session.ID, session.UserID, session.CreatedAt)
	if err != nil {
		return err
	}

	return dbutil.AssertRowsAffected(res, 1, dbutil.ErrFailedInsert)
}

// MockSessionRepo mock implementation of SessionRepo.
type MockSessionRepo struct {
	SaveErr error
	SaveArg domain.Session
}

// Save mock implementation of Saving a session.
func (sr *MockSessionRepo) Save(session domain.Session) error {
	sr.SaveArg = session
	return sr.SaveErr
}
