package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/henrywhitaker3/leader"
	"github.com/redis/go-redis/v9"
)

func main() {
	leader := &leader.LeaderManager{
		Name:     "bongo",
		Instance: uuid.NewString(),
		Locker:   leader.NewRedisLocker(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
		OnElection: func(instance string) {
			fmt.Printf("[%s] - elected\n", instance)
		},
		OnOusting: func(instance string) {
			fmt.Printf("[%s] - ousted\n", instance)
		},
		OnRenewal: func(instance string) {
			fmt.Printf("[%s] - renewed\n", instance)
		},
		OnError: func(instance string, err error) {
			fmt.Printf("[%s] - error - %s", instance, err.Error())
		},
	}

	fmt.Println(leader.Instance)

	leader.Run(context.Background())
}
