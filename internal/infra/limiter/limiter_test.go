package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/model"
	"github.com/brecabral/rate-limiter/internal/infra/repository"
)

func TestAllowIP_WithinLimit(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 5, 60)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if !rl.Allow(ctx, "192.168.1.1", "") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestAllowIP_ExceedsLimit(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 3, 60)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if !rl.Allow(ctx, "10.0.0.1", "") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	if rl.Allow(ctx, "10.0.0.1", "") {
		t.Fatal("request 4 should be blocked after exceeding limit")
	}
}

func TestAllowIP_BlockPersists(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 2, 60)
	ctx := context.Background()

	// exhaust limit
	rl.Allow(ctx, "10.0.0.1", "")
	rl.Allow(ctx, "10.0.0.1", "")
	rl.Allow(ctx, "10.0.0.1", "") // triggers block

	// subsequent requests should still be blocked
	if rl.Allow(ctx, "10.0.0.1", "") {
		t.Fatal("should remain blocked")
	}
}

func TestAllowIP_DifferentIPsAreIndependent(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 2, 60)
	ctx := context.Background()

	rl.Allow(ctx, "10.0.0.1", "")
	rl.Allow(ctx, "10.0.0.1", "")
	rl.Allow(ctx, "10.0.0.1", "") // blocks 10.0.0.1

	if !rl.Allow(ctx, "10.0.0.2", "") {
		t.Fatal("different IP should not be affected")
	}
}

func TestAllowKey_WithinLimit(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 2, 60)
	ctx := context.Background()

	apiKey := model.CreateApiKey(1*time.Hour, 5)
	repo.SaveKey(ctx, apiKey)

	for i := 0; i < 5; i++ {
		if !rl.Allow(ctx, "10.0.0.1", apiKey.Key) {
			t.Fatalf("token request %d should be allowed", i+1)
		}
	}
}

func TestAllowKey_ExceedsLimit(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 2, 60)
	ctx := context.Background()

	apiKey := model.CreateApiKey(1*time.Hour, 3)
	repo.SaveKey(ctx, apiKey)

	for i := 0; i < 3; i++ {
		if !rl.Allow(ctx, "10.0.0.1", apiKey.Key) {
			t.Fatalf("token request %d should be allowed", i+1)
		}
	}

	if rl.Allow(ctx, "10.0.0.1", apiKey.Key) {
		t.Fatal("token request 4 should be blocked")
	}
}

func TestAllowKey_OverridesIPLimit(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 2, 60) // IP limit = 2
	ctx := context.Background()

	apiKey := model.CreateApiKey(1*time.Hour, 10) // token limit = 10
	repo.SaveKey(ctx, apiKey)

	// should allow more than IP limit when using token
	for i := 0; i < 10; i++ {
		if !rl.Allow(ctx, "10.0.0.1", apiKey.Key) {
			t.Fatalf("token request %d should be allowed (token limit is 10)", i+1)
		}
	}
}

func TestAllowKey_InvalidKeyDenied(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 5, 60)
	ctx := context.Background()

	if rl.Allow(ctx, "10.0.0.1", "nonexistent-key") {
		t.Fatal("invalid token should be denied")
	}
}

func TestAllowKey_ExpiredKeyDenied(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 5, 60)
	ctx := context.Background()

	apiKey := model.CreateApiKey(1*time.Millisecond, 10)
	repo.SaveKey(ctx, apiKey)

	time.Sleep(5 * time.Millisecond)

	if rl.Allow(ctx, "10.0.0.1", apiKey.Key) {
		t.Fatal("expired token should be denied")
	}
}

func TestAllowIP_BlockExpiresAfterBlockTime(t *testing.T) {
	repo := repository.NewMemoryRepository()
	rl := NewRateLimiter(repo, 2, 1) // blockTime = 1 segundo
	ctx := context.Background()

	// exhaust limit to trigger block
	rl.Allow(ctx, "10.0.0.1", "")
	rl.Allow(ctx, "10.0.0.1", "")
	rl.Allow(ctx, "10.0.0.1", "") // triggers block

	if rl.Allow(ctx, "10.0.0.1", "") {
		t.Fatal("should be blocked immediately after exceeding limit")
	}

	// wait for block to expire
	time.Sleep(1100 * time.Millisecond)

	if !rl.Allow(ctx, "10.0.0.1", "") {
		t.Fatal("should be allowed after block time expires")
	}
}
