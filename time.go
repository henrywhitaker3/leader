package leader

import "time"

type NowFunc func() time.Time

var (
	Now NowFunc = func() time.Time {
		return time.Now()
	}
)
