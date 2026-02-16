package repository

import (
	"context"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/model"
)

type StoreKey interface {
	SaveKey(ctx context.Context, apiKey model.ApiKey) error
	GetApiKeyAttributes(ctx context.Context, key string) (rate int, valid bool, block bool, err error)
	GetRequestsLastSecond(ctx context.Context, id string) (int, error)
	AddRequest(ctx context.Context, id string) error
	Block(ctx context.Context, id string, blockTime time.Duration) error
	IsBlocked(ctx context.Context, id string) (bool, error)
}
