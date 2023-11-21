package leader

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestItElectsALeaderWhenThereIsNoLock(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("bongo-leader").RedisNil()
	mock.ExpectSet("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, 0).SetVal("OK")

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader := &LeaderManager{
		Name:     "bongo",
		Instance: "1",
		Locker:   NewRedisLocker(client),

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
	}

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItRenewsItsOwnLock(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("bongo-leader").RedisNil()
	mock.ExpectSet("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, 0).SetVal("OK")

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader := &LeaderManager{
		Name:     "bongo",
		Instance: "1",
		Locker:   NewRedisLocker(client),

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
	}

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)

	onElectionCalled = false
	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"1","Expires":"2023-11-21T15:04:00Z"}`)
	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"1","Expires":"2023-11-21T15:04:00Z"}`)
	mock.ExpectSet("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, 0).SetVal("OK")
	leader.run(context.Background())

	assert.False(t, onElectionCalled)
	assert.True(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItCallsOnOustedWhenAnotherInstanceTakesOverTheLock(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("bongo-leader").RedisNil()
	mock.ExpectSet("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, 0).SetVal("OK")

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader := &LeaderManager{
		Name:     "bongo",
		Instance: "1",
		Locker:   NewRedisLocker(client),

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
	}

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)

	onElectionCalled = false
	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"2","Expires":"2023-11-21T15:25:00Z"}`)

	leader.run(context.Background())

	assert.False(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.True(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItTakesOverTheLockWhenTheCurrentLockHasExpired(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()

	onElectionCalled := false
	onRenewalCalled := false
	onOustingCalled := false
	onErrorCalled := false

	leader := &LeaderManager{
		Name:     "bongo",
		Instance: "1",
		Locker:   NewRedisLocker(client),

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
	}

	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"2","Expires":"2023-11-21T15:03:00Z"}`)
	mock.ExpectSet("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, 0).SetVal("OK")

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}
