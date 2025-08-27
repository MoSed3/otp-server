package router

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/MoSed3/otp-server/docs"
	"github.com/MoSed3/otp-server/middleware"
)

const BasePath = "/api/v1"

func newRouter() chi.Router {
	r := chi.NewRouter()

	// Apply logging middleware globally
	r.Use(cMiddleware.Logger)
	r.Use(middleware.Transaction)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Auth routes
	r.Route(BasePath+"/auth", func(r chi.Router) {
		r.Use(middleware.RateLimit("auth", 5, 60, true))
		r.Post("/request-otp", requestOTP)
		r.Post("/verify-otp", verifyOTP)
	})

	// User routes
	r.Route(BasePath+"/user", func(r chi.Router) {
		r.Use(middleware.AuthenticateUser)
		r.Use(middleware.RateLimit("user", 30, 60, true))
		r.Get("/profile", getCurrentUser)
		r.Put("/profile", updateProfile)
	})

	return r
}

type Config struct {
	router   chi.Router
	address  string
	certFile string
	keyFile  string
}

func New(address, cert, key string) Config {
	return Config{
		router:   newRouter(),
		address:  address,
		certFile: cert,
		keyFile:  key,
	}
}

func (c *Config) Start() *http.Server {
	server := &http.Server{
		Addr:    c.address,
		Handler: c.router,
	}

	go func() {
		if c.certFile != "" && c.keyFile != "" {
			fmt.Printf("Starting HTTPS server on %s\n", c.address)
			log.Printf("HTTPS server listening on %s", c.address)
			if err := server.ListenAndServeTLS(c.certFile, c.keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTPS server failed to start: %v", err)
			}
		} else {
			fmt.Printf("Starting HTTP server on %s\n", c.address)
			log.Printf("HTTP server listening on %s", c.address)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server failed to start: %v", err)
			}
		}
	}()

	return server
}
