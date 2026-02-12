package limiter

type RateLimiter struct {
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (m *RateLimiter) AllowIP(key string) bool {
	return true
}

func (m *RateLimiter) AllowToken(key string) bool {
	return true
}
