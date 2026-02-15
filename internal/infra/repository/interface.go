package repository

import (
	"github.com/brecabral/rate-limiter/internal/infra/model"
)

type StoreKey interface {
	SaveKey(model.Token) error
	ValidKey(key string) bool
	GetRequestsLastSecond(key string) int
	AddRequest(key string) error
}
