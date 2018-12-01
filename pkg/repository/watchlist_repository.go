package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/mimir-news/pkg/dbutil"
	"github.com/mimir-news/pkg/schema/stock"
	"github.com/mimir-news/pkg/schema/user"
)

// Common watchlist related errors.
var (
	ErrNoSuchWatchlist      = errors.New("No such watchlist")
	ErrWatchlistExist       = errors.New("Watchlist already exists")
	ErrNoSuchWatchlistStock = errors.New("No such stock in watchlist")
)

var (
	emptyWatchlist = user.Watchlist{}
)

// WatchlistRepo interface for getting and storing watchlists in a database.
type WatchlistRepo interface {
	Get(userID, watchlistID string) (user.Watchlist, error)
	Save(userID string, watchlist user.Watchlist) error
	AddStock(userID, symbol, watchlistID string) error
	DeleteStock(userID, symbol, watchlistID string) error
	Delete(userID, watchlistID string) error
}

// NewWatchlistRepo creates a new watchlist using the default implementation.
func NewWatchlistRepo(db *sql.DB) WatchlistRepo {
	return &pgWatchlistRepo{
		db: db,
	}
}

type pgWatchlistRepo struct {
	db *sql.DB
}

const getWatchlistQuery = `
	SELECT w.id, w.name, w.created_at, s.symbol, s.name 
	FROM watchlist w 
	INNER JOIN watchlist_member m ON m.watchlist_id = w.id
	INNER JOIN stock s ON s.symbol = m.symbol
	WHERE w.user_id = $1 
	AND w.id = $2 
	ORDER BY m.created_at`

// Get gets a watchlist of stocks.
func (wr *pgWatchlistRepo) Get(userID, watchlistID string) (user.Watchlist, error) {
	rows, err := wr.db.Query(getWatchlistQuery, userID, watchlistID)
	if err == sql.ErrNoRows {
		return emptyWatchlist, ErrNoSuchWatchlist
	} else if err != nil {
		return emptyWatchlist, err
	}

	members, err := extractWatchlistMembers(rows)
	if err != nil {
		return emptyWatchlist, err
	}

	return mapMembersToWatchlist(members)
}

const saveWatchlistQuery = `
	INSERT INTO watchlist(id, name, user_id, created_at) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT ON CONSTRAINT watchlist_name_user_id_key
	UPDATE name = $2`

// Save saves a watchlist.
func (wr *pgWatchlistRepo) Save(userID string, wl user.Watchlist) error {
	tx, err := wr.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(saveWatchlistQuery, wl.ID, wl.Name, userID, wl.CreatedAt)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	if len(wl.Stocks) == 0 {
		return tx.Commit()
	}

	err = saveStocks(tx, userID, wl.ID, wl.Stocks...)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	return tx.Commit()
}

// AddStock adds a stock to a users given watchlist.
func (wr *pgWatchlistRepo) AddStock(userID, symbol, watchlistID string) error {
	tx, err := wr.db.Begin()
	if err != nil {
		return err
	}

	newStock := stock.Stock{Symbol: symbol}
	err = saveStocks(tx, userID, watchlistID, newStock)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	return tx.Commit()
}

const deleteStockQuery = `
	DELETE watchlist_member WHERE symbol = $1 AND watchlist_id = $2`

// DeleteStock deletes a stock from a given users watchlist.
func (wr *pgWatchlistRepo) DeleteStock(userID, symbol, watchlistID string) error {
	tx, err := wr.db.Begin()
	if err != nil {
		return err
	}

	err = assertUserWatchlist(tx, userID, watchlistID)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	res, err := tx.Exec(deleteStockQuery, symbol, watchlistID)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	err = dbutil.AssertRowsAffected(res, 1, ErrNoSuchWatchlistStock)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	return tx.Commit()
}

const deleteWatchlistQuery = `
	DELETE watchlist WHERE id = $1 AND user_id = $2`

// Delete deletes a watchlist and all its members.
func (wr *pgWatchlistRepo) Delete(userID, watchlistID string) error {
	tx, err := wr.db.Begin()
	if err != nil {
		return err
	}

	err = deleteStocks(tx, userID, watchlistID)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	res, err := tx.Exec(deleteWatchlistQuery, watchlistID, userID)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	err = dbutil.AssertRowsAffected(res, 1, ErrNoSuchWatchlist)
	if err != nil {
		dbutil.RollbackTx(tx)
		return err
	}

	return tx.Commit()
}

const deleteStocksQuery = `
	DELETE watchlist_member WHERE watchlist_id = $1`

func deleteStocks(tx *sql.Tx, userID, watchlistID string) error {
	err := assertUserWatchlist(tx, userID, watchlistID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(deleteStockQuery, watchlistID)
	return err
}

const saveStockQuery = `
	INSERT INTO watchlist_member(symbol, watchlist_id, created_at)
	VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`

// Save saves a watchlist.
func saveStocks(tx *sql.Tx, userID, watchlistID string, stocks ...stock.Stock) error {
	err := assertUserWatchlist(tx, userID, watchlistID)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(saveStockQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stock := range stocks {
		now := time.Now().UTC()
		_, err = stmt.Exec(stock.Symbol, watchlistID, now)
		if err != nil {
			return err
		}
	}
	return nil
}

const asserUserWatchlistQuery = `
	SELECT w.id FROM watchlist w
	WHERE w.id = $1 AND w.user_id = $2`

func assertUserWatchlist(q dbutil.Querier, userID, watchlistID string) error {
	_, err := q.Query(asserUserWatchlistQuery, userID, watchlistID)
	if err == sql.ErrNoRows {
		return ErrNoSuchWatchlist
	}
	return err
}

// MockWatchlistRepo mock implementation for watchlist repo.
type MockWatchlistRepo struct {
	GetWatchlist      user.Watchlist
	GetErr            error
	GetArgUserID      string
	GetArgWatchlistID string

	SaveErr          error
	SaveArgUserID    string
	SaveArgWatchlist user.Watchlist

	AddStockErr            error
	AddStockArgUserID      string
	AddStockArgSymbol      string
	AddStockArgWatchlistID string

	DeleteStockErr            error
	DeleteStockArgUserID      string
	DeleteStockArgSymbol      string
	DeleteStockArgWatchlistID string

	DeleteErr            error
	DeleteArgUserID      string
	DeleteArgWatchlistID string
}

// UnsetArgs unsets all recorded arguments.
func (wr *MockWatchlistRepo) UnsetArgs() {
	wr.GetArgUserID = ""
	wr.GetArgWatchlistID = ""

	wr.SaveArgUserID = ""
	wr.SaveArgWatchlist = emptyWatchlist

	wr.AddStockArgUserID = ""
	wr.AddStockArgSymbol = ""
	wr.AddStockArgWatchlistID = ""

	wr.DeleteStockArgUserID = ""
	wr.DeleteStockArgSymbol = ""
	wr.DeleteStockArgWatchlistID = ""

	wr.DeleteArgUserID = ""
	wr.DeleteArgWatchlistID = ""
}

// Get mock implemntation of Get.
func (wr *MockWatchlistRepo) Get(userID, watchlistID string) (user.Watchlist, error) {
	wr.GetArgUserID = userID
	wr.GetArgWatchlistID = watchlistID

	return wr.GetWatchlist, wr.GetErr
}

// Save mock implementation of Save.
func (wr *MockWatchlistRepo) Save(userID string, watchlist user.Watchlist) error {
	wr.SaveArgUserID = userID
	wr.SaveArgWatchlist = watchlist

	return wr.SaveErr
}

// AddStock mock implementation of AddStock.
func (wr *MockWatchlistRepo) AddStock(userID, symbol, watchlistID string) error {
	wr.AddStockArgUserID = userID
	wr.AddStockArgSymbol = symbol
	wr.AddStockArgWatchlistID = watchlistID

	return wr.AddStockErr
}

// DeleteStock mock implementation of DeleteStock.
func (wr *MockWatchlistRepo) DeleteStock(userID, symbol, watchlistID string) error {
	wr.DeleteStockArgUserID = userID
	wr.DeleteStockArgSymbol = symbol
	wr.DeleteStockArgWatchlistID = watchlistID

	return wr.DeleteStockErr
}

// Delete mock implementation of Delete.
func (wr *MockWatchlistRepo) Delete(userID, watchlistID string) error {
	wr.DeleteArgUserID = userID
	wr.DeleteArgWatchlistID = watchlistID

	return wr.DeleteErr
}
