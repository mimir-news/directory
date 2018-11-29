package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mimir-news/pkg/schema/stock"

	"github.com/mimir-news/directory/pkg/repository"

	"github.com/mimir-news/pkg/id"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

func TestHandleCreateWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()
	listName := "my-list"

	listRepo := &repository.MockWatchlistRepo{}

	conf := getTestConfig()
	e := getTestEnv(conf, nil, nil, listRepo)
	authToken := getTestToken(conf, userID, clientID)
	server := newServer(e, conf)

	req := createTestPostRequest(clientID, authToken, "/v1/watchlists/"+listName, nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)
	assert.Equal(userID, listRepo.SaveArgUserID)
	assert.Equal(listName, listRepo.SaveArgWatchlist.Name)
}

func TestHandleDeleteWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()
	listID := id.New()

	listRepo := &repository.MockWatchlistRepo{}

	conf := getTestConfig()
	e := getTestEnv(conf, nil, nil, listRepo)
	authToken := getTestToken(conf, userID, clientID)
	server := newServer(e, conf)

	req := createTestDeleteRequest(clientID, authToken, "/v1/watchlists/"+listID)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)
	assert.Equal(userID, listRepo.DeleteArgUserID)
	assert.Equal(listID, listRepo.DeleteArgWatchlistID)

	listRepo.UnsetArgs()
	listRepo.DeleteErr = repository.ErrNoSuchWatchlist

	req = createTestDeleteRequest(clientID, authToken, "/v1/watchlists/"+listID)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusNotFound, res.Code)
	assert.Equal(userID, listRepo.DeleteArgUserID)
	assert.Equal(listID, listRepo.DeleteArgWatchlistID)
}

func TestHandleGetWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()
	listID := id.New()

	expectedList := user.Watchlist{
		ID:   listID,
		Name: "my-list",
		Stocks: []stock.Stock{
			stock.Stock{Symbol: "S0", Name: "my-stock"},
		},
	}

	listRepo := &repository.MockWatchlistRepo{
		GetWatchlist: expectedList,
	}

	conf := getTestConfig()
	e := getTestEnv(conf, nil, nil, listRepo)
	authToken := getTestToken(conf, userID, clientID)
	server := newServer(e, conf)

	req := createTestGetRequest(clientID, authToken, "/v1/watchlists/"+listID)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)
	assert.Equal(userID, listRepo.GetArgUserID)
	assert.Equal(listID, listRepo.GetArgWatchlistID)

	var wl user.Watchlist
	err := json.NewDecoder(res.Body).Decode(&wl)
	assert.NoError(err)
	assert.Equal(listID, wl.ID)
	assert.Equal(expectedList.Name, wl.Name)
	assert.Equal(expectedList.Stocks[0].Symbol, wl.Stocks[0].Symbol)

	listRepo.UnsetArgs()
	listRepo.GetErr = repository.ErrNoSuchWatchlist

	req = createTestGetRequest(clientID, authToken, "/v1/watchlists/"+listID)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusNotFound, res.Code)
}

func TestHandleRenameWatchlist(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	listID := id.New()
	clientID := id.New()
	newName := "my-list"

	listRepo := &repository.MockWatchlistRepo{}

	conf := getTestConfig()
	e := getTestEnv(conf, nil, nil, listRepo)
	authToken := getTestToken(conf, userID, clientID)
	server := newServer(e, conf)

	req := createTestPutRequest(clientID, authToken, "/v1/watchlists/"+listID+"/name/"+newName, nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)
	assert.Equal(userID, listRepo.SaveArgUserID)
	assert.Equal(newName, listRepo.SaveArgWatchlist.Name)
}
