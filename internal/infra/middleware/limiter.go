package middleware

import (
	"net"
	"net/http"

	"github.com/brecabral/rate-limiter/internal/infra/limiter"
)

type RateLimiterMiddleware struct {
	RateLimiter *limiter.RateLimiter
}

func NewRateLimiterMiddleware(RateLimiter *limiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		RateLimiter: RateLimiter,
	}
}

func (m *RateLimiterMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("API_KEY")

		if token == "" {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip != "" && m.RateLimiter.AllowIP(ip) {
				next.ServeHTTP(w, r)
				return
			}
		}

		if m.RateLimiter.AllowToken(token) {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
	})
}
