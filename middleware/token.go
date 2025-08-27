package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/MoSed3/otp-server/setting"
)

type Claims struct {
	jwt.RegisteredClaims
	ID uint `json:"id"`
}

func GenerateToken(id uint) (string, error) {
	expireDuration := setting.AccessTokenExpire()
	now := time.Now().UTC()

	claims := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireDuration) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Sign the token with the retrieved key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(setting.SecretKey())
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

func ParseToken(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("no authorization header")
	}
	tokenString := strings.Split(authHeader, " ")[1]
	tokenBytes := setting.SecretKey()

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tokenBytes, nil
	})
	switch {
	case err != nil, !token.Valid:
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
