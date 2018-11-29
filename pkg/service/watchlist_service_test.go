package service_test

import (
	"net/http"
	"testing"

	"github.com/mimir-news/pkg/httputil"

	"github.com/mimir-news/directory/pkg/service"

	"github.com/mimir-news/directory/pkg/repository"
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

	listRepo.GetArgUserID = ""
	listRepo.GetArgWatchlistID = ""
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
		listRepo.SaveArgUserID = ""
		listRepo.SaveArgWatchlist = user.Watchlist{}

		err := listSvc.Create(userID, listName)
		assert.NoError(err)

		savedListID := listRepo.SaveArgWatchlist.ID
		assert.Equal(userID, listRepo.SaveArgUserID)
		assert.NotEqual(lastListID, savedListID)
		assert.NotEqual("", listRepo.SaveArgWatchlist.ID)

		lastListID = savedListID
	}

	listRepo.SaveErr = repository.ErrNoSuchUser
	err := listSvc.Create(userID, listName)
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

		listRepo.AddStockArgUserID = ""
		listRepo.AddStockArgWatchlistID = ""
		listRepo.AddStockArgSymbol = ""

		err := listSvc.AddStock(userID, symbol, listID)
		assert.NoError(err)
		assert.Equal(userID, listRepo.AddStockArgUserID)
		assert.Equal(listID, listRepo.AddStockArgWatchlistID)
		assert.Equal(symbol, listRepo.AddStockArgSymbol)

	}

	listRepo.AddStockErr = repository.ErrNoSuchWatchlist
	err := listSvc.AddStock(userID, "S0", listID)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)

	listRepo.AddStockErr = repository.ErrNoSuchUser
	err = listSvc.AddStock(userID, "S0", listID)
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

		listRepo.DeleteStockArgUserID = ""
		listRepo.DeleteStockArgWatchlistID = ""
		listRepo.DeleteStockArgSymbol = ""

		err := listSvc.DeleteStock(userID, symbol, listID)
		assert.NoError(err)
		assert.Equal(userID, listRepo.DeleteStockArgUserID)
		assert.Equal(listID, listRepo.DeleteStockArgWatchlistID)
		assert.Equal(symbol, listRepo.DeleteStockArgSymbol)

	}

	listRepo.DeleteStockErr = repository.ErrNoSuchWatchlist
	err := listSvc.DeleteStock(userID, "S0", listID)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)

	listRepo.DeleteStockErr = repository.ErrNoSuchUser
	err = listSvc.DeleteStock(userID, "S0", listID)
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
