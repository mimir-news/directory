package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
)

func main() {
	conf := getConfig()
	log.Println(conf)
	e := setupEnv(conf)
	defer e.close()
	server := newServer(e, conf)

	log.Printf("Starting %s on port: %s\n", ServiceName, conf.Port)
	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

func newServer(e *env, conf config) *http.Server {
	r := newRouter(e, conf)

	// Unsecured enpoints
	r.POST("/v1/users", e.handleUserCreation)
	r.POST("/v1/login", e.handleLogin)

	// Secured user routes
	r.GET("/v1/users/:userId", e.handleGetUser)
	r.PUT("/v1/users/:userId/password", e.handleChangePassword)
	r.PUT("/v1/users/:userId/email", e.handleChangeEmail)
	r.DELETE("/v1/users/:userId", e.handleDeleteUser)

	// Secured watchlist routes
	r.POST("/v1/watchlists/:name", e.handleCreateWatchlist)
	r.DELETE("/v1/watchlists/:watchlistId", e.handleDeleteWatchlist)
	r.GET("/v1/watchlists/:watchlistId", e.handleGetWatchlist)
	r.PUT("/v1/watchlists/:watchlistId/name/:name", e.handleRenameWatchlist)
	r.PUT("/v1/watchlists/:watchlistId/stock/:symbol", e.handleAddStockToWatchlist)
	r.DELETE("/v1/watchlists/:watchlistId/stock/:symbol", e.handleDeleteStockFromWatchlist)

	return &http.Server{
		Addr:    ":" + conf.Port,
		Handler: r,
	}
}

func newRouter(e *env, conf config) *gin.Engine {
	authOpts := auth.NewOptions(
		conf.TokenSecret, conf.TokenVerificationKey, conf.UnsecuredRoutes...)
	r := httputil.NewRouter(ServiceName, ServiceVersion, e.healthCheck)
	r.Use(auth.RequireToken(authOpts))

	return r
}

func (e *env) healthCheck() error {
	return e.db.Ping()
}

func placeholderHandler(c *gin.Context) {
	httputil.SendOK(c)
}
