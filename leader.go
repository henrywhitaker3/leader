package leader

import (
	"context"
	"errors"
	"time"
)

type Locker interface {
	ObtainLock(ctx context.Context, name string, instance string, expiry time.Duration) (*Lock, error)
	RenewLock(ctx context.Context, name string, instance string, expiry time.Duration) (*Lock, error)
	ReleaseLock(ctx context.Context, name string, instance string) error
	ClearLock(ctx context.Context, name string) error
	GetLock(ctx context.Context, name string) (*Lock, error)
}

type LeaderManager struct {
	Config *LeaderManagerConfig

	previousLock *Lock
	isLeader     bool
}

func NewLeaderManager(conf *LeaderManagerConfig) (*LeaderManager, error) {
	lm := &LeaderManager{
		Config: conf,
	}
	if err := lm.Config.validate(); err != nil {
		return nil, err
	}
	return lm, nil
}

func (m *LeaderManager) Run(ctx context.Context) {
	for {
		m.run(ctx)
		select {
		case <-ctx.Done():
			m.Config.Locker.ReleaseLock(ctx, m.Config.Name, m.Config.Instance)
			return
		case <-time.After(m.Config.RenewInterval):
			continue
		}
	}
}

func (m *LeaderManager) run(ctx context.Context) {
	lock, err := m.attemptLock(ctx)
	if err != nil && lock == nil {
		if m.Config.Callbacks.OnError != nil {
			m.Config.Callbacks.OnError(m.Config.Instance, err)
		}
		return
	}

	if errors.Is(err, ErrLockExists) {
		if m.previousLock != nil {
			if m.previousLock.Instance == m.Config.Instance && lock.Instance != m.Config.Instance && m.Config.Callbacks.OnOusting != nil {
				m.Config.Callbacks.OnOusting(m.Config.Instance)
			}
		}
	}

	m.previousLock = lock
	m.isLeader = (lock.Instance == m.Config.Instance)
}

func (m *LeaderManager) attemptLock(ctx context.Context) (*Lock, error) {
	lock, err := m.Config.Locker.GetLock(ctx, m.Config.Name)
	if err != nil {
		if errors.Is(err, ErrNoLock) {
			return m.obtainLock(ctx)
		}

		if m.Config.Callbacks.OnError != nil {
			m.Config.Callbacks.OnError(m.Config.Instance, err)
		}

		return lock, err
	}

	if lock.Instance == m.Config.Instance {
		if m.Config.Callbacks.OnRenewal != nil {
			m.Config.Callbacks.OnRenewal(m.Config.Instance)
		}
		return m.Config.Locker.RenewLock(ctx, m.Config.Name, m.Config.Instance, m.Config.LockDuration)
	}

	if lock.IsValid() {
		return lock, ErrLockExists
	}
	// The lock currently there is invalid, clear it and get a new one
	if err := m.Config.Locker.ClearLock(ctx, m.Config.Name); err != nil {
		return nil, err
	}
	return m.obtainLock(ctx)
}

func (m *LeaderManager) obtainLock(ctx context.Context) (*Lock, error) {
	lock, err := m.Config.Locker.ObtainLock(ctx, m.Config.Name, m.Config.Instance, m.Config.LockDuration)

	if err != nil {
		if m.Config.Callbacks.OnError != nil && !errors.Is(err, ErrLockExists) {
			m.Config.Callbacks.OnError(m.Config.Instance, err)
		}
		return lock, err
	}
	if m.Config.Callbacks.OnElection != nil {
		m.Config.Callbacks.OnElection(m.Config.Instance)
	}

	return lock, nil
}

func (m *LeaderManager) IsLeader() bool {
	return m.isLeader
}
