package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateChannelID(t *testing.T) {
	type testCase struct {
		channelID string
		//
		errExpected bool
	}

	testCases := []testCase{
		{
			channelID: "channel-1",
		},
		{
			channelID: "channel-1234567890",
		},
		{
			channelID:   "",
			errExpected: true,
		},
		{
			channelID:   "1channel-1",
			errExpected: true,
		},
		{
			channelID:   "channel-1#2",
			errExpected: true,
		},
		{
			channelID:   "channel-12345678901",
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.channelID, func(t *testing.T) {
			if tc.errExpected {
				assert.Error(t, ValidateChannelID(tc.channelID))
				return
			}
			assert.NoError(t, ValidateChannelID(tc.channelID))
		})
	}
}
