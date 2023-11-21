package leader

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Locker interface {
	ObtainLock(ctx context.Context, instance string) (*Lock, error)
	RenewLock(ctx context.Context, instance string) (*Lock, error)
	ReleaseLock(ctx context.Context, instance string) error
	GetLock(context.Context) (*Lock, error)
}

type LeaderManager struct {
	name     string
	instance string
	locker   Locker
	close    chan struct{}
}

func NewLeaderManager(name string, l Locker) *LeaderManager {
	return &LeaderManager{
		name:     name,
		instance: uuid.NewString(),
		locker:   l,
		close:    make(chan struct{}),
	}
}

func (m *LeaderManager) Run(ctx context.Context) {
	for {
		// Try to get a lock
		if _, err := m.AttemptLock(ctx); err != nil {
			if errors.Is(err, ErrLockExists) {
				time.Sleep(time.Second * 10)
				continue
			}
			fmt.Printf("[%s] %s\n", m.instance, err.Error())
			time.Sleep(time.Second * 10)
		}

		select {
		case <-m.close:
			return
		default:
			continue
		}
	}
}

func (m *LeaderManager) AttemptLock(ctx context.Context) (*Lock, error) {
	lock, err := m.locker.GetLock(ctx)
	if err != nil {
		if errors.Is(err, ErrNoLock) {
			fmt.Printf("[%s] no lock exists, obtaining\n", m.instance)
			return m.locker.ObtainLock(ctx, m.instance)
		}
		return nil, err
	}

	if lock.IsValid() {
		return nil, ErrLockExists
	}

	if !lock.IsValid() {
		fmt.Printf("[%s] lock expired\n", m.instance)
		if lock.Instance == m.instance {
			fmt.Printf("[%s] renewing lock\n", m.instance)
			return m.locker.RenewLock(ctx, m.instance)
		}

		fmt.Printf("[%s] obtaining lock\n", m.instance)
		return m.locker.ObtainLock(ctx, m.instance)
	}

	return nil, ErrDidntObtain
}

// Stop the leader manager
func (m *LeaderManager) Stop() {
	fmt.Printf("[%s] stopping, releasing lock\n", m.instance)
	m.locker.ReleaseLock(context.Background(), m.instance)
	close(m.close)
}

func (m *LeaderManager) IsLeader() (bool, error) {
	lock, err := m.locker.GetLock(context.Background())
	if err != nil {
		if errors.Is(err, ErrNoLock) {
			return false, nil
		}
		return false, err
	}
	return lock.Instance == m.instance, nil
}
