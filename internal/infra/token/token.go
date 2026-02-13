package token

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Key                string
	Expiration         time.Time
	RateLimitPerSecond int
}

func CreateToken(duration time.Duration, rate int) Token {
	return Token{
		Key:                uuid.New().String(),
		Expiration:         time.Now().Add(duration),
		RateLimitPerSecond: rate,
	}
}
