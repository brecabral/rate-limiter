package model

import (
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	Key                string
	Expiration         time.Time
	RateLimitPerSecond int
}

func CreateApiKey(duration time.Duration, ratePerSecond int) ApiKey {
	return ApiKey{
		Key:                uuid.New().String(),
		Expiration:         time.Now().Add(duration),
		RateLimitPerSecond: ratePerSecond,
	}
}
