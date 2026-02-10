package middleware

import "net/http"

type RateLimiterMiddleware struct {
}

func NewRateLimiterMiddleware() *RateLimiterMiddleware {
	return &RateLimiterMiddleware{}
}

func (m *RateLimiterMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
