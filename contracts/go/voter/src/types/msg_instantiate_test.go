package types

import (
	"testing"

	"github.com/CosmWasm/cosmwasm-go/std/mock"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	"github.com/stretchr/testify/assert"
)

func TestMsgInstantiateValidate(t *testing.T) {
	type testCase struct {
		name string
		info stdTypes.MessageInfo
		msg  MsgInstantiate
		//
		errExpected bool
	}

	mockApi := mock.API()

	senderAccAddr := "SenderAccAddress"
	otherAccAddr := "OtherAccAddress"

	testCases := []testCase{
		{
			name: "OK",
			info: stdTypes.MessageInfo{Sender: senderAccAddr},
			msg: MsgInstantiate{
				Params: Params{
					OwnerAddr:      senderAccAddr,
					NewVotingCost:  stdTypes.NewCoinFromUint64(100, "uatom").String(),
					VoteCost:       stdTypes.NewCoinFromUint64(100, "uatom").String(),
					IBCSendTimeout: 100,
				},
			},
		},
		{
			name: "Fail: OwnerAddr: mismatch",
			info: stdTypes.MessageInfo{Sender: otherAccAddr},
			msg: MsgInstantiate{
				Params: Params{
					OwnerAddr:      senderAccAddr,
					NewVotingCost:  stdTypes.NewCoinFromUint64(100, "uatom").String(),
					VoteCost:       stdTypes.NewCoinFromUint64(100, "uatom").String(),
					IBCSendTimeout: 100,
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: NewVotingCost: invalid denom",
			info: stdTypes.MessageInfo{Sender: senderAccAddr},
			msg: MsgInstantiate{
				Params: Params{
					OwnerAddr:      senderAccAddr,
					NewVotingCost:  stdTypes.NewCoinFromUint64(100, "#uatom").String(),
					VoteCost:       stdTypes.NewCoinFromUint64(100, "uatom").String(),
					IBCSendTimeout: 100,
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: VoteCost: invalid denom",
			info: stdTypes.MessageInfo{Sender: senderAccAddr},
			msg: MsgInstantiate{
				Params: Params{
					OwnerAddr:      senderAccAddr,
					NewVotingCost:  stdTypes.NewCoinFromUint64(100, "uatom").String(),
					VoteCost:       stdTypes.NewCoinFromUint64(100, "#uatom").String(),
					IBCSendTimeout: 100,
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: IBCSendTimeout: 0",
			info: stdTypes.MessageInfo{Sender: senderAccAddr},
			msg: MsgInstantiate{
				Params: Params{
					OwnerAddr:      senderAccAddr,
					NewVotingCost:  stdTypes.NewCoinFromUint64(100, "uatom").String(),
					VoteCost:       stdTypes.NewCoinFromUint64(100, "uatom").String(),
					IBCSendTimeout: 0,
				},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errExpected {
				_, err := tc.msg.Params.ValidateAndConvert(mockApi, tc.info)
				assert.Error(t, err)
				return
			}

			_, err := tc.msg.Params.ValidateAndConvert(mockApi, tc.info)
			assert.NoError(t, err)
		})
	}
}
