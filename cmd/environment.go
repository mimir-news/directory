package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/dbutil"
	"github.com/mimir-news/pkg/httputil/auth"
)

type env struct {
	passwordSvc  *service.PasswordService
	watchlistSvc service.WatchlistService
	userSvc      service.UserService
	db           *sql.DB
}

func setupEnv(conf config) *env {
	db, err := conf.DB.ConnectPostgres()
	if err != nil {
		log.Fatal(err)
	}
	runMigrations(db)

	userRepo := repository.NewUserRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	watchlsitRepo := repository.NewWatchlistRepo(db)

	passwordSvc := service.NewPasswordService(userRepo, conf.PasswordPepper, conf.PasswordEncryptionKey)
	signer := auth.NewSigner(conf.JWTCredentials, 24*time.Hour)
	verifier := auth.NewVerifier(conf.JWTCredentials, 365*24*time.Hour)

	userService := service.NewUserService(passwordSvc, signer, verifier, userRepo, sessionRepo)
	watchlistSvc := service.NewWatchlistService(watchlsitRepo)

	return &env{
		passwordSvc:  passwordSvc,
		watchlistSvc: watchlistSvc,
		userSvc:      userService,
		db:           db,
	}
}

func runMigrations(db *sql.DB) {
	err := dbutil.Migrate("./migrations", "postgres", db)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *env) close() {
	err := e.db.Close()
	if err != nil {
		log.Println(err)
	}
}
