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

	err = e.watchlistSvc.Create(userID, listName)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
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
	c.Error(errNotImplemented)
}

func (e *env) handleDeleteStockFromWatchlist(c *gin.Context) {
	c.Error(errNotImplemented)
}

func getUserAndWatchlistID(c *gin.Context) (string, string, error) {
	listID := c.Param("watchlistId")
	userID, err := auth.GetUserID(c)
	return userID, listID, err
}
