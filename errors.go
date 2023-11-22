package leader

import "errors"

var (
	ErrNoName          = errors.New("the leader manager name is not set")
	ErrNoLock          = errors.New("no lock found")
	ErrLockExists      = errors.New("valid lock exists")
	ErrDidntObtain     = errors.New("failed to obtain lock")
	ErrRenewNotOurLock = errors.New("attempted to renew a lock held by another instance")
)
