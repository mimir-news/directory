package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
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

	token, err := e.userSvc.Authenticate(credentials)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, token)
}

func (e *env) handleTokenRenewal(c *gin.Context) {
	oldToken, err := getRefreshToken(c)
	if err != nil {
		c.Error(err)
		return
	}

	newToken, err := e.userSvc.RefreshToken(oldToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, newToken)
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

func (e *env) handleChangeEmail(c *gin.Context) {
	userID, err := getUserIDFromPath(c)
	if err != nil {
		c.Error(err)
		return
	}

	u, err := getUserFromBody(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = e.userSvc.ChangeEmail(userID, u.Email)
	if err != nil {
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) getAnonymousToken(c *gin.Context) {
	token, err := e.userSvc.GetAnonymousToken()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, token)
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

func getUserFromBody(c *gin.Context) (user.User, error) {
	var u user.User
	err := c.ShouldBindJSON(&u)
	if err != nil {
		return u, httputil.ErrBadRequest()
	}
	if !u.Valid() {
		return u, httputil.ErrBadRequest()
	}
	return u, nil
}

func getCredentials(c *gin.Context) (user.Credentials, error) {
	var credentials user.Credentials
	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		return credentials, httputil.ErrBadRequest()
	}
	if !credentials.Valid() {
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
	if !change.Valid() {
		return change, httputil.ErrBadRequest()
	}
	return change, nil
}

func getRefreshToken(c *gin.Context) (user.Token, error) {
	var token user.Token
	err := c.ShouldBindJSON(&token)
	if err != nil {
		return token, httputil.ErrBadRequest()
	}
	if token.Token == "" || token.RefreshToken == "" {
		return token, httputil.ErrBadRequest()
	}

	return token, nil
}

func errUserAlreadyExists() error {
	return httputil.NewError("User already exists", http.StatusConflict)
}
