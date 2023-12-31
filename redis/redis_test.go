package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/henrywhitaker3/leader"
	"github.com/stretchr/testify/assert"
)

func TestItErrorsWhenThereIsNoLock(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").RedisNil()

	redis := NewRedisLocker(client)
	lock, err := redis.GetLock(context.Background(), "leader")
	assert.Nil(t, lock)
	assert.ErrorIs(t, err, leader.ErrNoLock)
}

func TestItSetsALockInRedisOnObtain(t *testing.T) {
	leader.Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectSetNX("leader-leader", &leader.Lock{
		Instance: "bongo",
		Expires:  leader.Now().Add(time.Second * 15),
	}, time.Second*15).SetVal(true)

	redis := NewRedisLocker(client)
	lock, err := redis.ObtainLock(context.Background(), "leader", "bongo", time.Second*15)
	assert.NotNil(t, lock)
	assert.Nil(t, err)
}

func TestItRetirevesTheLock(t *testing.T) {
	leader.Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").SetVal(`{"Instance":"bongo","Expires":"2023-11-21T15:04:20Z"}`)

	redis := NewRedisLocker(client)

	lock, err := redis.GetLock(context.Background(), "leader")
	assert.Nil(t, err)
	assert.Equal(t, &leader.Lock{
		Instance: "bongo",
		Expires:  leader.Now().Add(time.Second * 15),
	}, lock)
}

func TestItErrorsWhenRenewingSomeoneElsesLock(t *testing.T) {
	leader.Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").SetVal(`{"Instance":"bingo","Expires":"2023-11-21T15:04:20Z"}`)

	redis := NewRedisLocker(client)

	_, err := redis.RenewLock(context.Background(), "leader", "bongo", time.Second*15)
	assert.ErrorIs(t, err, leader.ErrRenewNotOurLock)
}

func TestItErrorsWhenRenewingAMissingLock(t *testing.T) {
	leader.Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").RedisNil()

	redis := NewRedisLocker(client)

	_, err := redis.RenewLock(context.Background(), "leader", "bongo", time.Second*15)
	assert.ErrorIs(t, err, leader.ErrNoLock)
}

func TestItRenewsALock(t *testing.T) {
	leader.Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").SetVal(`{"Instance":"bongo","Expires":"2023-11-21T15:04:20Z"}`)
	mock.ExpectSet("leader-leader", &leader.Lock{
		Instance: "bongo",
		Expires:  leader.Now().Add(time.Second * 15),
	}, time.Second*15).SetVal("OK")

	redis := NewRedisLocker(client)

	lock, err := redis.RenewLock(context.Background(), "leader", "bongo", time.Second*15)
	assert.Nil(t, err)
	assert.Equal(t, lock, &leader.Lock{
		Instance: "bongo",
		Expires:  leader.Now().Add(time.Second * 15),
	})
}

func TestItErrorsWhenObtainingLockThatExists(t *testing.T) {
	leader.Now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectSetNX("leader-leader", &leader.Lock{
		Instance: "bongo",
		Expires:  leader.Now().Add(time.Second * 15),
	}, time.Second*15).SetVal(false)

	redis := NewRedisLocker(client)

	_, err := redis.ObtainLock(context.Background(), "leader", "bongo", time.Second*15)
	assert.ErrorIs(t, err, leader.ErrLockExists)
}
