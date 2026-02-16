package limiter

import (
	"context"
	"log"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/repository"
)

type RateLimiter struct {
	repo            repository.StoreKey
	maxRequestsByIP int
	blockTime       time.Duration
}

func NewRateLimiter(repo repository.StoreKey, maxRequestsByIP, blockTimeInSeconds int) *RateLimiter {
	if blockTimeInSeconds <= 0 {
		blockTimeInSeconds = 1
	}
	b := time.Duration(blockTimeInSeconds) * time.Second
	if maxRequestsByIP <= 0 {
		maxRequestsByIP = 1
	}

	return &RateLimiter{
		repo:            repo,
		maxRequestsByIP: maxRequestsByIP,
		blockTime:       b,
	}
}

func (l *RateLimiter) Allow(ctx context.Context, ip, key string) bool {
	if key != "" {
		return l.AllowKey(ctx, key)
	}
	if ip != "" {
		return l.AllowKey(ctx, ip)
	}
	return false

}

func (l *RateLimiter) AllowIP(ctx context.Context, ip string) bool {
	block, err := l.repo.IsBlocked(ctx, ip)
	if err != nil {
		log.Print(err)
		return false
	}

	if block {
		return false
	}

	count, err := l.repo.GetRequestsLastSecond(ctx, ip)
	if err != nil {
		log.Print(err)
		return false
	}

	if count >= l.maxRequestsByIP {
		err = l.repo.Block(ctx, ip, l.blockTime)
		if err != nil {
			log.Print(err)
		}
		return false
	}

	err = l.repo.AddRequest(ctx, ip)
	if err != nil {
		log.Print(err)
	}
	return true
}

func (l *RateLimiter) AllowKey(ctx context.Context, key string) bool {
	rate, valid, block, err := l.repo.GetApiKeyAttributes(ctx, key)
	if err != nil {
		log.Print(err)
		return false
	}

	if !valid || block {
		return false
	}

	count, err := l.repo.GetRequestsLastSecond(ctx, key)
	if err != nil {
		log.Print(err)
		return false
	}

	if count >= rate {
		err = l.repo.Block(ctx, key, l.blockTime)
		if err != nil {
			log.Print(err)
		}
		return false
	}

	err = l.repo.AddRequest(ctx, key)
	if err != nil {
		log.Print(err)
	}
	return true
}
