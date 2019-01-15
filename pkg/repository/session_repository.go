package repository

import (
	"database/sql"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/pkg/dbutil"
	"github.com/pkg/errors"
)

// Session errors.
var (
	ErrNoSuchSession = errors.New("no such session")
)

var (
	emptySession = domain.Session{}
)

// SessionRepo interface for storing user session info.
type SessionRepo interface {
	Save(session domain.Session) error
	Find(id string) (domain.Session, error)
	Delete(id string) error
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

const saveSessionQuery = `INSERT INTO session(id, user_id, refresh_token, is_active, created_at) VALUES ($1, $2, $3, $4, $5)`

// Save stores a session in the database.
func (sr *pgSessionRepo) Save(s domain.Session) error {
	res, err := sr.db.Exec(saveSessionQuery, s.ID, s.UserID, s.RefreshToken, s.Active, s.CreatedAt)
	if err != nil {
		return errors.Wrap(err, "pgSessionRepo.Save failed")
	}

	return dbutil.AssertRowsAffected(res, 1, dbutil.ErrFailedInsert)
}

const findSessionQuery = `
	SELECT id, user_id, refresh_token, is_active, created_at 
	FROM session WHERE id = $1 AND is_active = 'TRUE'`

// Find retrieves a session from the database.
func (sr *pgSessionRepo) Find(id string) (domain.Session, error) {
	var s domain.Session
	err := sr.db.QueryRow(findSessionQuery, id).Scan(
		&s.ID, &s.UserID, &s.RefreshToken, &s.Active, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return s, ErrNoSuchSession
	} else if err != nil {
		return s, errors.Wrap(err, "pgSessionRepo.Find failed")
	}

	return s, nil
}

const deleteSessionQuery = `
	UPDATE session SET refresh_token = NULL, is_active = 'FALSE', deleted_at = NOW() 
	WHERE id = $1 AND is_active = 'TRUE'`

// Delete stores a session in the database.
func (sr *pgSessionRepo) Delete(id string) error {
	res, err := sr.db.Exec(deleteSessionQuery, id)
	if err != nil {
		return errors.Wrap(err, "pgSessionRepo.Delete failed")
	}

	return dbutil.AssertRowsAffected(res, 1, ErrNoSuchSession)
}

// MockSessionRepo mock implementation of SessionRepo.
type MockSessionRepo struct {
	SaveErr        error
	SaveArg        domain.Session
	SaveInvocation int

	FindSession    domain.Session
	FindErr        error
	FindArg        string
	FindInvocation int

	DeleteErr        error
	DeleteArg        string
	DeleteInvocation int
}

// Save mock implementation of Saving a session.
func (sr *MockSessionRepo) Save(session domain.Session) error {
	sr.SaveArg = session
	sr.SaveInvocation++
	return sr.SaveErr
}

// Find mock implementation of finding a session.
func (sr *MockSessionRepo) Find(id string) (domain.Session, error) {
	sr.FindArg = id
	sr.FindInvocation++
	return sr.FindSession, sr.FindErr
}

// Delete mock implementation of deleting a session.
func (sr *MockSessionRepo) Delete(id string) error {
	sr.DeleteArg = id
	sr.DeleteInvocation++
	return sr.DeleteErr
}

// UnsetArgs sets all MockSessionRepo fields to their default value.
func (sr *MockSessionRepo) UnsetArgs() {
	sr.SaveArg = domain.Session{}
	sr.SaveInvocation = 0

	sr.FindArg = ""
	sr.FindInvocation = 0

	sr.DeleteArg = ""
	sr.DeleteInvocation = 0
}
