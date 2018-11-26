package main

import (
	"fmt"
	"net/http"

	"github.com/mimir-news/pkg/httputil/auth"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/schema/user"
)

var errNotImplemented = httputil.NewError("not implemented", http.StatusNotImplemented)

func (e *env) handleUserCreation(c *gin.Context) {
	credentials, err := getCredentials(c)
	if err != nil {
		c.Error(err)
		return
	}

	newUser, err := e.userSvc.Create(credentials)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, newUser)
}

func (e *env) handleLogin(c *gin.Context) {
	credentials, err := getCredentials(c)
	if err != nil {
		c.Error(err)
		return
	}

	clientID := c.GetHeader(auth.ClientIDKey)
	token, err := e.userSvc.Authenticate(credentials, clientID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, token)
}

func (e *env) handleGetUser(c *gin.Context) {
	userID, err := getUserIDFromPath(c)
	if err != nil {
		c.Error(err)
		return
	}

	u, err := e.userSvc.Get(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (e *env) handleDeleteUser(c *gin.Context) {
	userID, err := getUserIDFromPath(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = e.userSvc.Delete(userID)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) handleChangePassword(c *gin.Context) {
	_, err := getUserIDFromPath(c)
	if err != nil {
		c.Error(err)
		return
	}

	change, err := getPasswordChange(c)
	fmt.Println(err)
	if err != nil {
		c.Error(err)
		return
	}

	err = e.userSvc.ChangePassword(change)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func getUserIDFromPath(c *gin.Context) (string, error) {
	authID, err := auth.GetUserID(c)
	if err != nil {
		return "", err
	}

	userID := c.Param("userId")
	if userID != authID {
		return "", httputil.ErrForbidden()
	}

	return userID, nil
}

func getCredentials(c *gin.Context) (user.Credentials, error) {
	var credentials user.Credentials
	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		return credentials, httputil.ErrBadRequest()
	}
	return credentials, nil
}

func getPasswordChange(c *gin.Context) (user.PasswordChange, error) {
	var change user.PasswordChange
	err := c.ShouldBindJSON(&change)
	if err != nil {
		return change, httputil.ErrBadRequest()
	}
	return change, nil
}

func errUserAlreadyExists() error {
	return httputil.NewError("User already exists", http.StatusConflict)
}
