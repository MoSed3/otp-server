package middleware

import (
	"context"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/MoSed3/otp-server/internal/db"
)

type TransactionKey struct{}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// Transaction wraps each request in a database transaction
// It automatically commits on successful responses (2xx status codes)
// and rolls back on errors (4xx, 5xx status codes) or panics
func Transaction(database *db.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start a new database transaction using the request context
			tx := database.GetTransaction(r.Context())
			if tx.Error != nil {
				log.Printf("Failed to begin transaction: %v", tx.Error)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Add transaction to request context
			ctx := context.WithValue(r.Context(), TransactionKey{}, tx)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: 0}

			// Handle panics and ensure proper transaction cleanup
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic occurred, rolling back transaction: %v", r)
					if err := tx.Rollback().Error; err != nil {
						log.Printf("Failed to rollback transaction after panic: %v", err)
					}
					// Re-panic to let other middleware handle it
					panic(r)
				}
			}()

			// Process the request
			next.ServeHTTP(rw, r)

			// Determine transaction outcome based on response status
			if rw.statusCode >= 200 && rw.statusCode < 300 {
				// Success: commit the transaction
				if err := tx.Commit().Error; err != nil {
					log.Printf("Failed to commit transaction: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				log.Printf("Transaction committed successfully for %s %s (status: %d)",
					r.Method, r.URL.Path, rw.statusCode)
			} else {
				// Error: rollback the transaction
				if err := tx.Rollback().Error; err != nil {
					log.Printf("Failed to rollback transaction: %v", err)
				} else {
					log.Printf("Transaction rolled back for %s %s (status: %d)",
						r.Method, r.URL.Path, rw.statusCode)
				}
			}
		})
	}
}

// GetTxFromContext retrieves the database transaction from the request context
// Returns nil if no transaction is found in the context
func GetTxFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(TransactionKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// GetTxFromRequest retrieves the database transaction from the HTTP request context
// Returns nil if no transaction is found in the context
func GetTxFromRequest(r *http.Request) *gorm.DB {
	return GetTxFromContext(r.Context())
}
