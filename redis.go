package leader

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	_ Locker = &RedisLocker{}
)

type RedisLocker struct {
	redis *redis.Client
	name  string
}

func NewRedisLocker(name string, redis *redis.Client) *RedisLocker {
	return &RedisLocker{
		redis: redis,
		name:  name,
	}
}

func (r *RedisLocker) ObtainLock(ctx context.Context, instance string) (*Lock, error) {
	lock := NewLock(instance)
	if err := r.redis.Set(ctx, r.getKey(), lock, 0); err.Err() != nil {
		return nil, err.Err()
	}
	return lock, nil
}

func (r *RedisLocker) RenewLock(ctx context.Context, instance string) (*Lock, error) {
	return r.ObtainLock(ctx, instance)
}

func (r *RedisLocker) GetLock(ctx context.Context) (*Lock, error) {
	out := r.redis.Get(ctx, r.getKey())
	if out.Err() == redis.Nil {
		return nil, ErrNoLock
	}
	lock := &Lock{}
	if err := out.Scan(lock); err != nil {
		return nil, err
	}
	return lock, nil
}

func (r *RedisLocker) ReleaseLock(ctx context.Context, instance string) error {
	lock, err := r.GetLock(ctx)
	if err != nil {
		return err
	}
	if lock.Instance != instance {
		// The lock is not ours, do nothing
		return nil
	}
	if out := r.redis.Del(ctx, r.getKey()); out.Err() != nil {
		return err
	}
	return nil
}

func (r *RedisLocker) getKey() string {
	return fmt.Sprintf("%s-leader", r.name)
}
