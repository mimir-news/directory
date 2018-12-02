package repository

import (
	"database/sql"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/pkg/dbutil"
	"github.com/mimir-news/pkg/schema/stock"
	"github.com/mimir-news/pkg/schema/user"
)

// Common user related errors.
var (
	ErrNoSuchUser = errors.New("No such user")
)

var (
	emptyUser = domain.FullUser{}
)

// UserRepo interface for getting and storing users in the database.
type UserRepo interface {
	Find(userID string) (domain.FullUser, error)
	FindByEmail(email string) (domain.FullUser, error)
	Save(user domain.FullUser) error
	Delete(userID string) error
	FindWatchlists(userID string) ([]user.Watchlist, error)
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
	FROM app_user WHERE id = $1`

// Find attempts to find a user by ID.
func (ur *pgUserRepo) Find(userID string) (domain.FullUser, error) {
	var u domain.FullUser
	err := ur.db.QueryRow(findUserByIDQuery, userID).Scan(
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
	FROM app_user WHERE email = $1`

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
	ON CONFLICT ON CONSTRAINT app_user_pkey 
	DO UPDATE SET email = $2, password = $3, salt = $4`

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
		salt = NULL,
		locked = TRUE
	WHERE id = $1`

// Save upserts a user in the database.
func (ur *pgUserRepo) Delete(userID string) error {
	res, err := ur.db.Exec(deleteUserQuery, userID)
	if err != nil {
		return err
	}

	return dbutil.AssertRowsAffected(res, 1, ErrNoSuchUser)
}

type watchlistMember struct {
	listID        string
	listName      string
	listCreatedAt time.Time
	stockSymbol   sql.NullString
	stockName     sql.NullString
}

const findUserWatchlistsQuery = `
	SELECT w.id, w.name, w.created_at, s.symbol, s.name
	FROM watchlist w 
	LEFT JOIN watchlist_member m ON m.watchlist_id = w.id
	LEFT JOIN stock s ON s.symbol = m.symbol
	WHERE w.user_id = $1
	ORDER BY w.id, m.created_at`

func (ur *pgUserRepo) FindWatchlists(userID string) ([]user.Watchlist, error) {
	rows, err := ur.db.Query(findUserWatchlistsQuery, userID)
	if err == sql.ErrNoRows {
		return []user.Watchlist{}, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	members, err := extractWatchlistMembers(rows)
	if err != nil {
		return nil, err
	}

	return createWatchlists(members), nil
}

func createWatchlists(members []watchlistMember) []user.Watchlist {
	watchlists := make([]user.Watchlist, 0)
	for _, members := range mapMembersByListID(members) {
		watchlist, err := mapMembersToWatchlist(members)
		if err != nil {
			log.Println(err)
			continue
		}
		watchlists = append(watchlists, watchlist)
	}

	sortWatchlists(watchlists)
	return watchlists
}

func sortWatchlists(watchlists []user.Watchlist) {
	sort.Slice(watchlists, func(i, j int) bool {
		return watchlists[i].CreatedAt.Before(watchlists[j].CreatedAt)
	})
}

func mapMembersByListID(members []watchlistMember) map[string][]watchlistMember {
	listMap := make(map[string][]watchlistMember)
	for _, member := range members {
		existingMembers, ok := listMap[member.listID]
		if ok {
			listMap[member.listID] = append(existingMembers, member)
			continue
		}
		listMap[member.listID] = []watchlistMember{member}
	}
	return listMap
}

func mapMembersToWatchlist(members []watchlistMember) (user.Watchlist, error) {
	if len(members) < 1 {
		return user.Watchlist{}, ErrNoSuchWatchlist
	}
	firstMember := members[0]

	stocks := make([]stock.Stock, 0)
	for _, m := range members {
		if !m.stockName.Valid || !m.stockSymbol.Valid {
			continue
		}
		s := stock.Stock{Symbol: m.stockSymbol.String, Name: m.stockName.String}
		stocks = append(stocks, s)
	}

	watchlist := user.Watchlist{
		ID:        firstMember.listID,
		Name:      firstMember.listName,
		Stocks:    stocks,
		CreatedAt: firstMember.listCreatedAt,
	}
	return watchlist, nil
}

func extractWatchlistMembers(rows *sql.Rows) ([]watchlistMember, error) {
	members := make([]watchlistMember, 0)
	for rows.Next() {
		var m watchlistMember
		err := rows.Scan(&m.listID, &m.listName, &m.listCreatedAt, &m.stockSymbol, &m.stockName)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, nil
}

// MockUserRepo mock implementation of UserRepo.
type MockUserRepo struct {
	FindUser domain.FullUser
	FindErr  error
	FindArg  string

	FindByEmailUser domain.FullUser
	FindByEmailErr  error
	FindByEmailArg  string

	SaveErr error
	SaveArg domain.FullUser

	DeleteErr error
	DeleteArg string

	FindWatchlistsRes []user.Watchlist
	FindWatchlistsErr error
	FindWatchlistsArg string
}

// Find mock implementation of finding a user by id.
func (ur *MockUserRepo) Find(id string) (domain.FullUser, error) {
	ur.FindArg = id
	return ur.FindUser, ur.FindErr
}

// FindByEmail mock implementation of finding a user by email.
func (ur *MockUserRepo) FindByEmail(email string) (domain.FullUser, error) {
	ur.FindByEmailArg = email
	return ur.FindByEmailUser, ur.FindByEmailErr
}

// Save mock implementation of saving a user.
func (ur *MockUserRepo) Save(user domain.FullUser) error {
	ur.SaveArg = user
	return ur.SaveErr
}

// Delete mock implementation of deleting a user.
func (ur *MockUserRepo) Delete(id string) error {
	ur.DeleteArg = id
	return ur.DeleteErr
}

// FindWatchlists mock implementation of finding watchlists by user id.
func (ur *MockUserRepo) FindWatchlists(userID string) ([]user.Watchlist, error) {
	ur.FindWatchlistsArg = userID
	return ur.FindWatchlistsRes, ur.FindWatchlistsErr
}
