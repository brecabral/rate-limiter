package repository

import "github.com/brecabral/rate-limiter/internal/infra/token"

type StoreKey interface {
	ValidKey(key string, rate int) bool
	SaveKey(token.Token) error
}
