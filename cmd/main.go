package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
)

func main() {
	conf := config{}
	e := setupEnv(conf)
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
	r.PUT("/v1/password/reset", placeholderHandler)

	// Secured user routes
	r.GET("/v1/users/:userId", e.handleGetUser)
	r.PUT("/v1/users/:userId/password", placeholderHandler)
	r.PUT("/v1/users/:userId/email", placeholderHandler)
	r.DELETE("/v1/users/:userId", placeholderHandler)
	r.GET("/v1/users/:userId/watchlists", placeholderHandler)

	// Secured watchlist routes
	r.POST("/v1/watchlists/:name", placeholderHandler)
	r.GET("/v1/watchlists/:watchlistId", placeholderHandler)
	r.PUT("/v1/watchlists/:watchlistId", placeholderHandler)
	r.PUT("/v1/watchlists/:watchlistId/stock", placeholderHandler)
	r.DELETE("/v1/watchlists/:watchlistId/stock/:symbol", placeholderHandler)

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
