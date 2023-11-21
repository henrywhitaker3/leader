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
	mock.ExpectSetNX("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, time.Second*15).SetVal(true)

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
	mock.ExpectSetNX("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, time.Second*15).SetVal(true)

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
	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"1","Expires":"2023-11-21T15:04:20Z"}`)
	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"1","Expires":"2023-11-21T15:04:20Z"}`)
	mock.ExpectSet("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, time.Second*15).SetVal("OK")
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
	mock.ExpectSetNX("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, time.Second*15).SetVal(true)

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
	mock.ExpectDel("bongo-leader").SetVal(1)
	mock.ExpectSetNX("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, time.Second*15).SetVal(true)

	leader.run(context.Background())

	assert.True(t, onElectionCalled)
	assert.False(t, onRenewalCalled)
	assert.False(t, onOustingCalled)
	assert.False(t, onErrorCalled)
}

func TestItSetsIsLeaderCorrectlyAfterRun(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()

	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"2","Expires":"2023-11-21T15:06:00Z"}`)

	leader := &LeaderManager{
		Name:     "bongo",
		Instance: "1",
		Locker:   NewRedisLocker(client),
		OnError: func(instance string, err error) {
			fmt.Println(err)
		},
	}

	leader.run(context.Background())

	assert.False(t, leader.IsLeader())

	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"1","Expires":"2023-11-21T15:06:00Z"}`)
	mock.ExpectGet("bongo-leader").SetVal(`{"Instance":"1","Expires":"2023-11-21T15:06:00Z"}`)
	mock.ExpectSet("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, time.Second*15).SetVal("OK")

	leader.run(context.Background())

	assert.True(t, leader.IsLeader())
}

func TestItBehavesCorrectlyWhenSettingNXThatExists(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()

	leader := &LeaderManager{
		Name:     "bongo",
		Instance: "1",
		Locker:   NewRedisLocker(client),
	}

	mock.ExpectSetNX("bongo-leader", &Lock{
		Instance: "1",
		Expires:  now().Add(time.Second * 15),
	}, time.Second*15).SetVal(false)

	_, err := leader.obtainLock(context.Background())

	assert.ErrorIs(t, err, ErrLockExists)
}
