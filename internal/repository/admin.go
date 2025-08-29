package repository

import (
	"errors"
	"time"

	"github.com/MoSed3/otp-server/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminUpdate struct for partial updates
type AdminUpdate struct {
	Username    *string
	Role        *models.AdminRole
	NewPassword *string
}

// Admin defines the interface for admin user management.
type Admin interface {
	Create(tx *gorm.DB, username, password string, role models.AdminRole) (*models.Admin, error)
	Update(tx *gorm.DB, adminID uint, updates AdminUpdate) error
	Delete(tx *gorm.DB, adminID uint) error
	GetByUsername(tx *gorm.DB, username string) (*models.Admin, error)
	GetByID(tx *gorm.DB, adminID uint) (*models.Admin, error)
	ListAll(tx *gorm.DB) ([]models.Admin, error)
}

// gormAdmin implements Admin.
type gormAdmin struct{}

// NewAdmin creates a new instance of gormAdmin.
func NewAdmin() Admin {
	return &gormAdmin{}
}

func (r *gormAdmin) Create(tx *gorm.DB, username, password string, role models.AdminRole) (*models.Admin, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := &models.Admin{
		Username:       username,
		Role:           role,
		HashedPassword: string(hashedPassword),
	}

	if err = tx.Create(admin).Error; err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *gormAdmin) Update(tx *gorm.DB, adminID uint, updates AdminUpdate) error {
	admin := &models.Admin{}
	if err := tx.First(admin, adminID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("admin user not found")
		}
		return err
	}

	if updates.Username != nil {
		admin.Username = *updates.Username
	}
	if updates.Role != nil {
		admin.Role = *updates.Role
	}
	if updates.NewPassword != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updates.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin.HashedPassword = string(hashedPassword)
		admin.PasswordResetAt.Time = time.Now().UTC()
		admin.PasswordResetAt.Valid = true
	}

	if err := tx.Save(admin).Error; err != nil {
		return err
	}
	return nil
}

func (r *gormAdmin) Delete(tx *gorm.DB, adminID uint) error {
	result := tx.Where("id = ?", adminID).Delete(&models.Admin{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("admin user not found")
	}
	return nil
}

func (r *gormAdmin) GetByUsername(tx *gorm.DB, username string) (*models.Admin, error) {
	var admin models.Admin
	if err := tx.Where("username = ?", username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Admin not found
		}
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdmin) GetByID(tx *gorm.DB, adminID uint) (*models.Admin, error) {
	var admin models.Admin
	if err := tx.Where("id = ?", adminID).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Admin not found
		}
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdmin) ListAll(tx *gorm.DB) ([]models.Admin, error) {
	var admins []models.Admin
	if err := tx.Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}
