package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/henrywhitaker3/leader"
	"github.com/redis/go-redis/v9"
)

var (
	_ leader.Locker = &RedisLocker{}
)

type RedisLocker struct {
	Redis *redis.Client
}

func NewRedisLocker(redis *redis.Client) *RedisLocker {
	return &RedisLocker{
		Redis: redis,
	}
}

func (r *RedisLocker) ObtainLock(ctx context.Context, name string, instance string, expiry time.Duration) (*leader.Lock, error) {
	lock := leader.NewLock(instance)
	res := r.Redis.SetNX(ctx, r.getKey(name), lock, expiry)
	if res.Err() != nil {
		return nil, res.Err()
	}
	if !res.Val() {
		return nil, leader.ErrLockExists
	}
	return lock, nil
}

func (r *RedisLocker) RenewLock(ctx context.Context, name string, instance string, expiry time.Duration) (*leader.Lock, error) {
	lock, err := r.GetLock(ctx, name)
	if err != nil {
		return nil, err
	}
	if lock.Instance != instance {
		return nil, leader.ErrRenewNotOurLock
	}
	lock = leader.NewLock(instance)
	if res := r.Redis.Set(ctx, r.getKey(name), lock, expiry); res.Err() != nil {
		return nil, res.Err()
	}
	return lock, nil
}

func (r *RedisLocker) GetLock(ctx context.Context, name string) (*leader.Lock, error) {
	out := r.Redis.Get(ctx, r.getKey(name))
	if out.Err() == redis.Nil {
		return nil, leader.ErrNoLock
	}
	lock := &leader.Lock{}
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
	// TODO: clean this up
	if out := r.Redis.Del(ctx, r.getKey(name)); out.Err() != nil {
		return err
	}
	return nil
}

func (r *RedisLocker) ClearLock(ctx context.Context, name string) error {
	if out := r.Redis.Del(ctx, r.getKey(name)); out.Err() != nil && out.Err() != redis.Nil {
		return out.Err()
	}
	return nil
}

func (r *RedisLocker) getKey(name string) string {
	return fmt.Sprintf("%s-leader", name)
}
