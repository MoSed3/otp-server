package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/MoSed3/otp-server/internal/models"
)

// User defines the interface for user data access operations.
type User interface {
	Create(tx *gorm.DB, number string) (*models.User, error)
	GetByPhoneNumber(tx *gorm.DB, phoneNumber string) (*models.User, error)
	Update(tx *gorm.DB, user *models.User, firstName, lastName string) error
	GetByID(tx *gorm.DB, id uint) (*models.User, error)
	GetOrCreateByPhoneNumber(tx *gorm.DB, phoneNumber string) (*models.User, error)
	Search(tx *gorm.DB, params models.UserSearchParams) ([]models.User, int64, error)
	UpdateStatus(tx *gorm.DB, user *models.User) error
}

// gormUser implements User using GORM.
type gormUser struct{}

// NewUser creates a new instance of gormUser.
func NewUser() User {
	return &gormUser{}
}

func (r *gormUser) Create(tx *gorm.DB, number string) (*models.User, error) {
	user := &models.User{PhoneNumber: number}
	return user, tx.Create(user).Error
}

func (r *gormUser) GetByPhoneNumber(tx *gorm.DB, phoneNumber string) (*models.User, error) {
	user := &models.User{PhoneNumber: phoneNumber}
	return user, tx.Where(user).Find(user).Error
}

func (r *gormUser) Update(tx *gorm.DB, user *models.User, firstName, lastName string) error {
	user.FirstName = firstName
	user.LastName = lastName
	return tx.Save(user).Error
}

func (r *gormUser) UpdateStatus(tx *gorm.DB, user *models.User) error {
	return tx.Save(user).Error
}

func (r *gormUser) GetByID(tx *gorm.DB, id uint) (*models.User, error) {
	user := &models.User{}
	return user, tx.First(user, id).Error
}

func (r *gormUser) Search(tx *gorm.DB, params models.UserSearchParams) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	db := tx.Model(&models.User{})

	if params.ID != nil {
		db = db.Where("id = ?", *params.ID)
	}
	if params.PhoneNumber != nil {
		db = db.Where("phone_number LIKE ?", "%"+*params.PhoneNumber+"%")
	}
	if params.FirstName != nil {
		db = db.Where("first_name LIKE ?", "%"+*params.FirstName+"%")
	}
	if params.LastName != nil {
		db = db.Where("last_name LIKE ?", "%"+*params.LastName+"%")
	}
	if params.Status != nil {
		db = db.Where("status = ?", *params.Status)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if params.SortBy == "" {
		params.SortBy = "id" // Default sort by ID
	}
	if params.SortOrder == "" {
		params.SortOrder = "asc" // Default sort order ascending
	}

	orderClause := fmt.Sprintf("%s %s", params.SortBy, params.SortOrder)
	err = db.Order(orderClause).Limit(params.Limit).Offset(params.Offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *gormUser) GetOrCreateByPhoneNumber(tx *gorm.DB, phoneNumber string) (*models.User, error) {
	user, err := r.GetByPhoneNumber(tx, phoneNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || user.ID == 0 {
			return r.Create(tx, phoneNumber)
		}
		return nil, err
	}
	return user, nil
}
