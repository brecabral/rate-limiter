package repository

import (
	"context"
	"fmt"

	"strconv"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(addr, password string, db int) *RedisRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return &RedisRepository{client: client}
}

// API KEY
func (r *RedisRepository) SaveKey(ctx context.Context, apiKey model.ApiKey) error {
	ttl := time.Until(apiKey.Expiration)
	if ttl <= 0 {
		return fmt.Errorf("api key already expired")
	}

	redisKey := fmt.Sprintf("key:%s", apiKey.Key)
	rate := strconv.Itoa(apiKey.RateLimitPerSecond)
	return r.client.Set(ctx, redisKey, rate, ttl).Err()
}

func (r *RedisRepository) GetApiKeyAttributes(ctx context.Context, key string) (rate int, valid bool, block bool, err error) {
	redisKey := fmt.Sprintf("key:%s", key)

	value, err := r.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return 0, false, false, nil
	}
	if err != nil {
		return 0, false, false, err
	}
	valid = true

	block, err = r.IsBlocked(ctx, "apikey:", key)
	if err != nil {
		return
	}

	rate, err = strconv.Atoi(value)
	return
}

// API KEY and IP
func (r *RedisRepository) GetRequestsLastSecond(ctx context.Context, prefix, id string) (int, error) {
	windowEnd := time.Now().UnixNano()
	windowStart := time.Now().Add(-1 * time.Second).UnixNano()

	redisKey := fmt.Sprintf("requests:%s:%s", prefix, id)
	count, err := r.client.ZCount(ctx, redisKey, fmt.Sprintf("%d", windowStart), fmt.Sprintf("%d", windowEnd)).Result()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *RedisRepository) AddRequest(ctx context.Context, prefix, id string) error {
	now := time.Now().UnixNano()
	redisKey := fmt.Sprintf("requests:%s:%s", prefix, id)

	member := redis.Z{
		Score:  float64(now),
		Member: now,
	}

	pipe := r.client.Pipeline()
	pipe.ZAdd(ctx, redisKey, member)
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", time.Now().Add(-2*time.Second).UnixNano()))
	pipe.Expire(ctx, redisKey, 2*time.Second)
	_, err := pipe.Exec(ctx)

	return err
}

func (r *RedisRepository) Block(ctx context.Context, prefix, id string, blockTime time.Duration) error {
	redisKey := fmt.Sprintf("block:%s:%s", prefix, id)
	ttl := blockTime
	if ttl <= 0 {
		return fmt.Errorf("api key already expired")
	}
	return r.client.Set(ctx, redisKey, 1, blockTime).Err()
}

func (r *RedisRepository) IsBlocked(ctx context.Context, prefix, id string) (bool, error) {
	redisKey := fmt.Sprintf("block:%s:%s", prefix, id)

	exists, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	if exists == 0 {
		return false, nil
	}

	return true, nil
}
