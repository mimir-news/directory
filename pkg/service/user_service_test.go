package service_test

import (
	"net/http"
	"testing"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	assert := assert.New(t)

	userID := "user-id"
	expectedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: "mail@mail.com",
		},
	}

	userRepo := &mockUserRepo{
		findUser: expectedUser,
	}
	userSvc := service.NewUserService(nil, nil, userRepo, nil)

	u, err := userSvc.Get(userID)
	assert.NoError(err)
	assert.Equal(userID, u.ID)
	assert.Equal(expectedUser.User.Email, u.Email)
	assert.Equal(userID, userRepo.findArg)

	userRepo = &mockUserRepo{
		findErr: repository.ErrNoSuchUser,
	}
	userSvc = service.NewUserService(nil, nil, userRepo, nil)

	u, err = userSvc.Get(userID)
	assert.Error(err)
	assert.Equal("", u.ID)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)
	assert.Equal(userID, userRepo.findArg)

	userRepo = &mockUserRepo{
		findErr: testError,
	}
	userSvc = service.NewUserService(nil, nil, userRepo, nil)

	u, err = userSvc.Get(userID)
	assert.Equal(testError, err)
	assert.Equal("", u.ID)
}
