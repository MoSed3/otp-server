package setting

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/MoSed3/otp-server/internal/db"
	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/repository"
)

// Config holds the application settings.
type Config struct {
	mutex             sync.RWMutex
	secretKey         []byte
	accessTokenExpire uint
}

// New creates and initializes a new settings configuration.
func New() *Config {
	return &Config{}
}

// Update updates the settings with new values.
func (c *Config) Update(s *models.Setting) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.accessTokenExpire = s.AccessTokenExpire
	c.secretKey = []byte(s.SecretKey)
}

// Init initializes the settings by loading them from the database.
func (c *Config) Init(database *db.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx := database.GetTransaction(ctx)
	settingRepo := repository.NewSetting()

	s, err := settingRepo.Get(tx)
	if err != nil {
		log.Fatalf("Error while getting settings from db: %v", err)
	}

	c.Update(s)
}

// SecretKey returns the application's secret key.
func (c *Config) SecretKey() []byte {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.secretKey
}

// AccessTokenExpire returns the access token expiration time in minutes.
func (c *Config) AccessTokenExpire() uint {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.accessTokenExpire
}
