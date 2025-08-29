package middleware

import (
	"context"
	"net/http"

	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/repository"
	"github.com/MoSed3/otp-server/internal/token"
)

type AdminKey struct{}

// AdminAuthenticator holds dependencies for admin authentication middleware.
type AdminAuthenticator struct {
	adminRepo  repository.Admin
	jwtService *token.JWTService
}

// NewAdminAuthenticator creates a new AdminAuthenticator instance.
func NewAdminAuthenticator(adminRepo repository.Admin, jwtService *token.JWTService) *AdminAuthenticator {
	return &AdminAuthenticator{adminRepo: adminRepo, jwtService: jwtService}
}

func (a *AdminAuthenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.jwtService.ParseToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims.Audience != token.AudianceAdmin {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tx := GetTxFromRequest(r)
		admin, err := a.adminRepo.GetByID(tx, claims.ID)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		issueAtNum, err := claims.GetIssuedAt()
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		issueAt := issueAtNum.UTC()
		switch {
		case issueAt.Before(admin.CreatedAt):
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		case admin.PasswordResetAt.Valid && issueAt.Before(admin.PasswordResetAt.Time):
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AdminKey{}, admin)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (a *AdminAuthenticator) AuthorizeSudo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admin := GetAdminFromRequest(r)
		if admin == nil || (admin.Role > models.RoleSudoAdmin) {
			http.Error(w, "Forbidden: Insufficient privileges", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetAdminFromContext(ctx context.Context) *models.Admin {
	if admin, ok := ctx.Value(AdminKey{}).(*models.Admin); ok {
		return admin
	}
	return nil
}

func GetAdminFromRequest(r *http.Request) *models.Admin {
	return GetAdminFromContext(r.Context())
}
