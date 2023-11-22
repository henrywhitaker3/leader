# Leader election

Small package that provides an extensible leader election mechanism. Currently supports:

- redis

## Usage

```go
mgr, err := leader.NewLeaderManager(&leader.LeaderManagerConfig{
    Name: "bongo",
    Locker: redis.NewRedisLocker(redis.NewClient(...)),
    Callbacks: &leader.Callbacks{
        OnElection: func(instance string) {
            fmt.Printf("%s elected\n", instance)
        }
        OnRenewal: func(instance string) {
            fmt.Printf("%s renewed\n", instance)
        }
        OnOusting: func(instance string) {
            fmt.Printf("%s ousted\n", instance)
        }
        OnError: func(instance string, err error) {
            fmt.Printf("%s - %s\n", instance, err.Err())
        }
    }
})

ctx := context.Background()
ctx, cancel := context.WithCancel(ctx)
go mgr.Run(ctx)

time.Sleep(25)
cancel()
```
