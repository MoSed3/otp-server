package db

import (
	"context"
	"errors" // Added for errors.Is
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/MoSed3/otp-server/config"
)

var db *gorm.DB

func Init(autogenerate bool) {
	// Configure logger based on EchoSQLQueries setting
	var logMode logger.LogLevel
	if config.EchoSQLQueries {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}

	dbConn, err := gorm.Open(postgres.Open(config.DbUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	fmt.Println("Connected to PostgresSQL database successfully")

	// Configure connection pool
	sqlDB, err := dbConn.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxOpenConns(config.DbPoolSize + config.DbMaxOverflow)
	sqlDB.SetMaxIdleConns(config.DbPoolSize)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db = dbConn

	if autogenerate {
		autoMigrateAndSeedSettings(db)
	}
}

func autoMigrateAndSeedSettings(db *gorm.DB) {
	err := db.AutoMigrate(&User{}, &UserOtp{}, &Setting{})
	if err != nil {
		log.Fatalf("failed to auto migrate database: %v", err)
	}
	fmt.Println("Database auto-migrated successfully")

	// Check if setting table is empty and create default settings if it is
	_, err = GetSetting(db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := CreateSetting(db); err != nil {
				log.Fatalf("failed to create default settings: %v", err)
			}
			fmt.Println("Default settings created successfully")
		} else {
			log.Fatalf("failed to get settings: %v", err)
		}
	}
}

func GetTransaction(ctx context.Context) *gorm.DB {
	return db.Begin().WithContext(ctx)
}
