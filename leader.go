package leader

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Locker interface {
	ObtainLock(ctx context.Context, name string, instance string) (*Lock, error)
	RenewLock(ctx context.Context, name string, instance string) (*Lock, error)
	ReleaseLock(ctx context.Context, name string, instance string) error
	GetLock(ctx context.Context, name string) (*Lock, error)
}

type LeaderManager struct {
	Name       string
	Instance   string
	Locker     Locker
	OnElection func(instance string)
	OnOusting  func(instance string)
	OnRenewal  func(instance string)
	OnError    func(instance string, err error)

	previousLock *Lock
}

func NewLeaderManager(name string, l Locker) *LeaderManager {
	return &LeaderManager{
		Name:     name,
		Instance: uuid.NewString(),
		Locker:   l,
	}
}

func (m *LeaderManager) Run(ctx context.Context) {
	for {
		m.run(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 10):
			continue
		}
	}
}

func (m *LeaderManager) run(ctx context.Context) {
	lock, err := m.AttemptLock(ctx)
	if err != nil && lock == nil {
		if m.OnError != nil {
			m.OnError(m.Instance, err)
		}
		return
	}

	if errors.Is(err, ErrLockExists) {
		if m.previousLock != nil {
			if m.previousLock.Instance == m.Instance && lock.Instance != m.Instance && m.OnOusting != nil {
				m.OnOusting(m.Instance)
			}
		}
	}

	m.previousLock = lock
}

func (m *LeaderManager) AttemptLock(ctx context.Context) (*Lock, error) {
	lock, err := m.Locker.GetLock(ctx, m.Name)
	if err != nil {
		if errors.Is(err, ErrNoLock) {
			return m.obtainLock(ctx)
		}

		if m.OnError != nil {
			m.OnError(m.Instance, err)
		}

		return nil, err
	}

	if lock.Instance == m.Instance {
		if m.OnRenewal != nil {
			m.OnRenewal(m.Instance)
		}
		return m.Locker.RenewLock(ctx, m.Name, m.Instance)
	}

	if lock.IsValid() {
		return lock, ErrLockExists
	}
	return m.obtainLock(ctx)
}

func (m *LeaderManager) obtainLock(ctx context.Context) (*Lock, error) {
	lock, err := m.Locker.ObtainLock(ctx, m.Name, m.Instance)

	if err != nil && m.OnError != nil {
		m.OnError(m.Instance, err)
		return lock, err
	}
	if m.OnElection != nil {
		m.OnElection(m.Instance)
	}

	return lock, nil
}

func (m *LeaderManager) IsLeader() (bool, error) {
	lock, err := m.Locker.GetLock(context.Background(), m.Name)
	if err != nil {
		if errors.Is(err, ErrNoLock) {
			return false, nil
		}
		return false, err
	}
	return lock.Instance == m.Instance, nil
}
