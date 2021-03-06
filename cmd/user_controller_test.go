package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/pkg/httputil/auth"
	"github.com/mimir-news/pkg/id"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

func TestUserCreation(t *testing.T) {
	assert := assert.New(t)

	conf := getTestConfig()
	userRepo := &repository.MockUserRepo{
		FindByEmailErr: repository.ErrNoSuchUser,
	}
	mockEnv := getTestEnv(conf, userRepo, nil, nil)

	credentials := user.Credentials{
		Email:    "mail@mail.com",
		Password: "super-secret-password",
	}

	expectedUser := user.User{
		Email:     credentials.Email,
		Role:      auth.UserRole,
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
	assert.Equal(expectedUser.Role, u.Role)
	assert.Equal(expectedUser.Watchlists, u.Watchlists)
	assert.True(u.CreatedAt.After(expectedUser.CreatedAt))
	assert.Equal(credentials.Email, userRepo.SaveArg.User.Email)
	assert.NotEqual("", userRepo.SaveArg.Credentials.Password)
	assert.NotEqual(credentials.Password, userRepo.SaveArg.Credentials.Password)

	userRepo = &repository.MockUserRepo{
		FindByEmailUser: domain.FullUser{User: u},
		FindByEmailErr:  nil,
	}
	mockEnv = getTestEnv(conf, userRepo, nil, nil)
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
			Role:      auth.UserRole,
			CreatedAt: time.Now().UTC(),
		},
		Credentials: domain.StoredCredentials{
			Email:    credentials.Email,
			Password: encryptedPassword,
			Salt:     encryptedSalt,
		},
	}

	conf := getTestConfig()
	userRepo := &repository.MockUserRepo{
		FindByEmailUser: expectedUser,
	}
	sessionRepo := &repository.MockSessionRepo{}
	mockEnv := getTestEnv(conf, userRepo, sessionRepo, nil)
	server := newServer(mockEnv, conf)

	req := createTestPostRequest("client-id", "", "/v1/login", credentials)
	res := performTestRequest(server.Handler, req)

	assert.Equal(http.StatusOK, res.Code)
	var token user.Token
	err := json.NewDecoder(res.Body).Decode(&token)
	assert.NoError(err)
	verifier := auth.NewVerifier(conf.JWTCredentials, 0)
	authToken, err := verifier.Verify(token.Token)
	assert.NoError(err)
	assert.Equal(expectedUser.User.ID, authToken.User.ID)
	assert.Equal(expectedUser.User.ID, sessionRepo.SaveArg.UserID)
	assert.Equal(expectedUser.User.Role, token.User.Role)
	assert.Equal(sessionRepo.SaveArg.RefreshToken, token.RefreshToken)
	assert.Equal(sessionRepo.SaveArg.ID, authToken.ID)

	wrongCredentials := user.Credentials{
		Email:    "mail@mail.com",
		Password: "super-wrong-password",
	}
	sessionRepo.SaveArg = domain.Session{}
	req = createTestPostRequest("client-id", "", "/v1/login", wrongCredentials)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusUnauthorized, res.Code)
	assert.Equal("", sessionRepo.SaveArg.UserID)
}

func TestHandleGetUser(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()
	expectedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: "mail@mail.com",
			Role:  auth.UserRole,
		},
	}

	conf := getTestConfig()
	userRepo := &repository.MockUserRepo{
		FindUser: expectedUser,
	}
	mockEnv := getTestEnv(conf, userRepo, nil, nil)
	authToken := getTestToken(conf, userID, clientID)

	// Setup: Get user happy path.
	server := newServer(mockEnv, conf)
	req := createTestGetRequest(clientID, authToken, "/v1/users/"+userID)
	res := performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusOK, res.Code)
	var user user.User
	err := json.NewDecoder(res.Body).Decode(&user)
	assert.NoError(err)
	assert.Equal(expectedUser.User.ID, user.ID)
	assert.Equal(expectedUser.User.Role, user.Role)
	assert.Equal(expectedUser.User.ID, userRepo.FindArg)

	// Setup: Missing token.
	userRepo.FindArg = ""
	req = createTestGetRequest(clientID, "", "/v1/users/"+userID)
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusUnauthorized, res.Code)
	assert.Equal("", userRepo.FindArg)

	// Setup: Missmatching user ids.
	userRepo.FindArg = ""
	req = createTestGetRequest(clientID, authToken, "/v1/users/wrong-user-id")
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusForbidden, res.Code)
	assert.Equal("", userRepo.FindArg)

	// Setup: No user found.
	userRepo.FindUser = domain.FullUser{}
	userRepo.FindErr = repository.ErrNoSuchUser
	userRepo.FindArg = ""
	req = createTestGetRequest(clientID, authToken, "/v1/users/"+userID)
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusNotFound, res.Code)
	assert.Equal(userID, userRepo.FindArg)

}

func TestHandleDeleteUser(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()

	conf := getTestConfig()
	userRepo := &repository.MockUserRepo{}
	mockEnv := getTestEnv(conf, userRepo, nil, nil)
	signer := getTestSigner(conf)
	authToken, err := signer.Sign(id.New(), auth.User{ID: userID, Role: auth.UserRole})
	assert.NoError(err)

	// Setup: Get user happy path.
	server := newServer(mockEnv, conf)
	req := createTestDeleteRequest(clientID, authToken, "/v1/users/"+userID)
	res := performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusOK, res.Code)
	assert.Equal(userID, userRepo.DeleteArg)

	// Setup: Missing token.
	userRepo.DeleteArg = ""
	req = createTestDeleteRequest(clientID, "", "/v1/users/"+userID)
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusUnauthorized, res.Code)
	assert.Equal("", userRepo.DeleteArg)

	// Setup: Missmatching user ids.
	userRepo.FindArg = ""
	req = createTestDeleteRequest(clientID, authToken, "/v1/users/wrong-user-id")
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusForbidden, res.Code)
	assert.Equal("", userRepo.FindArg)

	// Setup: No user found.
	userRepo.DeleteErr = repository.ErrNoSuchUser
	userRepo.DeleteArg = ""
	req = createTestDeleteRequest(clientID, authToken, "/v1/users/"+userID)
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusNotFound, res.Code)
	assert.Equal(userID, userRepo.DeleteArg)

}

func TestHandlePasswordChange(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()
	userEmail := "main@mail.com"

	expectedUser := domain.FullUser{
		User: user.User{
			ID:        userID,
			Email:     userEmail,
			CreatedAt: time.Now().UTC(),
		},
		Credentials: domain.StoredCredentials{
			Email:    userEmail,
			Password: encryptedPassword,
			Salt:     encryptedSalt,
		},
	}

	conf := getTestConfig()
	userRepo := &repository.MockUserRepo{
		FindByEmailUser: expectedUser,
	}
	mockEnv := getTestEnv(conf, userRepo, nil, nil)
	authToken := getTestToken(conf, userID, clientID)

	pwdChange := user.PasswordChange{
		New:      "new-password",
		Repeated: "new-password",
		Old: user.Credentials{
			Email:    userEmail,
			Password: correctPassword,
		},
	}

	// Setup: Change password happy path.
	server := newServer(mockEnv, conf)
	req := createTestPutRequest(clientID, authToken, "/v1/users/"+userID+"/password", pwdChange)
	res := performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusOK, res.Code)
	savedUser := userRepo.SaveArg
	assert.Equal(userID, savedUser.User.ID)
	assert.NotEqual(expectedUser.Credentials.Password, savedUser.Credentials.Password)
	assert.NotEqual("", savedUser.Credentials.Password)
	assert.NotEqual(expectedUser.Credentials.Salt, savedUser.Credentials.Salt)
	assert.NotEqual("", savedUser.Credentials.Salt)

	// Setup: Change password wrong user id.
	userRepo.SaveArg = domain.FullUser{}
	req = createTestPutRequest(clientID, authToken, "/v1/users/wrong-id/password", pwdChange)
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusForbidden, res.Code)
	savedUser = userRepo.SaveArg
	assert.Equal("", savedUser.User.ID)

	// Setup: Change password no change provided.
	userRepo.SaveArg = domain.FullUser{}
	req = createTestPutRequest(clientID, authToken, "/v1/users/"+userID+"/password", user.PasswordChange{})
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusBadRequest, res.Code)
	savedUser = userRepo.SaveArg
	assert.Equal("", savedUser.User.ID)

}

func TestHandleEmailChange(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	clientID := id.New()
	userEmail := "main@mail.com"

	expectedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: userEmail,
		},
		Credentials: domain.StoredCredentials{
			Email: userEmail,
		},
	}

	conf := getTestConfig()
	userRepo := &repository.MockUserRepo{
		FindUser: expectedUser,
	}
	mockEnv := getTestEnv(conf, userRepo, nil, nil)
	authToken := getTestToken(conf, userID, clientID)

	u := user.User{
		ID:    userID,
		Email: userEmail,
	}

	// Setup: Change email happy path.
	server := newServer(mockEnv, conf)
	req := createTestPutRequest(clientID, authToken, "/v1/users/"+userID+"/email", u)
	res := performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusOK, res.Code)
	savedUser := userRepo.SaveArg
	assert.Equal(u.Email, savedUser.User.Email)
	assert.Equal(u.Email, savedUser.Credentials.Email)
	assert.Equal(userID, savedUser.User.ID)
	assert.Equal(userID, userRepo.FindArg)

	// Setup: Change password no email provided.
	userRepo.SaveArg = domain.FullUser{}
	userRepo.FindArg = ""
	req = createTestPutRequest(clientID, authToken, "/v1/users/"+userID+"/email", user.User{})
	res = performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusBadRequest, res.Code)
	assert.Equal("", userRepo.SaveArg.User.ID)
	assert.Equal("", userRepo.FindArg)

}

func TestGetAnonymousToken(t *testing.T) {
	assert := assert.New(t)

	cfg := getTestConfig()
	verifier := auth.NewVerifier(cfg.JWTCredentials, 0)
	mockEnv := getTestEnv(cfg, nil, nil, nil)

	// Setup: Get anonymous token happy path.
	server := newServer(mockEnv, cfg)
	req := createTestGetRequest("", "", "/v1/login/anonymous")
	res := performTestRequest(server.Handler, req)
	// Test
	assert.Equal(http.StatusOK, res.Code)
	var token user.Token
	err := json.NewDecoder(res.Body).Decode(&token)
	assert.NoError(err)
	assert.Equal(auth.AnonymousRole, token.User.Role)
	assert.Equal("", token.RefreshToken)
	expectedSymbols := []string{"TSLA", "AAPL", "AMZN", "NFLX", "FB"}
	for i, stock := range token.User.Watchlists[0].Stocks {
		es := expectedSymbols[i]
		assert.Equal(es, stock.Symbol)
	}

	content, err := verifier.Verify(token.Token)
	assert.NoError(err)
	assert.Equal(auth.AnonymousRole, content.User.Role)
}

func TestHandleTokenRenewal(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	tokenID := id.New()

	oldSession := domain.Session{
		ID:           tokenID,
		UserID:       userID,
		Active:       true,
		RefreshToken: id.New(),
		CreatedAt:    time.Now().UTC().Add(-48 * time.Hour),
	}

	authUser := auth.User{
		ID:   userID,
		Role: auth.UserRole,
	}

	expectedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: "some@email.com",
			Role:  authUser.Role,
		},
	}

	userRepo := &repository.MockUserRepo{
		FindUser: expectedUser,
	}
	sessionRepo := &repository.MockSessionRepo{
		FindSession: oldSession,
	}

	cfg := getTestConfig()
	verifier := auth.NewVerifier(cfg.JWTCredentials, 0)
	signer := auth.NewSigner(cfg.JWTCredentials, 24*time.Hour)

	jwt, err := signer.Sign(tokenID, authUser)
	assert.NoError(err)
	oldToken := user.Token{
		Token:        jwt,
		RefreshToken: oldSession.RefreshToken,
	}
	mockEnv := getTestEnv(cfg, userRepo, sessionRepo, nil)

	// Setup: Renew token happy path.
	server := newServer(mockEnv, cfg)

	req := createTestPutRequest("", "", "/v1/login", oldToken)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)
	var newToken user.Token
	err = json.NewDecoder(res.Body).Decode(&newToken)
	assert.NoError(err)
	assert.NotEqual(oldToken.RefreshToken, newToken.RefreshToken)
	assert.Equal(userID, newToken.User.ID)
	newTokenContent, err := verifier.Verify(newToken.Token)
	assert.NoError(err)
	assert.Equal(userID, newTokenContent.User.ID)
	assert.NotEqual(tokenID, newTokenContent.ID)

	// Bad request part
	req = createTestPutRequest("", "", "/v1/login", user.User{})
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusBadRequest, res.Code)
}

func TestDisallowAnonymousToken(t *testing.T) {
	assert := assert.New(t)

	cfg := getTestConfig()
	signer := auth.NewSigner(cfg.JWTCredentials, 1*time.Hour)
	token, err := signer.Sign(id.New(), auth.User{ID: id.New(), Role: auth.AnonymousRole})
	assert.NoError(err)
	mockEnv := getTestEnv(cfg, nil, nil, nil)

	// Setup: Get anonymous token happy path.
	server := newServer(mockEnv, cfg)

	getRoutes := []string{
		"/v1/users/some-user-id",
		"/v1/watchlists/some-list-id",
	}
	for i, route := range getRoutes {
		name := fmt.Sprintf("%d - TestDisallowAnonymousToken GET %s", i+1, route)
		req := createTestGetRequest("", token, route)
		res := performTestRequest(server.Handler, req)
		assert.Equal(http.StatusForbidden, res.Code, name)
	}

	putRoutes := []string{
		"/v1/users/some-user-id/password",
		"/v1/users/some-user-id/email",
		"/v1/watchlists/some-list-id/name/list-name",
		"/v1/watchlists/some-list-id/stock/symbol-name",
	}
	for i, route := range putRoutes {
		name := fmt.Sprintf("%d - TestDisallowAnonymousToken PUT %s", i+1, route)
		req := createTestPutRequest("", token, route, nil)
		res := performTestRequest(server.Handler, req)
		assert.Equal(http.StatusForbidden, res.Code, name)
	}

	deleteRoutes := []string{
		"/v1/users/some-user-id",
		"/v1/watchlists/some-list-id",
		"/v1/watchlists/some-list-id/stock/symbol-name",
	}
	for i, route := range deleteRoutes {
		name := fmt.Sprintf("%d - TestDisallowAnonymousToken DELETE %s", i+1, route)
		req := createTestDeleteRequest("", token, route)
		res := performTestRequest(server.Handler, req)
		assert.Equal(http.StatusForbidden, res.Code, name)
	}

	postRoutes := []string{
		"/v1/watchlists/some-list-name",
	}
	for i, route := range postRoutes {
		name := fmt.Sprintf("%d - TestDisallowAnonymousToken POST %s", i+1, route)
		req := createTestPostRequest("", token, route, nil)
		res := performTestRequest(server.Handler, req)
		assert.Equal(http.StatusForbidden, res.Code, name)
	}

}
