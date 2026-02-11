package middleware

import (
	"net"
	"net/http"
	"time"
)

type RateLimiterMiddleware struct {
	limitByIP int
	interval  time.Duration
	storeIP   map[string][]time.Time
}

func NewRateLimiterMiddleware(limitByIP int, interval time.Duration) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limitByIP: limitByIP,
		interval:  interval,
		storeIP:   make(map[string][]time.Time),
	}
}

func (m *RateLimiterMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if m.Allow(ip) {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
	})
}

func (m *RateLimiterMiddleware) Allow(key string) bool {
	timestamp := time.Now()
	requests := 0
	newStoreIP := []time.Time{}

	for _, t := range m.storeIP[key] {
		if timestamp.Sub(t) < m.interval {
			newStoreIP = append(newStoreIP, t)
			requests++
		}
	}
	m.storeIP[key] = newStoreIP

	if requests >= m.limitByIP {
		return false
	}

	m.storeIP[key] = append(m.storeIP[key], timestamp)
	return true
}
