package repository

import (
	"github.com/brecabral/rate-limiter/internal/infra/model"
)

type StoreKey interface {
	ValidKey(key string, rate int) bool
	SaveKey(model.Token) error
}
