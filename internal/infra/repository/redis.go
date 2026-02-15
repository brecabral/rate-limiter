package repository

import "github.com/brecabral/rate-limiter/internal/infra/token"

type RedisRepository struct{}

func NewRedisRepository() *RedisRepository {
	return &RedisRepository{}
}

func (r *RedisRepository) SaveKey(token.Token) error {
	return nil
}

func (r *RedisRepository) ValidKey(key string, rate int) bool {
	return true
}
