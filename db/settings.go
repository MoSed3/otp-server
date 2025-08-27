package db

import (
	"encoding/base64"
	"math/rand/v2"

	"gorm.io/gorm"
)

type Setting struct {
	ID                uint   `gorm:"primaryKey"`
	SecretKey         string `gorm:"not null"`
	AccessTokenExpire uint   `gorm:"not null"` // Minutes
}

// CreateSetting creates initial settings in the database
func CreateSetting(tx *gorm.DB) error {
	randomBytes := make([]byte, 256)
	c := &rand.ChaCha8{}
	_, _ = c.Read(randomBytes)
	randomToken := base64.StdEncoding.EncodeToString(randomBytes)

	newSetting := Setting{
		SecretKey:         randomToken,
		AccessTokenExpire: 1440,
	}

	return tx.Create(&newSetting).Error
}

// UpdateSetting updates existing settings
func (s *Setting) UpdateSetting(tx *gorm.DB) error {
	return tx.Save(s).Error
}

// GetSetting retrieves the settings (assuming single row)
func GetSetting(tx *gorm.DB) (*Setting, error) {
	var setting Setting
	if err := tx.First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}
