package leader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItPassesValidation(t *testing.T) {
	type testCase struct {
		name        string
		conf        LeaderManagerConfig
		expectedErr error
	}

	cases := []testCase{
		{
			name: "just_name_set",
			conf: LeaderManagerConfig{
				Name: "bongo",
			},
		},
		{
			name:        "no_name_set",
			conf:        LeaderManagerConfig{},
			expectedErr: ErrNoName,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			original := c.conf
			err := c.conf.validate()

			if c.expectedErr == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, c.expectedErr, err)
			}

			if original.RenewInterval == 0 {
				assert.Equal(t, defaultRenewalInterval, c.conf.RenewInterval)
			}
			if original.LockDuration == 0 {
				assert.Equal(t, defaultLockDuration, c.conf.LockDuration)
			}
			assert.NotEmpty(t, c.conf.Instance)
			assert.NotNil(t, c.conf.Callbacks)
		})
	}
}
