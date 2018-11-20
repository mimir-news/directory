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
	userRepo    repository.UserRepo
	passwordSvc *service.PasswordService
	tokenSigner auth.Signer
	db          *sql.DB
}

func setupEnv(conf config) *env {
	db, err := conf.DB.ConnectPostgres()
	if err != nil {
		log.Fatal(err)
	}
	runMigrations(db)

	userRepo := repository.NewUserRepo(db)
	passwordSvc := service.NewPasswordService(
		userRepo, conf.PasswordPepper, conf.PasswordEncryptionKey)
	signer := auth.NewSigner(
		conf.TokenSecret, conf.TokenVerificationKey, 24*time.Hour)

	return &env{
		userRepo:    userRepo,
		passwordSvc: passwordSvc,
		tokenSigner: signer,
		db:          db,
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
