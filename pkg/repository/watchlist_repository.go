package repository

import (
	"errors"

	"github.com/mimir-news/pkg/schema/user"
)

// Common watchlist related errors.
var (
	ErrNoSuchWatchlist = errors.New("No such watchlist")
	ErrWatchlistExist  = errors.New("Watchlist already exists")
)

// WatchlistRepo interface for getting and storing watchlists in a database.
type WatchlistRepo interface {
	Get(userID, watchlistID string) (user.Watchlist, error)
	Save(userID string, watchlist user.Watchlist) error
	AddStock(userID, symbol, watchlistID string) error
	DeleteStock(userID, symbol, watchlistID string) error
	Delete(userID, watchlistID string) error
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
