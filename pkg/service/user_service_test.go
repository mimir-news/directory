package service_test

import (
	"net/http"
	"testing"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/id"
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

func TestDeleteUser(t *testing.T) {
	assert := assert.New(t)

	userID := "user-id"

	userRepo := &mockUserRepo{
		deleteErr: repository.ErrNoSuchUser,
	}
	userSvc := service.NewUserService(nil, nil, userRepo, nil)

	err := userSvc.Delete(userID)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)

	userRepo = &mockUserRepo{
		deleteErr: testError,
	}
	userSvc = service.NewUserService(nil, nil, userRepo, nil)

	err = userSvc.Delete(userID)
	assert.Equal(testError, err)
	assert.Equal(userID, userRepo.deleteArg)
}

func TestUserSvcChangePassword(t *testing.T) {
	assert := assert.New(t)

	storedUser := domain.FullUser{
		User: user.User{
			ID:    id.New(),
			Email: "mail@mail.com",
		},
		Credentials: domain.StoredCredentials{
			Email:    "mail@mail.com",
			Password: "S5UeZOWCDkIfP/5LUDpyhIY0l6+aow+CmkBEVtHqpebhe04vb6kDbPaD/wo05fs6x1lvJfI/6YZ66zbQ8X2lHaEThp4f1Zl0exk7j/wow740KbWZHf9DSA==", // Hashed and encrypted password.
			Salt:     "3MQEKd3NVnU+WQFQxo8JpYWrTrqXOiwro4MwLwnsckWXinE=",                                                                         // Encrypted salt
		},
	}

	userRepo := &mockUserRepo{
		findByEmailUser: storedUser,
	}

	passwordSvc := service.NewPasswordService(userRepo, "my-pepper", "my-encryption-key")
	userSvc := service.NewUserService(passwordSvc, nil, userRepo, nil)

	pwdChange := user.PasswordChange{
		New:      "new-password",
		Repeated: "new-password",
		Old: user.Credentials{
			Email:    storedUser.User.Email,
			Password: "super-secret-password",
		},
	}

	err := userSvc.ChangePassword(pwdChange)
	assert.NoError(err)
	assert.Equal(pwdChange.Old.Email, userRepo.findByEmailArg)
	savedCreds := userRepo.saveArg.Credentials
	assert.NotEqual(storedUser.Credentials.Salt, savedCreds.Salt)
	assert.NotEqual("", savedCreds.Salt)
	assert.NotEqual(storedUser.Credentials.Password, savedCreds.Password)
	assert.NotEqual("", savedCreds.Password)
	assert.Equal(storedUser.Credentials.Email, savedCreds.Email)

	inconsistentPwdChange := user.PasswordChange{
		New:      "new-password",
		Repeated: "other-new-password",
		Old: user.Credentials{
			Email:    storedUser.User.Email,
			Password: "super-secret-password",
		},
	}

	userRepo.findByEmailArg = ""
	userRepo.saveArg = domain.FullUser{}
	err = userSvc.ChangePassword(inconsistentPwdChange)
	assert.Error(err)
	httpError, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusBadRequest, httpError.StatusCode)
	assert.Equal("", userRepo.findByEmailArg)
	savedCreds = userRepo.saveArg.Credentials
	assert.Equal("", savedCreds.Salt)
	assert.Equal("", savedCreds.Password)
	assert.Equal("", savedCreds.Email)

	wrongPwdChange := user.PasswordChange{
		New:      "new-password",
		Repeated: "new-password",
		Old: user.Credentials{
			Email:    storedUser.User.Email,
			Password: "wrong-password",
		},
	}

	userRepo.findByEmailArg = ""
	userRepo.saveArg = domain.FullUser{}
	err = userSvc.ChangePassword(wrongPwdChange)
	assert.Error(err)
	assert.Equal(wrongPwdChange.Old.Email, userRepo.findByEmailArg)
	savedCreds = userRepo.saveArg.Credentials
	assert.Equal("", savedCreds.Salt)
	assert.Equal("", savedCreds.Password)
	assert.Equal("", savedCreds.Email)

}
