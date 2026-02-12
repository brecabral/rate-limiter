package token

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	key                string
	expiration         time.Time
	rateLimitPerSecond int
}

func CreateToken(duration time.Duration, rate int) Token {
	return Token{
		key:                uuid.New().String(),
		expiration:         time.Now().Add(duration),
		rateLimitPerSecond: rate,
	}
}
