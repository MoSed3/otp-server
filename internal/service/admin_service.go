package service

import (
	"context"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/MoSed3/otp-server/internal/middleware"
	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")

// AdminService defines the interface for admin-related business logic.
type AdminService interface {
	Login(ctx context.Context, username, password string) (*models.Admin, error)
	SearchUsers(tx *gorm.DB, params models.UserSearchParams) ([]models.User, int64, error)
	GetUserByID(tx *gorm.DB, userID uint) (*models.User, error)
	UpdateUserStatus(tx *gorm.DB, userID uint, status models.UserStatus) (*models.User, error)
}

// AdminServiceImpl implements AdminService.
type AdminServiceImpl struct {
	adminRepo repository.Admin
	userRepo  repository.User
}

// NewAdminService creates a new instance of AdminService.
func NewAdminService(adminRepo repository.Admin, userRepo repository.User) AdminService {
	return &AdminServiceImpl{
		adminRepo: adminRepo,
		userRepo:  userRepo,
	}
}

func (s *AdminServiceImpl) Login(ctx context.Context, username, password string) (*models.Admin, error) {
	log.Printf("Admin login attempt for username: %s", username)

	tx := middleware.GetTxFromContext(ctx)

	admin, err := s.adminRepo.GetByUsername(tx, username)
	if err != nil {
		log.Printf("Failed to get admin by username %s: %v", username, err)
		return nil, errors.New("invalid username or password")
	}
	if admin == nil {
		log.Printf("Admin not found for username: %s", username)
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.HashedPassword), []byte(password)); err != nil {
		log.Printf("Password mismatch for admin %s: %v", username, err)
		return nil, errors.New("invalid username or password")
	}

	log.Printf("Admin %s authenticated successfully", username)
	return admin, nil
}

func (s *AdminServiceImpl) SearchUsers(tx *gorm.DB, params models.UserSearchParams) ([]models.User, int64, error) {
	users, total, err := s.userRepo.Search(tx, params)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (s *AdminServiceImpl) GetUserByID(tx *gorm.DB, userID uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *AdminServiceImpl) UpdateUserStatus(tx *gorm.DB, userID uint, status models.UserStatus) (*models.User, error) {
	user, err := s.userRepo.GetByID(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.Status = status
	if err = s.userRepo.UpdateStatus(tx, user); err != nil {
		return nil, err
	}
	return user, nil
}
