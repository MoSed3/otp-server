// @title           OTP Server API
// @version         0.2.0
// @description     A REST API server for OTP-based authentication

// @contact.name   Mohammad Seddighi

// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuthUser
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token for user authentication.

// @securityDefinitions.apikey BearerAuthAdmin
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token for admin authentication.

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MoSed3/otp-server/internal/config"
	"github.com/MoSed3/otp-server/internal/db"
	"github.com/MoSed3/otp-server/internal/redis"
	"github.com/MoSed3/otp-server/internal/router"
	"github.com/MoSed3/otp-server/internal/setting"
	"github.com/MoSed3/otp-server/internal/token"
)

func main() {
	cfg := config.LoadConfig()
	database, err := db.Init(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Stop()

	appSettings := setting.New()
	appSettings.Init(database)

	jwtService := token.NewJWTService(appSettings)

	redisClient := redis.New(cfg.Redis)
	if err := redisClient.Start(); err != nil {
		log.Fatal("Failed to start redis, Error:", err)
	}
	defer redisClient.Stop()

	r := router.New(cfg.Server, database, redisClient, jwtService)
	server := r.Start()

	log.Println("Server started successfully")

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
