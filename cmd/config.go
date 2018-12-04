package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mimir-news/pkg/dbutil"
)

// Service metadata.
const (
	ServiceName    = "directory"
	ServiceVersion = "1.0"
)

var unsecuredRoutes = []string{
	"/health",
	"/v1/users",
	"/v1/login",
}

type config struct {
	DB                    dbutil.Config
	Port                  string
	PasswordPepper        string
	PasswordEncryptionKey string
	TokenSecret           string
	TokenVerificationKey  string
	UnsecuredRoutes       []string
}

func (c config) String() string {
	dsn := c.DB.PgDSN()
	routes := strings.Join(c.UnsecuredRoutes, ", ")
	return fmt.Sprintf(
		"config(db=[%s] port=%s pepper=%s encryptionKey=%s secret=%s tokenKey=%s routes=[%s])",
		dsn, c.Port, c.PasswordPepper, c.PasswordEncryptionKey,
		c.TokenSecret, c.TokenVerificationKey, routes)
}

func getConfig() config {
	passwordSecret := getSecret(mustGetenv("PASSWORD_SECRETS_FILE"))
	tokenSecret := getSecret(mustGetenv("TOKEN_SECRETS_FILE"))

	return config{
		DB:                    dbutil.MustGetConfig("DB"),
		Port:                  mustGetenv("SERVICE_PORT"),
		PasswordPepper:        passwordSecret.Secret,
		PasswordEncryptionKey: passwordSecret.Key,
		TokenSecret:           tokenSecret.Secret,
		TokenVerificationKey:  tokenSecret.Key,
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

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("No value for key: %s\n", key)
	}

	return val
}
