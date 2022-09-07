package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVotingRequestValidate(t *testing.T) {
	type testCase struct {
		name string
		msg  NewVotingRequest
		//
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK",
			msg: NewVotingRequest{
				Name:        "Test",
				VoteOptions: []string{"a", "b", "c"},
				Duration:    1000,
			},
		},
		{
			name: "Fail: Name: empty",
			msg: NewVotingRequest{
				Name:        "",
				VoteOptions: []string{"a", "b", "c"},
				Duration:    1000,
			},
			errExpected: true,
		},
		{
			name: "Fail: VoteOptions: empty",
			msg: NewVotingRequest{
				Name:        "Test",
				VoteOptions: []string{},
				Duration:    1000,
			},
			errExpected: true,
		},
		{
			name: "Fail: VoteOptions: empty option",
			msg: NewVotingRequest{
				Name:        "Test",
				VoteOptions: []string{"a", "", "c"},
				Duration:    1000,
			},
			errExpected: true,
		},
		{
			name: "Fail: Duration: 0",
			msg: NewVotingRequest{
				Name:        "Test",
				VoteOptions: []string{"a", "b", "c"},
				Duration:    0,
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errExpected {
				assert.Error(t, tc.msg.Validate())
				return
			}
			assert.NoError(t, tc.msg.Validate())
		})
	}
}

func TestVoteRequestValidate(t *testing.T) {
	type testCase struct {
		name string
		msg  VoteRequest
		//
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: yes",
			msg: VoteRequest{
				Option: "a",
				Vote:   "yes",
			},
		},
		{
			name: "OK: no",
			msg: VoteRequest{
				Option: "a",
				Vote:   "no",
			},
		},
		{
			name: "Fail: Option: empty",
			msg: VoteRequest{
				Vote: "no",
			},
			errExpected: true,
		},
		{
			name: "Fail: Vote: invalid",
			msg: VoteRequest{
				Option: "a",
				Vote:   "NO",
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errExpected {
				assert.Error(t, tc.msg.Validate())
				return
			}
			assert.NoError(t, tc.msg.Validate())
		})
	}
}
