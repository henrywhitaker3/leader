package leader

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestItElectsALeaderWhenThereIsNoLock(t *testing.T) {
	Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader, _ := NewLeaderManager(&LeaderManagerConfig{
		Name:     "bongo",
		Instance: "1",
		Locker:   &mockLocker{},
		Callbacks: &Callbacks{
			OnElection: func(instance string) {
				onElectionCalled = true
			},
			OnRenewal: func(instance string) {
				onRenewalCalled = true
			},
			OnOusting: func(instance string) {
				onOustingCalled = true
			},
			OnError: func(instance string, err error) {
				onErrorCalled = true
			},
		},
	})

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItRenewsItsOwnLock(t *testing.T) {
	Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader, _ := NewLeaderManager(&LeaderManagerConfig{
		Name:     "bongo",
		Instance: "1",
		Locker:   &mockLocker{},

		Callbacks: &Callbacks{
			OnElection: func(instance string) {
				onElectionCalled = true
			},
			OnRenewal: func(instance string) {
				onRenewalCalled = true
			},
			OnOusting: func(instance string) {
				onOustingCalled = true
			},
			OnError: func(instance string, err error) {
				fmt.Println(err)
				onErrorCalled = true
			},
		},
	})

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)

	onElectionCalled = false
	leader.run(context.Background())

	assert.False(t, onElectionCalled)
	assert.True(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItCallsOnOustedWhenAnotherInstanceTakesOverTheLock(t *testing.T) {
	Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	mock := &mockLocker{}

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader, _ := NewLeaderManager(&LeaderManagerConfig{
		Name:     "bongo",
		Instance: "1",
		Locker:   mock,

		Callbacks: &Callbacks{
			OnElection: func(instance string) {
				onElectionCalled = true
			},
			OnRenewal: func(instance string) {
				onRenewalCalled = true
			},
			OnOusting: func(instance string) {
				onOustingCalled = true
			},
			OnError: func(instance string, err error) {
				fmt.Println(err)
				onErrorCalled = true
			},
		},
	})

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)

	onElectionCalled = false
	mock.lock = NewLock("2")

	leader.run(context.Background())

	assert.False(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.True(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItTakesOverTheLockWhenTheCurrentLockHasExpired(t *testing.T) {
	Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader, _ := NewLeaderManager(&LeaderManagerConfig{
		Name:     "bongo",
		Instance: "1",
		Locker:   &mockLocker{},

		Callbacks: &Callbacks{
			OnElection: func(instance string) {
				onElectionCalled = true
			},
			OnRenewal: func(instance string) {
				onRenewalCalled = true
			},
			OnOusting: func(instance string) {
				onOustingCalled = true
			},
			OnError: func(instance string, err error) {
				fmt.Println(err)
				onErrorCalled = true
			},
		},
	})

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItSetsIsLeaderCorrectlyAfterRun(t *testing.T) {
	Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	mock := &mockLocker{}

	leader, _ := NewLeaderManager(&LeaderManagerConfig{
		Name:     "bongo",
		Instance: "1",
		Locker:   mock,
		Callbacks: &Callbacks{
			OnError: func(instance string, err error) {
				fmt.Println(err)
			},
		},
	})

	mock.lock = NewLock("2")
	leader.run(context.Background())
	assert.False(t, leader.IsLeader())

	mock.lock = NewLock("1")
	leader.run(context.Background())
	assert.True(t, leader.IsLeader())
}

func TestItBehavesCorrectlyWhenSettingNXThatExists(t *testing.T) {
	Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	mock := &mockLocker{}

	leader, _ := NewLeaderManager(&LeaderManagerConfig{
		Name:     "bongo",
		Instance: "1",
		Locker:   mock,
	})
	mock.lock = NewLock("2")

	_, err := leader.obtainLock(context.Background())

	assert.ErrorIs(t, err, ErrLockExists)
}
