package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	envConf "github.com/caarlos0/env"
	"github.com/mimir-news/pkg/dbutil"
)

// Service metadata.
const (
	ServiceName    = "directory"
	ServiceVersion = "0.1"
)

var unsecuredRoutes = []string{
	"/health",
	"/v1/users",
	"/v1/login",
}

type config struct {
	DB                    dbutil.Config `env:"DB"`
	Port                  string        `env:"SERVICE_PORT"`
	PasswordPepper        string
	PasswordEncryptionKey string
	TokenSecret           string
	TokenVerificationKey  string
	UnsecuredRoutes       []string
}

func getConfig() config {
	var conf config
	err := envConf.Parse(&conf)
	if err != nil {
		log.Fatal(err)
	}

	passwordSecret := getSecret(mustGetenv("PASSWORD_SECRETS_FILE"))
	tokenSecret := getSecret(mustGetenv("TOKEN_SECRETS_FILE"))

	conf.PasswordPepper = passwordSecret.Secret
	conf.PasswordEncryptionKey = passwordSecret.Key
	conf.TokenSecret = tokenSecret.Secret
	conf.TokenVerificationKey = tokenSecret.Key
	conf.UnsecuredRoutes = unsecuredRoutes

	return conf
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
