package repository

import (
	"github.com/brecabral/rate-limiter/internal/infra/model"
)

type RedisRepository struct{}

func NewRedisRepository() *RedisRepository {
	return &RedisRepository{}
}

func (r *RedisRepository) SaveKey(model.Token) error {
	return nil
}

func (r *RedisRepository) ValidKey(key string, rate int) bool {
	return true
}
