package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/mimir-news/pkg/dbutil"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
)

func main() {
	conf := getConfig()
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

	disallowAnonymous := auth.DisallowRoles(auth.AnonymousRole)

	// Unsecured enpoints
	r.POST("/v1/users", e.handleUserCreation)
	r.POST("/v1/login", e.handleLogin)
	r.PUT("/v1/login", e.handleTokenRenewal)
	r.GET("/v1/login/anonymous", e.getAnonymousToken)

	// Secured user routes
	userGroup := r.Group("/v1/users", disallowAnonymous)
	userGroup.GET("/:userId", e.handleGetUser)
	userGroup.PUT("/:userId/password", e.handleChangePassword)
	userGroup.PUT("/:userId/email", e.handleChangeEmail)
	userGroup.DELETE("/:userId", e.handleDeleteUser)

	// Secured watchlist routes
	watchlistGroup := r.Group("/v1/watchlists", disallowAnonymous)
	watchlistGroup.POST("/:name", e.handleCreateWatchlist)
	watchlistGroup.DELETE("/:watchlistId", e.handleDeleteWatchlist)
	watchlistGroup.GET("/:watchlistId", e.handleGetWatchlist)
	watchlistGroup.PUT("/:watchlistId/name/:name", e.handleRenameWatchlist)
	watchlistGroup.PUT("/:watchlistId/stock/:symbol", e.handleAddStockToWatchlist)
	watchlistGroup.DELETE("/:watchlistId/stock/:symbol", e.handleDeleteStockFromWatchlist)

	return &http.Server{
		Addr:    ":" + conf.Port,
		Handler: r,
	}
}

func newRouter(e *env, cfg config) *gin.Engine {
	authOpts := auth.NewOptions(cfg.JWTCredentials, cfg.UnsecuredRoutes...)
	r := httputil.NewRouter(ServiceName, ServiceVersion, e.healthCheck)
	r.Use(auth.RequireToken(authOpts))

	return r
}

func (e *env) healthCheck() error {
	return dbutil.IsConnected(e.db)
}

func placeholderHandler(c *gin.Context) {
	httputil.SendOK(c)
}
