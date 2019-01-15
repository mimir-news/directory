package main

import (
	"encoding/json"
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
