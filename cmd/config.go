package main

import (
	"github.com/mimir-news/pkg/dbutil"
)

// Service metadata.
const (
	ServiceName    = "directory"
	ServiceVersion = "0.1"
)

type config struct {
	DB                    dbutil.Config `env:"DB"`
	PasswordPepper        string        `env:"PASSWORD_PEPPER"`
	PasswordEncryptionKey string        `env:"PASSWORD_ENCRYPTION_KEY"`
	TokenSecret           string        `env:"TOKEN_SECRET"`
	TokenVerificationKey  string        `env:"TOKEN_VERIFICATION_KEY"`
	Port                  string        `env:"PORT"`
	UnsecuredRoutes       []string
}
