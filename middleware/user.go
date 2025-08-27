package middleware

import (
	"context"
	"net/http"

	"github.com/MoSed3/otp-server/db"
)

type UserKey struct{}

func AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := ParseToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tx := GetTxFromRequest(r)
		user, err := db.GetUserByID(tx, claims.ID)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey{}, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(ctx context.Context) *db.User {
	if user, ok := ctx.Value(UserKey{}).(*db.User); ok {
		return user
	}
	return nil
}

func GetUserFromRequest(r *http.Request) *db.User {
	return GetUserFromContext(r.Context())
}
