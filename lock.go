package leader

import (
	"encoding/json"
	"time"
)

type Lock struct {
	Instance string
	Expires  time.Time
}

func NewLock(instance string) *Lock {
	return &Lock{
		Instance: instance,
		Expires:  now().Add(time.Second * 15),
	}
}

// Determines if the lock has already expired
func (l *Lock) IsValid() bool {
	return time.Now().Before(l.Expires)
}

func (l Lock) MarshalBinary() ([]byte, error) {
	return json.Marshal(l)
}

func (l *Lock) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}
