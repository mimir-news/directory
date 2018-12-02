package service_test

import (
	"net/http"
	"testing"

	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/id"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

func TestGetWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	listID := id.New()
	expectedList := user.Watchlist{
		ID:   listID,
		Name: "l-0",
	}

	listRepo := &repository.MockWatchlistRepo{
		GetWatchlist: expectedList,
	}

	listSvc := service.NewWatchlistService(listRepo)

	l, err := listSvc.Get(userID, listID)
	assert.NoError(err)
	assert.Equal(listID, l.ID)
	assert.Equal(expectedList.Name, l.Name)
	assert.Equal(userID, listRepo.GetArgUserID)
	assert.Equal(listID, listRepo.GetArgWatchlistID)

	listRepo.UnsetArgs()
	listRepo.GetErr = repository.ErrNoSuchWatchlist

	l, err = listSvc.Get(userID, listID)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)
	assert.Equal("", l.ID)
	assert.Equal("", l.Name)
	assert.Equal(userID, listRepo.GetArgUserID)
	assert.Equal(listID, listRepo.GetArgWatchlistID)
}

func TestCreateWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	listName := "list-name"

	listRepo := &repository.MockWatchlistRepo{}
	listSvc := service.NewWatchlistService(listRepo)

	var lastListID string
	for i := 0; i < 3; i++ {
		listRepo.UnsetArgs()

		list, err := listSvc.Create(userID, listName)
		assert.NoError(err)

		assert.Equal(userID, listRepo.SaveArgUserID)
		assert.NotEqual(lastListID, list.ID)
		assert.NotEqual("", listRepo.SaveArgWatchlist.ID)

		lastListID = list.ID
	}

	listRepo.SaveErr = repository.ErrNoSuchUser
	_, err := listSvc.Create(userID, listName)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)
}

func TestAddStockToWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	listID := id.New()
	symbols := []string{"S0", "S1", "S3"}

	listRepo := &repository.MockWatchlistRepo{}
	listSvc := service.NewWatchlistService(listRepo)

	for _, symbol := range symbols {
		listRepo.UnsetArgs()

		err := listSvc.AddStock(userID, listID, symbol)
		assert.NoError(err)
		assert.Equal(userID, listRepo.AddStockArgUserID)
		assert.Equal(listID, listRepo.AddStockArgWatchlistID)
		assert.Equal(symbol, listRepo.AddStockArgSymbol)

	}

	listRepo.AddStockErr = repository.ErrNoSuchWatchlist
	err := listSvc.AddStock(userID, listID, "S0")
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)

	listRepo.AddStockErr = repository.ErrNoSuchUser
	err = listSvc.AddStock(userID, listID, "S0")
	assert.Error(err)
	httpErr, ok = err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)
}

func TestDeleteStockFromWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	listID := id.New()
	symbols := []string{"S0", "S1", "S3"}

	listRepo := &repository.MockWatchlistRepo{}
	listSvc := service.NewWatchlistService(listRepo)

	for _, symbol := range symbols {
		listRepo.UnsetArgs()

		err := listSvc.DeleteStock(userID, listID, symbol)
		assert.NoError(err)
		assert.Equal(userID, listRepo.DeleteStockArgUserID)
		assert.Equal(listID, listRepo.DeleteStockArgWatchlistID)
		assert.Equal(symbol, listRepo.DeleteStockArgSymbol)

	}

	listRepo.DeleteStockErr = repository.ErrNoSuchWatchlist
	err := listSvc.DeleteStock(userID, listID, "S0")
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)

	listRepo.DeleteStockErr = repository.ErrNoSuchUser
	err = listSvc.DeleteStock(userID, listID, "S0")
	assert.Error(err)
	httpErr, ok = err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)
}

func TestDeleteWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	listID := id.New()

	listRepo := &repository.MockWatchlistRepo{}
	listSvc := service.NewWatchlistService(listRepo)

	err := listSvc.Delete(userID, listID)
	assert.NoError(err)
	assert.Equal(userID, listRepo.DeleteArgUserID)
	assert.Equal(listID, listRepo.DeleteArgWatchlistID)

	listRepo.DeleteErr = repository.ErrNoSuchWatchlist
	err = listSvc.Delete(userID, listID)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)

	listRepo.DeleteErr = repository.ErrNoSuchUser
	err = listSvc.Delete(userID, listID)
	assert.Error(err)
	httpErr, ok = err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)
}

func TestRenameWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	listID := id.New()
	newName := "new-list-name"

	listRepo := &repository.MockWatchlistRepo{}
	listSvc := service.NewWatchlistService(listRepo)

	err := listSvc.Rename(userID, listID, newName)
	assert.NoError(err)

	savedList := listRepo.SaveArgWatchlist
	assert.Equal(userID, listRepo.SaveArgUserID)
	assert.Equal(listID, savedList.ID)
	assert.Equal(newName, savedList.Name)
}
