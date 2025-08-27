package db

import (
	"database/sql"
	"errors"
	"math/rand/v2"
	"time"

	"gorm.io/gorm"
)

type UserOtp struct {
	gorm.Model
	Code   string       `gorm:"not null"`
	UserID uint         `gorm:"not null"`
	User   *User        `gorm:"foreignkey:UserID;not null"`
	UsedAt sql.NullTime `gorm:""`
}

func (o UserOtp) Waste(tx *gorm.DB) error {
	o.UsedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}

	return tx.Save(o).Error
}

func GetOtpByID(tx *gorm.DB, id uint) (*UserOtp, error) {
	otp := &UserOtp{Model: gorm.Model{ID: id}}
	return otp, tx.Where(otp).Find(otp).Error
}

func generateOTPCode() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.IntN(len(chars))]
	}
	return string(code)
}

func CreateOTP(tx *gorm.DB, user *User) (*UserOtp, error) {
	now := time.Now().UTC()
	twoMinutesAgo := now.Add(-2 * time.Minute)
	oneHourAgo := now.Add(-1 * time.Hour)

	var recentOtp UserOtp
	err := tx.Where("user_id = ? AND created_at > ? AND used_at IS NULL", user.ID, twoMinutesAgo).First(&recentOtp).Error

	switch {
	case err == nil:
		return nil, errors.New("user has generated OTP within the last 2 minutes")
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	var otpCount int64
	err = tx.Model(&UserOtp{}).Where("user_id = ? AND created_at > ?", user.ID, oneHourAgo).Count(&otpCount).Error

	switch {
	case err != nil:
		return nil, err
	case otpCount >= 5:
		return nil, errors.New("user has exceeded maximum OTP limit (5) within the last hour")
	}

	otp := &UserOtp{
		Code: generateOTPCode(),
		User: user,
	}

	err = tx.Create(otp).Error
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func GetUserByOtpID(tx *gorm.DB, otpID uint) (*User, error) {
	otp, err := GetOtpByID(tx, otpID)
	if err != nil {
		return nil, err
	}

	user := &User{}
	if err := tx.First(user, otp.UserID).Error; err != nil {
		return nil, err
	}

	return user, nil
}
