package main

import (
	"net/http"

	"github.com/mimir-news/pkg/httputil/auth"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
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

	err = e.ensureUserDoesNotExist(credentials.Email)
	if err != nil {
		c.Error(err)
		return
	}

	newUser, err := e.createNewUser(credentials)
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

	err = e.passwordSvc.Verify(credentials)
	if err != nil {
		c.Error(httputil.ErrUnauthorized())
		return
	}

	u, err := e.userRepo.FindByEmail(credentials.Email)
	if err != nil {
		c.Error(err)
		return
	}

	encodedToken, err := e.createSessionToken(u.User.ID, c.GetHeader(auth.ClientIDKey))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user.NewToken(encodedToken))
}

func (e *env) handleGetUser(c *gin.Context) {
	userID, err := getUserIDFromPath(c)
	if err != nil {
		c.Error(err)
		return
	}

	fullUser, err := e.userRepo.Find(userID)
	if err == repository.ErrNoSuchUser {
		c.Error(httputil.ErrNotFound())
		return
	}
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, fullUser.User)
}

func (e *env) createSessionToken(userID, clientID string) (string, error) {
	token, err := e.tokenSigner.New(userID, clientID)
	if err != nil {
		return "", err
	}

	session := domain.NewSession(userID)
	err = e.sessionRepo.Save(session)
	if err != nil {
		return "", err
	}

	return token, nil
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

func (e *env) createNewUser(credentials user.Credentials) (user.User, error) {
	secureCreds, err := e.passwordSvc.Create(credentials)
	if err != nil {
		return user.User{}, err
	}

	newUser := domain.NewUser(secureCreds, nil)

	err = e.userRepo.Save(newUser)
	if err != nil {
		return user.User{}, err
	}

	return newUser.User, nil
}

func (e *env) ensureUserDoesNotExist(email string) error {
	_, err := e.userRepo.FindByEmail(email)
	if err == repository.ErrNoSuchUser {
		return nil
	}
	if err == nil {
		return errUserAlreadyExists()
	}
	return err
}

func getCredentials(c *gin.Context) (user.Credentials, error) {
	var credentials user.Credentials
	err := c.BindJSON(&credentials)
	if err != nil {
		return credentials, httputil.ErrBadRequest()
	}
	return credentials, nil
}

func errUserAlreadyExists() error {
	return httputil.NewError("User already exists", http.StatusConflict)
}
