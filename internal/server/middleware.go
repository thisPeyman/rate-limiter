package server

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/thisPeyman/rate-limiter/internal/limiter"
)

type RateLimitMiddleware struct {
	limiter limiter.RateLimiter
}

func NewRateLimitMiddleware(l limiter.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: l}
}

func (m *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}
			userID = ip
		}

		limit := 10

		allowed, err := m.limiter.Allow(r.Context(), userID, limit)
		if err != nil {
			log.Printf("Rate Limit Error: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Too Many Requests"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
