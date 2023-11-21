package leader

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestItErrorsWhenThereIsNoLock(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").RedisNil()

	redis := NewRedisLocker(client)
	lock, err := redis.GetLock(context.Background(), "leader")
	assert.Nil(t, lock)
	assert.ErrorIs(t, err, ErrNoLock)
}

func TestItSetsALockInRedisOnObtain(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectSet("leader-leader", &Lock{
		Instance: "bongo",
		Expires:  now().Add(time.Second * 15),
	}, 0).SetVal("OK")

	redis := NewRedisLocker(client)
	lock, err := redis.ObtainLock(context.Background(), "leader", "bongo")
	assert.NotNil(t, lock)
	assert.Nil(t, err)
}

func TestItRetirevesTheLock(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").SetVal(`{"Instance":"bongo","Expires":"2023-11-21T15:04:20Z"}`)

	redis := NewRedisLocker(client)

	lock, err := redis.GetLock(context.Background(), "leader")
	assert.Nil(t, err)
	assert.Equal(t, &Lock{
		Instance: "bongo",
		Expires:  now().Add(time.Second * 15),
	}, lock)
}

func TestItErrorsWhenRenewingSomeoneElsesLock(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").SetVal(`{"Instance":"bingo","Expires":"2023-11-21T15:04:20Z"}`)

	redis := NewRedisLocker(client)

	_, err := redis.RenewLock(context.Background(), "leader", "bongo")
	assert.ErrorIs(t, err, ErrRenewNotOurLock)
}

func TestItErrorsWhenRenewingAMissingLock(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").RedisNil()

	redis := NewRedisLocker(client)

	_, err := redis.RenewLock(context.Background(), "leader", "bongo")
	assert.ErrorIs(t, err, ErrNoLock)
}

func TestItRenewsALock(t *testing.T) {
	now = func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2023-11-21T15:04:05Z")
		return fakeTime
	}
	client, mock := redismock.NewClientMock()
	mock.ExpectGet("leader-leader").SetVal(`{"Instance":"bongo","Expires":"2023-11-21T15:04:00Z"}`)
	mock.ExpectSet("leader-leader", &Lock{
		Instance: "bongo",
		Expires:  now().Add(time.Second * 15),
	}, 0).SetVal("OK")

	redis := NewRedisLocker(client)

	lock, err := redis.RenewLock(context.Background(), "leader", "bongo")
	assert.Nil(t, err)
	assert.Equal(t, lock, &Lock{
		Instance: "bongo",
		Expires:  now().Add(time.Second * 15),
	})
}
