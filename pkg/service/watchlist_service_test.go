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
