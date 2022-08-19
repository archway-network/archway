package integration

import (
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	mocks "github.com/CosmWasm/wasmvm/api"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"

	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

func (s *ContractTestSuite) TestReplyBankSend() {
	env := mocks.MockEnv()

	// Add voting
	s.AddVoting(env, s.creatorAddr, "Test", 100, "a")

	s.Run("Fail: no reply ID found", func() {
		msg := wasmVmTypes.Reply{
			ID: 0,
		}

		_, _, err := s.instance.Reply(env, msg)
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), "not found")
	})

	// Release funds (replyID 0 is created here)
	{
		info := mocks.MockInfo(s.creatorAddr, nil)
		msg := types.MsgExecute{
			Release: &EmptyStruct,
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Require().NoError(err)
	}

	s.Run("Fail: invalid reply: with error", func() {
		msg := wasmVmTypes.Reply{
			ID: 0,
			Result: wasmVmTypes.SubMsgResult{
				Err: "some error",
			},
		}

		_, _, err := s.instance.Reply(env, msg)
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), "x/bank reply: error received")
	})

	s.Run("Fail: invalid reply: invalid message type received (wrong events)", func() {
		msg := wasmVmTypes.Reply{
			ID: 0,
			Result: wasmVmTypes.SubMsgResult{
				Ok: &wasmVmTypes.SubMsgResponse{},
			},
		}

		_, _, err := s.instance.Reply(env, msg)
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), "x/bank reply: transfer.amount attribute: not found")
	})

	s.Run("OK", func() {
		releaseAmtExpected := stdTypes.NewCoinFromUint64(1000, "uatom")

		msg := wasmVmTypes.Reply{
			ID: 0,
			Result: wasmVmTypes.SubMsgResult{
				Ok: &wasmVmTypes.SubMsgResponse{
					Events: wasmVmTypes.Events{
						{
							Type: "transfer",
							Attributes: wasmVmTypes.EventAttributes{
								{Key: "amount", Value: releaseAmtExpected.String()},
							},
						},
					},
				},
			},
		}

		_, _, err := s.instance.Reply(env, msg)
		s.Assert().NoError(err)

		// Verify stats changed
		{
			query := types.MsgQuery{
				ReleaseStats: &EmptyStruct,
			}

			respBz, _, err := s.instance.Query(env, query)
			s.Require().NoError(err)
			s.Require().NotNil(respBz)

			var stats state.ReleaseStats
			s.Require().NoError(stats.UnmarshalJSON(respBz))

			s.Assert().EqualValues(1, stats.Count)
			s.Assert().ElementsMatch([]stdTypes.Coin{releaseAmtExpected}, stats.TotalAmount)
		}
	})
}
