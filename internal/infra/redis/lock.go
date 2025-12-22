package redis

import (
	"context"
	"errors"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Locker struct {
	Client *redis.Client
}

type redisLock struct {
	client *redis.Client
	key    string
	token  string
}

var releaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

func (l *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (ports.Lock, error) {
	if l == nil || l.Client == nil {
		return nil, errors.New("redis client not configured")
	}
	token := uuid.NewString()
	ok, err := l.Client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.NewConflictError("lock not acquired")
	}
	return &redisLock{client: l.Client, key: key, token: token}, nil
}

func (l *redisLock) Release(ctx context.Context) error {
	if l == nil || l.client == nil {
		return nil
	}
	_, err := releaseScript.Run(ctx, l.client, []string{l.key}, l.token).Result()
	return err
}
