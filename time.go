package leader

import "time"

type NowFunc func() time.Time

var (
	now NowFunc = func() time.Time {
		return time.Now()
	}
)
