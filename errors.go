package leader

import "errors"

var (
	ErrNoLock      = errors.New("no lock found")
	ErrLockExists  = errors.New("valid lock exists")
	ErrDidntObtain = errors.New("failed to obtain lock")
)
