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
	return l.AllowIP(ctx, ip)
}

func (l *RateLimiter) AllowIP(ctx context.Context, ip string) bool {
	block, err := l.repo.IsBlocked(ctx, "ip", ip)
	if err != nil {
		log.Printf("limiter:ip:is-block:%s", err)
		return false
	}

	if block {
		return false
	}

	count, err := l.repo.GetRequestsLastSecond(ctx, "ip", ip)
	if err != nil {
		log.Printf("limiter:ip:get-requests:%s", err)
		return false
	}

	if count >= l.maxRequestsByIP {
		err = l.repo.Block(ctx, "ip", ip, l.blockTime)
		if err != nil {
			log.Printf("limiter:ip:block:%s", err)
		}
		return false
	}

	err = l.repo.AddRequest(ctx, "ip", ip)
	if err != nil {
		log.Printf("limiter:ip:add-request:%s", err)
	}
	return true
}

func (l *RateLimiter) AllowKey(ctx context.Context, key string) bool {
	rate, valid, block, err := l.repo.GetApiKeyAttributes(ctx, key)
	if err != nil {
		log.Printf("limiter:key:get-attributes:%s", err)
		return false
	}

	if !valid || block {
		return false
	}

	count, err := l.repo.GetRequestsLastSecond(ctx, "apikey", key)
	if err != nil {
		log.Printf("limiter:key:get-requests:%s", err)
		return false
	}

	if count >= rate {
		err = l.repo.Block(ctx, "apikey", key, l.blockTime)
		if err != nil {
			log.Printf("limiter:key:block:%s", err)
		}
		return false
	}

	err = l.repo.AddRequest(ctx, "apikey", key)
	if err != nil {
		log.Printf("limiter:key:add-request:%s", err)
	}
	return true
}
