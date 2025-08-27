package controller

import (
	"context"
	"log"
	"net/http"

	"github.com/MoSed3/otp-server/db"
	"github.com/MoSed3/otp-server/middleware"
	"github.com/MoSed3/otp-server/redis"
)

type User struct {
	Controller
}

func NewUser(o Operator) *User {
	return &User{Controller: New(o)}
}

func (u *User) Login(ctx context.Context, r *http.Request, phoneNumber string) (string, error) {
	log.Printf("Login attempt for phone number: %s", phoneNumber)

	tx := middleware.GetTxFromRequest(r)

	user, err := db.GetOrCreateUserByPhoneNumber(tx, phoneNumber)
	if err != nil {
		log.Printf("Failed to get/create user for phone %s: %v", phoneNumber, err)
		return "", err
	}
	log.Printf("User retrieved/created successfully: UserID=%d, Phone=%s", user.ID, phoneNumber)

	otp, err := db.CreateOTP(tx, user)
	if err != nil {
		log.Printf("Failed to create OTP for user %d: %v", user.ID, err)
		return "", err
	}
	log.Printf("OTP created successfully: UserID=%d, OtpID=%d, Code=%s", user.ID, otp.ID, otp.Code)

	token, err := redis.CreateUserLoginSession(ctx, otp.ID, otp.Code)
	if err != nil {
		log.Printf("Failed to create login session for OtpID %d: %v", otp.ID, err)
		return "", err
	}
	log.Printf("Login session created successfully: UserID=%d, Token=%s", user.ID, token)

	return token, nil
}

func (u *User) VerifyOTP(ctx context.Context, r *http.Request, token, code string) (*db.User, error) {
	log.Printf("OTP verification attempt: Token=%s", token)

	tx := middleware.GetTxFromRequest(r)

	otpID, err := redis.CheckUserLoginCode(ctx, token, code)
	if err != nil {
		log.Printf("Invalid OTP verification attempt: Token=%s, Error=%v", token, err)
		return nil, err
	}
	log.Printf("OTP code verified successfully: Token=%s, OtpID=%d", token, otpID)

	otp, err := db.GetOtpByID(tx, otpID)
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

	user, err := db.GetUserByOtpID(tx, otpID)
	if err != nil {
		log.Printf("Failed to retrieve user by OTP: OtpID=%d, Error=%v", otpID, err)
		return nil, err
	}
	log.Printf("OTP verification completed successfully: UserID=%d, Phone=%s", user.ID, user.PhoneNumber)

	return user, nil
}
