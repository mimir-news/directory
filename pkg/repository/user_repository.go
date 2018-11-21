package repository

import (
	"database/sql"
	"errors"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/pkg/dbutil"
)

// Common errors.
var (
	ErrNoSuchUser = errors.New("No such user")
)

var (
	emptyUser = domain.FullUser{}
)

// UserRepo interface for getting and storing users in the database.
type UserRepo interface {
	Find(id string) (domain.FullUser, error)
	FindByEmail(email string) (domain.FullUser, error)
	Save(user domain.FullUser) error
	Delete(id string) error
}

// NewUserRepo creates a new UserRepo using the default implementation.
func NewUserRepo(db *sql.DB) UserRepo {
	return &pgUserRepo{
		db: db,
	}
}

type pgUserRepo struct {
	db *sql.DB
}

const findUserByIDQuery = `SELECT 
	id, email, password, salt, created_at
	FROM app_user id = $1`

// Find attempts to find a user by ID.
func (ur *pgUserRepo) Find(id string) (domain.FullUser, error) {
	var u domain.FullUser
	err := ur.db.QueryRow(findUserByIDQuery, id).Scan(
		&u.User.ID, &u.User.Email, &u.Credentials.Password,
		&u.Credentials.Salt, &u.User.CreatedAt)

	if err == sql.ErrNoRows {
		return emptyUser, ErrNoSuchUser
	} else if err != nil {
		return emptyUser, err
	}
	return u, nil
}

const findUserByEmailQuery = `SELECT 
	id, email, password, salt, created_at
	FROM app_user email = $1`

// FindByEmail attempts to find a user by email.
func (ur *pgUserRepo) FindByEmail(email string) (domain.FullUser, error) {
	var user domain.FullUser
	err := ur.db.QueryRow(findUserByEmailQuery, email).Scan(
		&user.User.ID, &user.User.Email, &user.Credentials.Password,
		&user.Credentials.Salt, &user.User.CreatedAt)

	if err == sql.ErrNoRows {
		return emptyUser, ErrNoSuchUser
	} else if err != nil {
		return emptyUser, err
	}
	return user, nil
}

const saveUserQuery = `
	INSERT INTO 
	app_user(id, email, password, salt, created_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT UPDATE
		email = $2, 
		password = $3,
		salt = $4
		`

// Save upserts a user in the database.
func (ur *pgUserRepo) Save(user domain.FullUser) error {
	u := user.User
	c := user.Credentials
	res, err := ur.db.Exec(saveUserQuery, u.ID, u.Email, c.Password, c.Salt, u.CreatedAt)
	if err != nil {
		return err
	}

	return dbutil.AssertRowsAffected(res, 1, dbutil.ErrFailedInsert)
}

const deleteUserQuery = `
	UPDATE app_user SET
		email = NULL, 
		password = NULL,
		salt = NULL
		locked = TRUE
	WHERE id = $1`

// Save upserts a user in the database.
func (ur *pgUserRepo) Delete(id string) error {
	res, err := ur.db.Exec(deleteUserQuery, id)
	if err != nil {
		return err
	}

	return dbutil.AssertRowsAffected(res, 1, ErrNoSuchUser)
}
