package middleware

import (
	"context"
	"net/http"

	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/repository"
	"github.com/MoSed3/otp-server/internal/token"
)

type UserKey struct{}

// Authenticator holds dependencies for user authentication middleware.
type Authenticator struct {
	userRepo   repository.User
	jwtService *token.JWTService
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator(userRepo repository.User, jwtService *token.JWTService) *Authenticator {
	return &Authenticator{userRepo: userRepo, jwtService: jwtService}
}

func (a *Authenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.jwtService.ParseToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims.Audience != token.AudianceUser {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tx := GetTxFromRequest(r)
		user, err := a.userRepo.GetByID(tx, claims.ID)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if user.Status == models.UserStatusDisabled {
			http.Error(w, "Forbidden: User is disabled", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey{}, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(UserKey{}).(*models.User); ok {
		return user
	}
	return nil
}

func GetUserFromRequest(r *http.Request) *models.User {
	return GetUserFromContext(r.Context())
}
