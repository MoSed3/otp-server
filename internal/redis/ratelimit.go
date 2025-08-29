package redis

import (
	"context"
	"fmt"
	"time"
)

func (c *Config) CheckRateLimit(ctx context.Context, key string, maxRequests int, windowSeconds int) (bool, int, error) {
	pipe := c.client.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := incr.Val()
	remaining := maxRequests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return count <= int64(maxRequests), remaining, nil
}

func (c *Config) GetRateLimitKey(prefix, identifier string) string {
	return fmt.Sprintf("rate_limit:%s:%s", prefix, identifier)
}
