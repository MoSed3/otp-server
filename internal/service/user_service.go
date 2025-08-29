package service

import (
	"context"
	"errors"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/MoSed3/otp-server/internal/middleware"
	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/redis"
	"github.com/MoSed3/otp-server/internal/repository"
)

var (
	ErrUserDisabled = errors.New("user is disabled")
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	Login(ctx context.Context, r *http.Request, phoneNumber string) (string, error)
	VerifyOTP(ctx context.Context, r *http.Request, token, code string) (*models.User, error)
	UpdateProfile(tx *gorm.DB, user *models.User, firstName, lastName string) error
	GetUserByID(tx *gorm.DB, id uint) (*models.User, error)
}

// UserServiceImpl implements UserService.
type UserServiceImpl struct {
	userRepo repository.User
	otpRepo  repository.Otp
	redisCli *redis.Config
}

// NewUserService creates a new instance of UserServiceImpl.
func NewUserService(userRepo repository.User, otpRepo repository.Otp, redisCli *redis.Config) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
		otpRepo:  otpRepo,
		redisCli: redisCli,
	}
}

func (s *UserServiceImpl) Login(ctx context.Context, r *http.Request, phoneNumber string) (string, error) {
	log.Printf("Login attempt for phone number: %s", phoneNumber)

	tx := middleware.GetTxFromRequest(r)

	user, err := s.userRepo.GetOrCreateByPhoneNumber(tx, phoneNumber)
	if err != nil {
		log.Printf("Failed to get/create user for phone %s: %v", phoneNumber, err)
		return "", err
	}
	log.Printf("User retrieved/created successfully: UserID=%d, Phone=%s", user.ID, phoneNumber)

	if user.Status == models.UserStatusDisabled {
		return "", ErrUserDisabled
	}

	otp, err := s.otpRepo.Create(tx, user)
	if err != nil {
		log.Printf("Failed to create OTP for user %d: %v", user.ID, err)
		return "", err
	}
	log.Printf("OTP created successfully: UserID=%d, OtpID=%d, Code=%s", user.ID, otp.ID, otp.Code)

	token, err := s.redisCli.CreateUserLoginSession(ctx, otp.ID, otp.Code)
	if err != nil {
		log.Printf("Failed to create login session for OtpID %d: %v", otp.ID, err)
		return "", err
	}
	log.Printf("Login session created successfully: UserID=%d, Token=%s", user.ID, token)

	return token, nil
}

func (s *UserServiceImpl) VerifyOTP(ctx context.Context, r *http.Request, token, code string) (*models.User, error) {
	log.Printf("OTP verification attempt: Token=%s", token)

	tx := middleware.GetTxFromRequest(r)

	otpID, err := s.redisCli.CheckUserLoginCode(ctx, token, code)
	if err != nil {
		log.Printf("Invalid OTP verification attempt: Token=%s, Error=%v", token, err)
		return nil, err
	}
	log.Printf("OTP code verified successfully: Token=%s, OtpID=%d", token, otpID)

	otp, err := s.otpRepo.GetByID(tx, otpID)
	if err != nil {
		log.Printf("Failed to retrieve OTP from database: OtpID=%d, Error=%v", otpID, err)
		return nil, err
	}
	log.Printf("OTP retrieved from database: OtpID=%d, UserID=%d", otp.ID, otp.UserID)

	if err = otp.Waste(tx); err != nil {
		log.Printf("Failed to mark OTP as used: OtpID=%d, Error=%v", otp.ID, err)
		return nil, err
	}
	log.Printf("OTP marked as used successfully: OtpID=%d", otp.ID)

	user, err := s.otpRepo.GetUserByOtpID(tx, otpID)
	if err != nil {
		log.Printf("Failed to retrieve user by OTP: OtpID=%d, Error=%v", otpID, err)
		return nil, err
	}

	if user.Status == models.UserStatusDisabled {
		return nil, ErrUserDisabled
	}

	log.Printf("OTP verification completed successfully: UserID=%d, Phone=%s", user.ID, user.PhoneNumber)

	return user, nil
}

func (s *UserServiceImpl) UpdateProfile(tx *gorm.DB, user *models.User, firstName, lastName string) error {
	return s.userRepo.Update(tx, user, firstName, lastName)
}

func (s *UserServiceImpl) GetUserByID(tx *gorm.DB, id uint) (*models.User, error) {
	return s.userRepo.GetByID(tx, id)
}
