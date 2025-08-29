package router

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/MoSed3/otp-server/docs" // Keep this as is
	"github.com/MoSed3/otp-server/internal/config"
	"github.com/MoSed3/otp-server/internal/db"
	"github.com/MoSed3/otp-server/internal/middleware"
	"github.com/MoSed3/otp-server/internal/redis"
	"github.com/MoSed3/otp-server/internal/repository"
	"github.com/MoSed3/otp-server/internal/service"
	"github.com/MoSed3/otp-server/internal/token"
)

const BasePath = "/api/v1"

func newRouter(database *db.DB, redisCli *redis.Config, jwtService *token.JWTService) chi.Router {
	r := chi.NewRouter()

	// Initialize repositories
	userRepo := repository.NewUser()
	otpRepo := repository.NewOtp()
	adminRepo := repository.NewAdmin()

	// Initialize services
	userService := service.NewUserService(userRepo, otpRepo, redisCli)
	adminService := service.NewAdminService(adminRepo, userRepo)

	// Initialize handlers (controllers)
	userHandler := NewUserHandler(userService, jwtService)
	adminHandler := NewAdminHandler(adminService, jwtService)

	// Initialize middleware components
	userAuthenticator := middleware.NewAuthenticator(userRepo, jwtService)
	adminAuthenticator := middleware.NewAdminAuthenticator(adminRepo, jwtService)
	rateLimiter := middleware.NewRateLimiter(redisCli)

	// Apply logging middleware globally
	r.Use(cMiddleware.Logger)
	r.Use(middleware.Transaction(database))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Auth routes
	r.Route(BasePath+"/auth", func(r chi.Router) {
		r.Use(rateLimiter.RateLimit("auth", 5, 60, true))
		r.Post("/request-otp", userHandler.requestOTP)
		r.Post("/verify-otp", userHandler.verifyOTP)
		r.Post("/admin", adminHandler.adminLogin)
	})

	// User routes
	r.Route(BasePath+"/user/profile", func(r chi.Router) {
		r.Use(userAuthenticator.Authenticate)
		r.Use(rateLimiter.RateLimit("user", 30, 60, true))
		r.Get("/", userHandler.getCurrentUser)
		r.Put("/", userHandler.updateProfile)
	})

	// Admin routes
	r.Route(BasePath+"/admin", func(r chi.Router) {
		r.Use(adminAuthenticator.Authenticate)
		r.Use(rateLimiter.RateLimit("admin", 60, 60, true))
		r.Get("/profile", adminHandler.getCurrentAdmin)
		r.Get("/users", adminHandler.searchUsers)
		r.Get("/user/{id}", adminHandler.getUserByID)
		r.Patch("/user/{id}/status", http.HandlerFunc(adminAuthenticator.AuthorizeSudo(http.HandlerFunc(adminHandler.updateUserStatus)).ServeHTTP))
	})

	return r
}

type Config struct {
	router       chi.Router
	serverConfig config.ServerConfig
}

func New(serverConfig config.ServerConfig, database *db.DB, redisCli *redis.Config, jwtService *token.JWTService) Config {
	return Config{
		router:       newRouter(database, redisCli, jwtService),
		serverConfig: serverConfig,
	}
}

func (c *Config) Start() *http.Server {
	server := &http.Server{
		Addr:    net.JoinHostPort(c.serverConfig.WebAppHost, fmt.Sprintf("%d", c.serverConfig.WebAppPort)),
		Handler: c.router,
	}

	go func() {
		if c.serverConfig.CertFile != "" && c.serverConfig.KeyFile != "" {
			fmt.Printf("Starting HTTPS server on %s:%d\n", c.serverConfig.WebAppHost, c.serverConfig.WebAppPort)
			log.Printf("HTTPS server listening on %s:%d", c.serverConfig.WebAppHost, c.serverConfig.WebAppPort)
			if err := server.ListenAndServeTLS(c.serverConfig.CertFile, c.serverConfig.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTPS server failed to start: %v", err)
			}
		} else {
			fmt.Printf("Starting HTTP server on %s:%d\n", c.serverConfig.WebAppHost, c.serverConfig.WebAppPort)
			log.Printf("HTTP server listening on %s:%d", c.serverConfig.WebAppHost, c.serverConfig.WebAppPort)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server failed to start: %v", err)
			}
		}
	}()

	return server
}
