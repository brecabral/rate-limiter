package middleware

import (
	"context"
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
		ctx := context.Background()
		key := r.Header.Get("API_KEY")
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		if m.RateLimiter.Allow(ctx, ip, key) {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
	})
}
