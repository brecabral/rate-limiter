package model

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Key                string
	Expiration         time.Time
	RateLimitPerSecond int
}

func CreateToken(duration time.Duration, ratePerSecond int) Token {
	return Token{
		Key:                uuid.New().String(),
		Expiration:         time.Now().Add(duration),
		RateLimitPerSecond: ratePerSecond,
	}
}

func CreateManualToken(key string, duration time.Duration, ratePerSecond int) Token {
	return Token{
		Key:                key,
		Expiration:         time.Now().Add(duration),
		RateLimitPerSecond: ratePerSecond,
	}
}
