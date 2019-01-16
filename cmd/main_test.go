package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
	"github.com/mimir-news/pkg/id"
)

var (
	expectedTestError = httputil.NewInternalServerError("expected error")
	correctPassword   = "super-secret-password"
	encryptedPassword = "S5UeZOWCDkIfP/5LUDpyhIY0l6+aow+CmkBEVtHqpebhe04vb6kDbPaD/wo05fs6x1lvJfI/6YZ66zbQ8X2lHaEThp4f1Zl0exk7j/wow740KbWZHf9DSA=="
	encryptedSalt     = "3MQEKd3NVnU+WQFQxo8JpYWrTrqXOiwro4MwLwnsckWXinE="
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

func createTestDeleteRequest(clientID, token, route string) *http.Request {
	return createTestRequest(clientID, token, route, http.MethodDelete, nil)
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

func getTestEnv(cfg config, userRepo repository.UserRepo,
	sessionRepo repository.SessionRepo, listRepo repository.WatchlistRepo) *env {

	passwordSvc := service.NewPasswordService(userRepo, cfg.PasswordPepper, cfg.PasswordEncryptionKey)
	tokenSigner := getTestSigner(cfg)
	verifier := auth.NewVerifier(cfg.JWTCredentials, 365*24*time.Hour)
	userSvc := service.NewUserService(passwordSvc, tokenSigner, verifier, userRepo, sessionRepo)
	listSvc := service.NewWatchlistService(listRepo)
	return &env{
		passwordSvc:  passwordSvc,
		watchlistSvc: listSvc,
		userSvc:      userSvc,
	}
}

func getTestSigner(cfg config) auth.Signer {
	return auth.NewSigner(cfg.JWTCredentials, 24*time.Hour)
}

func getTestToken(conf config, userID, clientID string) string {
	signer := getTestSigner(conf)

	token, err := signer.Sign(id.New(), auth.User{ID: userID, Role: auth.UserRole})
	if err != nil {
		log.Fatal(err)
	}

	return token
}

func getTestConfig() config {
	return config{
		PasswordPepper:        "my-pepper",
		PasswordEncryptionKey: "my-encryption-key",
		Port:                  "8080",
		UnsecuredRoutes:       unsecuredRoutes,
		JWTCredentials: auth.JWTCredentials{
			Issuer: "directory",
			Secret: "my-secret",
		},
	}
}

func testHandler(c *gin.Context) {
	httputil.SendOK(c)
}
