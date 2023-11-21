package leader

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisLocker struct {
	redis *redis.Client
	name  string
	id    string
}

func NewRedisLocker(name string, redis *redis.Client) *RedisLocker {
	return &RedisLocker{
		redis: redis,
		name:  name,
		id:    uuid.NewString(),
	}
}

func (r *RedisLocker) ObtainLock(ctx context.Context) (*Lock, error) {
	lock := &Lock{
		Instance: r.id,
		Expires:  time.Now().Add(time.Second * 15),
	}
	if err := r.redis.Set(ctx, r.getKey(), lock, 0); err.Err() != nil {
		return nil, err.Err()
	}
	return lock, nil
}

func (r *RedisLocker) RenewLock(ctx context.Context) (*Lock, error) {
	return r.ObtainLock(ctx)
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

func (r *RedisLocker) GetId() string {
	return r.id
}

func (r *RedisLocker) getKey() string {
	return fmt.Sprintf("%s-leader", r.name)
}
