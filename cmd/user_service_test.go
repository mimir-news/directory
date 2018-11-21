package main

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/mimir-news/pkg/id"

	"github.com/mimir-news/pkg/httputil/auth"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

func TestUserCreation(t *testing.T) {
	assert := assert.New(t)

	conf := getTestConfig()
	userRepo := &mockUserRepo{
		findByEmailErr: repository.ErrNoSuchUser,
	}
	mockEnv := getTestEnv(conf, userRepo, nil)

	credentials := user.Credentials{
		Email:    "mail@mail.com",
		Password: "super-secret-password",
	}

	expectedUser := user.User{
		Email:     credentials.Email,
		CreatedAt: time.Now().UTC(),
	}

	server := newServer(mockEnv, conf)

	req := createTestPostRequest("client-id", "", "/v1/users", credentials)
	res := performTestRequest(server.Handler, req)

	assert.Equal(http.StatusOK, res.Code)
	var u user.User
	err := json.NewDecoder(res.Body).Decode(&u)
	assert.NoError(err)
	assert.Equal(expectedUser.Email, u.Email)
	assert.Equal(expectedUser.Watchlists, u.Watchlists)
	assert.True(u.CreatedAt.After(expectedUser.CreatedAt))
	assert.Equal(credentials.Email, userRepo.saveArg.User.Email)
	assert.NotEqual("", userRepo.saveArg.Credentials.Password)
	assert.NotEqual(credentials.Password, userRepo.saveArg.Credentials.Password)

	userRepo = &mockUserRepo{
		findByEmailUser: domain.FullUser{User: u},
		findByEmailErr:  nil,
	}
	mockEnv = getTestEnv(conf, userRepo, nil)
	server = newServer(mockEnv, conf)

	req = createTestPostRequest("client-id", "", "/v1/users", credentials)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusConflict, res.Code)
}

func TestHandleLogin(t *testing.T) {
	assert := assert.New(t)

	credentials := user.Credentials{
		Email:    "mail@mail.com",
		Password: "super-secret-password",
	}

	expectedUser := domain.FullUser{
		User: user.User{
			ID:        "user-id",
			Email:     credentials.Email,
			CreatedAt: time.Now().UTC(),
		},
		Credentials: domain.StoredCredentials{
			Email:    credentials.Email,
			Password: "S5UeZOWCDkIfP/5LUDpyhIY0l6+aow+CmkBEVtHqpebhe04vb6kDbPaD/wo05fs6x1lvJfI/6YZ66zbQ8X2lHaEThp4f1Zl0exk7j/wow740KbWZHf9DSA==", // Hashed and encrypted password.
			Salt:     "3MQEKd3NVnU+WQFQxo8JpYWrTrqXOiwro4MwLwnsckWXinE=",                                                                         // Encrypted salt
		},
	}

	conf := getTestConfig()
	userRepo := &mockUserRepo{
		findByEmailUser: expectedUser,
	}
	sessionRepo := &mockSessionRepo{}
	mockEnv := getTestEnv(conf, userRepo, sessionRepo)
	server := newServer(mockEnv, conf)

	req := createTestPostRequest("client-id", "", "/v1/login", credentials)
	res := performTestRequest(server.Handler, req)

	assert.Equal(http.StatusOK, res.Code)
	var token user.Token
	err := json.NewDecoder(res.Body).Decode(&token)
	assert.NoError(err)
	verifier := auth.NewVerifier(conf.TokenSecret, conf.TokenVerificationKey)
	authToken, err := verifier.Verify("client-id", token.Token)
	assert.NoError(err)
	assert.Equal(expectedUser.User.ID, authToken.Body.Subject)
	assert.Equal(expectedUser.User.ID, sessionRepo.saveArg.UserID)

	wrongCredentials := user.Credentials{
		Email:    "mail@mail.com",
		Password: "super-wrong-password",
	}
	sessionRepo.saveArg = domain.Session{}
	req = createTestPostRequest("client-id", "", "/v1/login", wrongCredentials)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusUnauthorized, res.Code)
	assert.Equal("", sessionRepo.saveArg.UserID)
}

func TestHandleGetUser(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()
	expectedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: "mail@mail.com",
		},
	}

	conf := getTestConfig()
	userRepo := &mockUserRepo{
		findUser: expectedUser,
	}
	mockEnv := getTestEnv(conf, userRepo, nil)
	authToken, err := mockEnv.tokenSigner.New(userID, clientID)
	assert.NoError(err)

	// Setup: Get user happy path.
	server := newServer(mockEnv, conf)
	req := createTestGetRequest(clientID, authToken, "/v1/users/"+userID)
	res := performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusOK, res.Code)
	var user user.User
	err = json.NewDecoder(res.Body).Decode(&user)
	assert.NoError(err)
	assert.Equal(expectedUser.User.ID, user.ID)
	assert.Equal(expectedUser.User.ID, userRepo.findArg)

	// Setup: Missing token.
	userRepo.findArg = ""
	req = createTestGetRequest(clientID, "", "/v1/users/"+userID)
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusUnauthorized, res.Code)
	assert.Equal("", userRepo.findArg)

	// Setup: Missmatching user ids.
	userRepo.findArg = ""
	req = createTestGetRequest(clientID, authToken, "/v1/users/wrong-user-id")
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusForbidden, res.Code)
	assert.Equal("", userRepo.findArg)

	// Setup: No user found.
	userRepo.findUser = domain.FullUser{}
	userRepo.findErr = repository.ErrNoSuchUser
	userRepo.findArg = ""
	req = createTestGetRequest(clientID, authToken, "/v1/users/"+userID)
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusNotFound, res.Code)
	assert.Equal(userID, userRepo.findArg)

}
