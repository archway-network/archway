package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProposalsRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		query       ProposalsRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name:  "OK: Empty proposal request",
			query: ProposalsRequest{},
		},
		{
			name: "OK: Valid voter",
			query: ProposalsRequest{
				Voter: "cosmos14450hpujwlct9x0la3wv46sgk79czrl9phh0dm",
			},
		},
		{
			name: "Fail: Invalid voter address",
			query: ProposalsRequest{
				Voter: "invalid",
			},
			errExpected: true,
		},
		{
			name: "OK: Valid depositor",
			query: ProposalsRequest{
				Depositor: "cosmos14450hpujwlct9x0la3wv46sgk79czrl9phh0dm",
			},
		},
		{
			name: "Fail: Invalid depositor address",
			query: ProposalsRequest{
				Depositor: "invalid",
			},
			errExpected: true,
		},
		{
			name: "OK: Valid status",
			query: ProposalsRequest{
				Status: "PROPOSAL_STATUS_DEPOSIT_PERIOD",
			},
		},
		{
			name: "Fail: Invalid status",
			query: ProposalsRequest{
				Status: "NON_EXISTENT",
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.query.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestVoteRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		query       VoteRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: Valid request",
			query: VoteRequest{
				ProposalId: 1,
				Voter:      "cosmos14450hpujwlct9x0la3wv46sgk79czrl9phh0dm",
			},
		},
		{
			name: "Fail: Missing proposal id",
			query: VoteRequest{
				Voter: "cosmos14450hpujwlct9x0la3wv46sgk79czrl9phh0dm",
			},
			errExpected: true,
		},
		{
			name: "Fail: Missing voter",
			query: VoteRequest{
				ProposalId: 1,
			},
			errExpected: true,
		},
		{
			name:        "Fail: Empty request",
			query:       VoteRequest{},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.query.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
