package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/mimir-news/pkg/dbutil"
	"github.com/mimir-news/pkg/httputil/auth"
)

// Service metadata.
const (
	ServiceName    = "directory"
	ServiceVersion = "1.2"
)

var unsecuredRoutes = []string{
	"/health",
	"/v1/users",
	"/v1/login",
	"/v1/login/anonymous",
}

type config struct {
	DB                    dbutil.Config
	Port                  string
	PasswordPepper        string
	PasswordEncryptionKey string
	JWTCredentials        auth.JWTCredentials
	UnsecuredRoutes       []string
}

func getConfig() config {
	passwordSecret := getSecret(mustGetenv("PASSWORD_SECRETS_FILE"))
	jwtCredentials := getJWTCredentials(mustGetenv("JWT_CREDENTIALS_FILE"))

	return config{
		DB:                    dbutil.MustGetConfig("DB"),
		Port:                  mustGetenv("SERVICE_PORT"),
		PasswordPepper:        passwordSecret.Secret,
		PasswordEncryptionKey: passwordSecret.Key,
		JWTCredentials:        jwtCredentials,
		UnsecuredRoutes:       unsecuredRoutes,
	}
}

type secret struct {
	Secret string `json:"secret"`
	Key    string `json:"key"`
}

func getSecret(filename string) secret {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	var s secret
	err = json.Unmarshal(content, &s)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func getJWTCredentials(filename string) auth.JWTCredentials {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	credentials, err := auth.ReadJWTCredentials(f)
	if err != nil {
		log.Fatal(err)
	}

	return credentials
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("No value for key: %s\n", key)
	}

	return val
}
