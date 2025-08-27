package middleware

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/MoSed3/otp-server/redis"
)

func getClientIP(r *http.Request, allowForwarded bool) string {
	if allowForwarded {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if ip := net.ParseIP(xff); ip != nil {
				return ip.String()
			}
		}

		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			if ip := net.ParseIP(xri); ip != nil {
				return ip.String()
			}
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func RateLimit(prefix string, maxRequests int, windowSeconds int, allowForwarded bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r, allowForwarded)
			key := redis.GetRateLimitKey(prefix, clientIP)

			allowed, remaining, err := redis.CheckRateLimit(r.Context(), key, maxRequests, windowSeconds)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(windowSeconds)*time.Second).Unix(), 10))

			if !allowed {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
