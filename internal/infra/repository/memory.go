package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/model"
)

type MemoryRepository struct {
	mu       sync.Mutex
	keys     map[string]memoryKey
	requests map[string][]int64
	blocks   map[string]time.Time
}

type memoryKey struct {
	rate       int
	expiration time.Time
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		keys:     make(map[string]memoryKey),
		requests: make(map[string][]int64),
		blocks:   make(map[string]time.Time),
	}
}

func (m *MemoryRepository) SaveKey(_ context.Context, apiKey model.ApiKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ttl := time.Until(apiKey.Expiration)
	if ttl <= 0 {
		return fmt.Errorf("api key already expired")
	}

	m.keys[fmt.Sprintf("key:%s", apiKey.Key)] = memoryKey{
		rate:       apiKey.RateLimitPerSecond,
		expiration: apiKey.Expiration,
	}
	return nil
}

func (m *MemoryRepository) GetApiKeyAttributes(_ context.Context, key string) (rate int, valid bool, block bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	k, ok := m.keys[fmt.Sprintf("key:%s", key)]
	if !ok || time.Now().After(k.expiration) {
		return 0, false, false, nil
	}

	blockKey := fmt.Sprintf("block:%s:%s", PrefixAPIKey, key)
	if expiry, blocked := m.blocks[blockKey]; blocked {
		if time.Now().Before(expiry) {
			return k.rate, true, true, nil
		}
		delete(m.blocks, blockKey)
	}

	return k.rate, true, false, nil
}

func (m *MemoryRepository) GetRequestsLastSecond(_ context.Context, prefix, id string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	redisKey := fmt.Sprintf("requests:%s:%s", prefix, id)
	windowStart := time.Now().Add(-1 * time.Second).UnixNano()

	count := 0
	for _, ts := range m.requests[redisKey] {
		if ts >= windowStart {
			count++
		}
	}
	return count, nil
}

func (m *MemoryRepository) AddRequest(_ context.Context, prefix, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	redisKey := fmt.Sprintf("requests:%s:%s", prefix, id)
	now := time.Now().UnixNano()
	cutoff := time.Now().Add(-2 * time.Second).UnixNano()

	var filtered []int64
	for _, ts := range m.requests[redisKey] {
		if ts >= cutoff {
			filtered = append(filtered, ts)
		}
	}
	m.requests[redisKey] = append(filtered, now)
	return nil
}

func (m *MemoryRepository) Block(_ context.Context, prefix, id string, blockTime time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if blockTime <= 0 {
		return fmt.Errorf("block time must be positive")
	}

	blockKey := fmt.Sprintf("block:%s:%s", prefix, id)
	m.blocks[blockKey] = time.Now().Add(blockTime)
	return nil
}

func (m *MemoryRepository) IsBlocked(_ context.Context, prefix, id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	blockKey := fmt.Sprintf("block:%s:%s", prefix, id)
	expiry, ok := m.blocks[blockKey]
	if !ok {
		return false, nil
	}

	if time.Now().After(expiry) {
		delete(m.blocks, blockKey)
		return false, nil
	}

	return true, nil
}
