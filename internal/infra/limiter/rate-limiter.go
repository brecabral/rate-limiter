package limiter

import (
	"log"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/repository"
)

type RateLimiter struct {
	repo          repository.StoreKey
	maxRequestsIP int
	blockTime     time.Duration
	blockList     map[string]time.Time
}

func NewRateLimiter(repo repository.StoreKey, maxRequestsIP, blockTimeSeconds int) *RateLimiter {
	return &RateLimiter{
		repo:          repo,
		maxRequestsIP: maxRequestsIP,
		blockTime:     time.Duration(blockTimeSeconds) * time.Second,
	}
}

func (l *RateLimiter) AllowIP(key string) bool {
	if time.Since(l.blockList[key]) < l.blockTime {
		return false
	}
	delete(l.blockList, key)

	count := l.repo.GetRequestsLastSecond(key)
	if count >= l.maxRequestsIP {
		l.blockList[key] = time.Now()
		return false
	}

	err := l.repo.AddRequest(key)
	if err != nil {
		log.Print(err)
	}
	return true
}

func (m *RateLimiter) AllowToken(key string) bool {
	return true
}
