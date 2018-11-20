package main

import (
	"net/http"
	"github.com/mimir-news/pkg/httputil"
	"github.com/gin-gonic/gin"
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
	authOpts := auth.NewOptions(
		conf.TokenSecret, conf.TokenVerificationKey, ...conf.UnsecuredRoutes)
	
	r := httputil.NewRouter(ServiceName, ServiceVersion, e.healthCheck)
	r.Use(auth.RequireToken(authOpts))
	
	return &http.Server {
		Addr: ":"+conf.Port,
		Handler: r,
	}
}

func (e *env) heathCheck() error {
	return e.db.Ping()
}
