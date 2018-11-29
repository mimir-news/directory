package service

import (
	"net/http"

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
	Rename(userID, watchlistID, newName string) error
	AddStock(userID, watchlistID, symbol string) error
	DeleteStock(userID, watchlistID, symbol string) error
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
		return emptyWatchlist, httputil.NewError(err.Error(), http.StatusNotFound)
	}

	return list, err
}

// Create creates and saves a new watchlist.
func (ws *watchlistSvc) Create(userID, listName string) error {
	newList := user.NewWatchlist(listName)
	return ws.saveList(userID, newList)
}

// Create creates and saves a new watchlist.
func (ws *watchlistSvc) Rename(userID, watchlistID, newName string) error {
	renamedList := user.Watchlist{
		ID:   watchlistID,
		Name: newName,
	}

	return ws.saveList(userID, renamedList)
}

// AddStock adds a stock to a watchlist.
func (ws *watchlistSvc) AddStock(userID, watchlistID, symbol string) error {
	err := ws.listRepo.AddStock(userID, symbol, watchlistID)
	if err == repository.ErrNoSuchWatchlist || err == repository.ErrNoSuchUser {
		return httputil.NewError(err.Error(), http.StatusNotFound)
	}

	return err
}

// DeleteStock removes a stock form a watchlist.
func (ws *watchlistSvc) DeleteStock(userID, watchlistID, symbol string) error {
	err := ws.listRepo.DeleteStock(userID, symbol, watchlistID)
	if err == repository.ErrNoSuchWatchlist || err == repository.ErrNoSuchUser {
		return httputil.NewError(err.Error(), http.StatusNotFound)
	}

	return err
}

// Delete deletes a watchlist.
func (ws *watchlistSvc) Delete(userID, watchlistID string) error {
	err := ws.listRepo.Delete(userID, watchlistID)
	if err == repository.ErrNoSuchWatchlist || err == repository.ErrNoSuchUser {
		return httputil.NewError(err.Error(), http.StatusNotFound)
	}

	return err
}

func (ws *watchlistSvc) saveList(userID string, watchlist user.Watchlist) error {
	err := ws.listRepo.Save(userID, watchlist)
	if err == repository.ErrNoSuchUser {
		return httputil.NewError(err.Error(), http.StatusNotFound)
	}

	if err == repository.ErrWatchlistExist {
		return httputil.NewError(err.Error(), http.StatusConflict)
	}

	return err
}
