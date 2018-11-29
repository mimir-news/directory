package service

import (
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/schema/user"
)

var (
	emptyWatchlist = user.Watchlist{}
)

// WatchlistService service responsible for handling watchlists
type WatchlistService interface {
	Get(userID, watchlistID string) (user.Watchlist, error)
	Create(userID, listName string) error
	AddStock(userID, symbol, watchlistID string) error
	DeleteStock(userID, symbol, watchlistID string) error
	Delete(userID, watchlistID string) error
}

// NewWatchlistService returns the default implemntation of WatcklistService.
func NewWatchlistService(listRepo repository.WatchlistRepo) WatchlistService {
	return &watchlistSvc{
		listRepo: listRepo,
	}
}

// watchlistSvc default implementation of WatchlistService
type watchlistSvc struct {
	listRepo repository.WatchlistRepo
}

// Get gets a watchlist of a given id belonging to a given user.
func (ws *watchlistSvc) Get(userID, watchlistID string) (user.Watchlist, error) {
	list, err := ws.listRepo.Get(userID, watchlistID)
	if err == repository.ErrNoSuchWatchlist {
		return emptyWatchlist, httputil.ErrNotFound()
	}

	return list, err
}

func (ws *watchlistSvc) Create(userID, listName string) error {
	return nil
}

func (ws *watchlistSvc) AddStock(userID, symbol, watchlistID string) error {
	return nil
}

func (ws *watchlistSvc) DeleteStock(userID, symbol, watchlistID string) error {
	return nil
}

func (ws *watchlistSvc) Delete(userID, watchlistID string) error {
	return nil
}
