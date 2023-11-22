package leader

import (
	"context"
	"time"
)

type mockLocker struct {
	lock *Lock
}

var _ Locker = &mockLocker{}

func (m *mockLocker) ObtainLock(ctx context.Context, name string, instance string, expiry time.Duration) (*Lock, error) {
	if m.lock == nil {
		m.lock = NewLock(instance)
		return m.lock, nil
	}

	if m.lock.IsValid() {
		return m.lock, ErrLockExists
	}

	m.lock = NewLock(instance)
	return m.lock, nil
}

func (m *mockLocker) RenewLock(ctx context.Context, name string, instance string, expiry time.Duration) (*Lock, error) {
	if m.lock == nil {
		return nil, ErrNoLock
	}

	if m.lock.Instance != instance {
		return m.lock, ErrRenewNotOurLock
	}

	m.lock = NewLock(instance)
	return m.lock, nil
}

func (m *mockLocker) ReleaseLock(ctx context.Context, name string, instance string) error {
	if m.lock.Instance != instance {
		return nil
	}
	return m.ClearLock(ctx, name)
}

func (m *mockLocker) ClearLock(ctx context.Context, name string) error {
	m.lock = nil
	return nil
}

func (m *mockLocker) GetLock(ctx context.Context, name string) (*Lock, error) {
	if m.lock == nil {
		return nil, ErrNoLock
	}
	return m.lock, nil
}
