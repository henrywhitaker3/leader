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
	Redis *redis.Client
}

func NewRedisLocker(redis *redis.Client) *RedisLocker {
	return &RedisLocker{
		Redis: redis,
	}
}

func (r *RedisLocker) ObtainLock(ctx context.Context, name string, instance string) (*Lock, error) {
	lock := NewLock(instance)
	if err := r.Redis.Set(ctx, r.getKey(name), lock, 0); err.Err() != nil {
		return nil, err.Err()
	}
	return lock, nil
}

func (r *RedisLocker) RenewLock(ctx context.Context, name string, instance string) (*Lock, error) {
	lock, err := r.GetLock(ctx, name)
	if err != nil {
		return nil, err
	}
	if lock.Instance != instance {
		return nil, ErrRenewNotOurLock
	}
	return r.ObtainLock(ctx, name, instance)
}

func (r *RedisLocker) GetLock(ctx context.Context, name string) (*Lock, error) {
	out := r.Redis.Get(ctx, r.getKey(name))
	if out.Err() == redis.Nil {
		return nil, ErrNoLock
	}
	lock := &Lock{}
	if err := out.Scan(lock); err != nil {
		return nil, err
	}
	return lock, nil
}

func (r *RedisLocker) ReleaseLock(ctx context.Context, name string, instance string) error {
	lock, err := r.GetLock(ctx, name)
	if err != nil {
		return err
	}
	if lock.Instance != instance {
		// The lock is not ours, do nothing
		return nil
	}
	if out := r.Redis.Del(ctx, r.getKey(name)); out.Err() != nil {
		return err
	}
	return nil
}

func (r *RedisLocker) getKey(name string) string {
	return fmt.Sprintf("%s-leader", name)
}
