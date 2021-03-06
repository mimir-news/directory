package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
)

func (e *env) handleCreateWatchlist(c *gin.Context) {
	listName := c.Param("name")
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	watchlist, err := e.watchlistSvc.Create(userID, listName)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, watchlist)
}

func (e *env) handleDeleteWatchlist(c *gin.Context) {
	userID, listID, err := getUserAndWatchlistID(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = e.watchlistSvc.Delete(userID, listID)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) handleGetWatchlist(c *gin.Context) {
	userID, listID, err := getUserAndWatchlistID(c)
	if err != nil {
		c.Error(err)
		return
	}

	watchlist, err := e.watchlistSvc.Get(userID, listID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, watchlist)
}

func (e *env) handleRenameWatchlist(c *gin.Context) {
	newName := c.Param("name")
	userID, listID, err := getUserAndWatchlistID(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = e.watchlistSvc.Rename(userID, listID, newName)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) handleAddStockToWatchlist(c *gin.Context) {
	newStockSymbol := c.Param("symbol")
	userID, listID, err := getUserAndWatchlistID(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = e.watchlistSvc.AddStock(userID, listID, newStockSymbol)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) handleDeleteStockFromWatchlist(c *gin.Context) {
	stockSymbol := c.Param("symbol")
	userID, listID, err := getUserAndWatchlistID(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = e.watchlistSvc.DeleteStock(userID, listID, stockSymbol)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func getUserAndWatchlistID(c *gin.Context) (string, string, error) {
	listID := c.Param("watchlistId")
	userID, err := auth.GetUserID(c)
	return userID, listID, err
}
