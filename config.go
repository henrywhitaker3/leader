package leader

import (
	"time"

	"github.com/google/uuid"
)

var (
	defaultLockDuration    time.Duration = time.Second * 15
	defaultRenewalInterval time.Duration = time.Second * 10
)

type LeaderManagerConfig struct {
	// The shared name of the lock/election
	Name string

	// The unique id of the instance
	Instance string

	// The locking system
	Locker Locker

	// The length of time before the lock expires
	LockDuration time.Duration

	// The freuqency that the leader attemps to renew the lock
	RenewInterval time.Duration

	// Functions that are called during various stages of the election process
	Callbacks *Callbacks
}

type Callbacks struct {
	// Called when this instance is elected leader
	OnElection func(instance string)
	// Called when renewing the lock
	OnRenewal func(instance string)
	// Called when the instance loses the lock
	OnOusting func(instance string)
	// Called when an error occurs during the leader election
	OnError func(instance string, err error)
	// Called when another member becomes the leader
	OnNewLeader func(instance string, newLeader string)
}

func (c *LeaderManagerConfig) validate() error {
	if c.LockDuration == 0 {
		c.LockDuration = defaultLockDuration
	}
	if c.RenewInterval == 0 {
		c.RenewInterval = defaultRenewalInterval
	}
	if c.Callbacks == nil {
		c.Callbacks = &Callbacks{}
	}
	if c.Instance == "" {
		c.Instance = uuid.NewString()
	}
	if c.Name == "" {
		return ErrNoName
	}
	return nil
}
