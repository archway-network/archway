package integration

import (
	mocks "github.com/CosmWasm/wasmvm/api"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"

	"github.com/archway-network/voter/src/types"
)

func (s *ContractTestSuite) TestExecuteNewVoting() {
	env := mocks.MockEnv()

	// Test OK
	s.AddVoting(env, s.creatorAddr, "Test", 100, "a")

	s.Run("Fail: invalid input", func() {
		info := mocks.MockInfo(s.creatorAddr, []wasmVmTypes.Coin{s.ParamsNewVotingCoin().ToWasmVMCoin()})
		msg := types.MsgExecute{
			NewVoting: &types.NewVotingRequest{
				Name:        "Test",
				VoteOptions: nil,
				Duration:    100,
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})

	s.Run("Fail: invalid payment", func() {
		payment := s.ParamsNewVotingCoin()
		payment.Amount = payment.Amount.Sub64(1)

		info := mocks.MockInfo(s.creatorAddr, []wasmVmTypes.Coin{payment.ToWasmVMCoin()})
		msg := types.MsgExecute{
			NewVoting: &types.NewVotingRequest{
				Name:        "Test",
				VoteOptions: []string{"a"},
				Duration:    100,
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})
}

func (s *ContractTestSuite) TestExecuteVote() {
	env := mocks.MockEnv()

	voter1Addr, voter2Addr := "Voter1Addr", "Voter2Addr"

	// Test OK
	votingID := s.AddVoting(env, s.creatorAddr, "Test", 100, "a")
	s.Vote(env, voter1Addr, votingID, "a", "yes")

	s.Run("Fail: invalid input", func() {
		info := mocks.MockInfo(voter2Addr, []wasmVmTypes.Coin{s.ParamsVoteCoin().ToWasmVMCoin()})
		msg := types.MsgExecute{
			Vote: &types.VoteRequest{
				ID:     votingID,
				Option: "",
				Vote:   "yes",
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})

	s.Run("Fail: invalid payment", func() {
		payment := s.ParamsVoteCoin()
		payment.Amount = payment.Amount.Sub64(1)

		info := mocks.MockInfo(voter2Addr, []wasmVmTypes.Coin{payment.ToWasmVMCoin()})
		msg := types.MsgExecute{
			Vote: &types.VoteRequest{
				ID:     votingID,
				Option: "a",
				Vote:   "yes",
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})

	s.Run("Fail: non-existing voting", func() {
		info := mocks.MockInfo(voter2Addr, []wasmVmTypes.Coin{s.ParamsVoteCoin().ToWasmVMCoin()})
		msg := types.MsgExecute{
			Vote: &types.VoteRequest{
				ID:     votingID + 1,
				Option: "a",
				Vote:   "yes",
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})

	s.Run("Fail: already voted", func() {
		info := mocks.MockInfo(voter1Addr, []wasmVmTypes.Coin{s.ParamsVoteCoin().ToWasmVMCoin()})
		msg := types.MsgExecute{
			Vote: &types.VoteRequest{
				ID:     votingID,
				Option: "a",
				Vote:   "no",
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})

	s.Run("Fail: voting is closed", func() {
		env := mocks.MockEnv()
		env.Block.Time += 200
		info := mocks.MockInfo(voter2Addr, []wasmVmTypes.Coin{s.ParamsVoteCoin().ToWasmVMCoin()})
		msg := types.MsgExecute{
			Vote: &types.VoteRequest{
				ID:     votingID,
				Option: "a",
				Vote:   "yes",
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})

	s.Run("Fail: non-existing option", func() {
		info := mocks.MockInfo(voter2Addr, []wasmVmTypes.Coin{s.ParamsVoteCoin().ToWasmVMCoin()})
		msg := types.MsgExecute{
			Vote: &types.VoteRequest{
				ID:     votingID,
				Option: "c",
				Vote:   "no",
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})
}

func (s *ContractTestSuite) TestExecuteRelease() {
	env := mocks.MockEnv()

	voter1Addr, voter2Addr := "Voter1Addr", "Voter2Addr"

	// Add voting and votes (1000 + 2 * 100 of raised funds)
	votingID := s.AddVoting(env, voter1Addr, "Test", 100, "a")
	s.Vote(env, voter1Addr, votingID, "a", "yes")
	s.Vote(env, voter2Addr, votingID, "a", "no")

	s.Run("Fail: unauthorized", func() {
		info := mocks.MockInfo(voter2Addr, nil)
		msg := types.MsgExecute{
			Release: &EmptyStruct,
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Assert().Error(err)
	})

	s.Run("OK", func() {
		info := mocks.MockInfo(s.creatorAddr, nil)
		msg := types.MsgExecute{
			Release: &EmptyStruct,
		}

		res, _, err := s.instance.Execute(env, info, msg)
		s.Require().NoError(err)
		s.Require().NotNil(res)

		var resp types.ReleaseResponse
		s.Require().NoError(resp.UnmarshalJSON(res.Data))

		s.Assert().ElementsMatch(s.genFunds, resp.ReleasedAmount)
	})
}

func (s *ContractTestSuite) TestExecuteSendIBCVote() {
	env := mocks.MockEnv()
	senderAddr := "SenderAddr"

	s.Run("OK", func() {
		s.IBCVote(env, senderAddr, 1, "a", "yes", "channel-1")
	})

	s.Run("Fail: invalid input", func() {
		info := mocks.MockInfo(senderAddr, nil)
		msg := types.MsgExecute{
			SendIBCVote: &types.SendIBCVoteRequest{
				VoteRequest: types.VoteRequest{
					ID:     1,
					Option: "a",
					Vote:   "yes",
				},
				ChannelID: "invalid",
			},
		}

		_, _, err := s.instance.Execute(env, info, msg)
		s.Require().Error(err)
	})
}
