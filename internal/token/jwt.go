package token

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/MoSed3/otp-server/internal/setting"
)

// JWTService provides methods for JWT token generation and parsing.
type JWTService struct {
	appSettings *setting.Config
}

// NewJWTService creates a new JWTService instance.
func NewJWTService(appSettings *setting.Config) *JWTService {
	return &JWTService{
		appSettings: appSettings,
	}
}

type Claims struct {
	jwt.RegisteredClaims
	ID       uint     `json:"id"`
	Audience Audiance `json:"aud"`
}

func (s *JWTService) GenerateToken(id uint, audiance Audiance) (string, error) {
	expireDuration := s.appSettings.AccessTokenExpire()
	now := time.Now().UTC()

	claims := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireDuration) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Audience: audiance,
	}

	// Sign the token with the retrieved key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.appSettings.SecretKey())
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

func (s *JWTService) ParseToken(r *http.Request) (*Claims, error) {
	return s.ParseTokenFromHeader(r, "Authorization")
}

func (s *JWTService) ParseTokenFromHeader(r *http.Request, headerName string) (*Claims, error) {
	authHeader := r.Header.Get(headerName)
	if authHeader == "" {
		return nil, fmt.Errorf("no authorization header '%s'", headerName)
	}

	// Check if the Authorization header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("invalid authorization header format: missing Bearer prefix")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenBytes := s.appSettings.SecretKey()

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tokenBytes, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
