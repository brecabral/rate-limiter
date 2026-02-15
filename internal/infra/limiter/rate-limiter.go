package limiter

import "github.com/brecabral/rate-limiter/internal/infra/repository"

type RateLimiter struct {
	repo          repository.StoreKey
	maxRequestsIP int
	blockTime     int
}

func NewRateLimiter(repo repository.StoreKey, maxRequestsIP, blockTime int) *RateLimiter {
	return &RateLimiter{
		repo:          repo,
		maxRequestsIP: maxRequestsIP,
		blockTime:     blockTime,
	}
}

func (m *RateLimiter) AllowIP(key string) bool {
	return true
}

func (m *RateLimiter) AllowToken(key string) bool {
	return true
}
