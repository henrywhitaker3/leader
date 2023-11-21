package leader

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Locker interface {
	ObtainLock(context.Context) (*Lock, error)
	RenewLock(context.Context) (*Lock, error)
	ReleaseLock(context.Context) error
	GetLock(context.Context) (*Lock, error)
	GetId() string
}

type LeaderManager struct {
	name   string
	locker Locker
	close  chan struct{}
}

func NewLeaderManager(name string, l Locker) *LeaderManager {
	return &LeaderManager{
		name:   name,
		locker: l,
		close:  make(chan struct{}),
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
			fmt.Printf("[%s] %s\n", m.locker.GetId(), err.Error())
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
			fmt.Printf("[%s] no lock exists, obtaining\n", m.locker.GetId())
			return m.locker.ObtainLock(ctx)
		}
		return nil, err
	}

	if lock.IsValid() {
		return nil, ErrLockExists
	}

	if !lock.IsValid() {
		fmt.Printf("[%s] lock expired\n", m.locker.GetId())
		if lock.Instance == m.locker.GetId() {
			fmt.Printf("[%s] renewing lock\n", m.locker.GetId())
			return m.locker.RenewLock(ctx)
		}

		fmt.Printf("[%s] obtaining lock\n", m.locker.GetId())
		return m.locker.ObtainLock(ctx)
	}

	return nil, ErrDidntObtain
}

// Stop the leader manager
func (m *LeaderManager) Stop() {
	fmt.Printf("[%s] stopping, releasing lock\n", m.locker.GetId())
	m.locker.ReleaseLock(context.Background())
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
	return lock.Instance == m.locker.GetId(), nil
}
