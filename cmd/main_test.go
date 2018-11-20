package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
)

var (
	expectedTestError = httputil.NewInternalServerError("expected error")
)

func performTestRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createTestPostRequest(clientID, token, route string, body interface{}) *http.Request {
	return createTestRequest(clientID, token, route, http.MethodPost, body)
}

func createTestPutRequest(clientID, token, route string, body interface{}) *http.Request {
	return createTestRequest(clientID, token, route, http.MethodPut, body)
}

func createTestGetRequest(clientID, token, route string) *http.Request {
	return createTestRequest(clientID, token, route, http.MethodGet, nil)
}

func createTestRequest(clientID, token, route, method string, body interface{}) *http.Request {
	var reqBody io.Reader
	if body != nil {
		bytesBody, err := json.Marshal(body)
		if err != nil {
			log.Fatal(err)
		}
		reqBody = bytes.NewBuffer(bytesBody)
	}

	req, err := http.NewRequest(method, route, reqBody)
	if err != nil {
		log.Fatal(err)
	}

	if clientID != "" {
		req.Header.Set(auth.ClientIDKey, clientID)
	}
	if token != "" {
		bearerToken := auth.AuthTokenPrefix + token
		req.Header.Set(auth.AuthHeaderKey, bearerToken)
	}

	return req
}

func getTestEnv(conf config, userRepo repository.UserRepo) *env {
	return &env{
		userRepo:    userRepo,
		passwordSvc: service.NewPasswordService(userRepo, conf.PasswordPepper, conf.PasswordEncryptionKey),
		tokenSigner: auth.NewSigner(conf.TokenSecret, conf.TokenVerificationKey, 24*time.Hour),
	}
}

func getTestConfig() config {
	return config{
		PasswordPepper:        "my-pepper",
		PasswordEncryptionKey: "my-encryption-key",
		TokenSecret:           "my-secret",
		TokenVerificationKey:  "my-verification-key",
		Port:                  "8080",
		UnsecuredRoutes: []string{
			"/health",
			"/v1/users",
		},
	}
}

type mockUserRepo struct {
	findUser domain.FullUser
	findErr  error
	findArg  string

	findByEmailUser domain.FullUser
	findByEmailErr  error
	findByEmailArg  string

	saveErr error
	saveArg domain.FullUser
}

func (r *mockUserRepo) Find(id string) (domain.FullUser, error) {
	r.findArg = id
	return r.findUser, r.findErr
}

func (r *mockUserRepo) FindByEmail(email string) (domain.FullUser, error) {
	r.findByEmailArg = email
	return r.findByEmailUser, r.findByEmailErr
}

func (r *mockUserRepo) Save(user domain.FullUser) error {
	r.saveArg = user
	return r.saveErr
}
