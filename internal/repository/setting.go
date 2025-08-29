package repository

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"gorm.io/gorm"

	"github.com/MoSed3/otp-server/internal/models"
)

// Setting defines the interface for setting data access operations.
type Setting interface {
	Create(tx *gorm.DB) error
	Update(tx *gorm.DB, setting *models.Setting) error
	Get(tx *gorm.DB) (*models.Setting, error)
}

// gormSetting implements Setting using GORM.
type gormSetting struct{}

// NewSetting creates a new instance of gormSetting.
func NewSetting() Setting {
	return &gormSetting{}
}

// Create creates initial settings in the database
func (r *gormSetting) Create(tx *gorm.DB) error {
	randomBytes := make([]byte, 256)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomToken := base64.StdEncoding.EncodeToString(randomBytes)

	newSetting := models.Setting{
		SecretKey:         randomToken,
		AccessTokenExpire: 1440,
	}

	return tx.Create(&newSetting).Error
}

// Update updates existing settings
func (r *gormSetting) Update(tx *gorm.DB, s *models.Setting) error {
	return tx.Save(s).Error
}

// Get retrieves the settings (assuming single row)
func (r *gormSetting) Get(tx *gorm.DB) (*models.Setting, error) {
	var setting models.Setting
	if err := tx.First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}
