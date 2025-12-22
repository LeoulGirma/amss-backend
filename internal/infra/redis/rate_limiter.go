package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Client *redis.Client
}

func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	if r == nil || r.Client == nil {
		return true, limit, time.Now().UTC(), nil
	}
	if limit <= 0 {
		return true, limit, time.Now().UTC(), nil
	}

	now := time.Now().UTC()
	windowStart := now.Truncate(window)
	bucketKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	pipe := r.Client.TxPipeline()
	countCmd := pipe.Incr(ctx, bucketKey)
	pipe.Expire(ctx, bucketKey, window+time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	count := int(countCmd.Val())
	remaining := limit - count
	resetAt := windowStart.Add(window)
	return count <= limit, remaining, resetAt, nil
}
