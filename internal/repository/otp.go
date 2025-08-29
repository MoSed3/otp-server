package repository

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"gorm.io/gorm"

	"github.com/MoSed3/otp-server/internal/models"
)

// Otp defines the interface for OTP data access operations.
type Otp interface {
	GetByID(tx *gorm.DB, id uint) (*models.UserOtp, error)
	Create(tx *gorm.DB, user *models.User) (*models.UserOtp, error)
	GetUserByOtpID(tx *gorm.DB, otpID uint) (*models.User, error)
}

// gormOtp implements Otp using GORM.
type gormOtp struct{}

// NewOtp creates a new instance of gormOtp.
func NewOtp() Otp {
	return &gormOtp{}
}

func (r *gormOtp) GetByID(tx *gorm.DB, id uint) (*models.UserOtp, error) {
	otp := &models.UserOtp{Model: gorm.Model{ID: id}}
	return otp, tx.Where(otp).Find(otp).Error
}

func generateOTPCode() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	return string(code)
}

func (r *gormOtp) Create(tx *gorm.DB, user *models.User) (*models.UserOtp, error) {
	now := time.Now().UTC()
	twoMinutesAgo := now.Add(-2 * time.Minute)
	tenMinutesAgo := now.Add(-10 * time.Minute)

	var recentOtp models.UserOtp
	err := tx.Where("user_id = ? AND created_at > ? AND used_at IS NULL", user.ID, twoMinutesAgo).First(&recentOtp).Error

	switch {
	case err == nil:
		return nil, errors.New("user has generated OTP within the last 2 minutes")
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	var otpCount int64
	err = tx.Model(&models.UserOtp{}).Where("user_id = ? AND created_at > ?", user.ID, tenMinutesAgo).Count(&otpCount).Error

	switch {
	case err != nil:
		return nil, err
	case otpCount >= 3:
		return nil, errors.New("user has exceeded maximum OTP limit (5) within the last hour")
	}

	otp := &models.UserOtp{
		Code: generateOTPCode(),
		User: user,
	}

	err = tx.Create(otp).Error
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (r *gormOtp) GetUserByOtpID(tx *gorm.DB, otpID uint) (*models.User, error) {
	otp, err := r.GetByID(tx, otpID)
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	if err = tx.First(user, otp.UserID).Error; err != nil {
		return nil, err
	}

	return user, nil
}
