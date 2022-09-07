package integration

import (
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	mocks "github.com/CosmWasm/wasmvm/api"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"

	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

func (s *ContractTestSuite) TestQueryParams() {
	env := mocks.MockEnv()

	query := types.MsgQuery{Params: &EmptyStruct}
	res, _, err := s.instance.Query(env, query)
	s.Require().NoError(err)
	s.Require().NotNil(res)

	var resp types.QueryParamsResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	s.Assert().Equal(s.genParams.OwnerAddr, resp.OwnerAddr)
	s.Assert().Equal(s.genParams.NewVotingCost, resp.NewVotingCost)
	s.Assert().Equal(s.genParams.VoteCost, resp.VoteCost)
	s.Assert().Equal(s.genParams.IBCSendTimeout, resp.IBCSendTimeout)
}

func (s *ContractTestSuite) TestQueryVoting() {
	env := mocks.MockEnv()

	s.Run("Fail: non-existing", func() {
		query := types.MsgQuery{
			Voting: &types.QueryVotingRequest{ID: 0},
		}

		_, _, err := s.instance.Query(env, query)
		s.Assert().Error(err)
	})

	s.Run("OK", func() {
		// Add voting
		votingID := s.AddVoting(env, s.creatorAddr, "Test", 1000, "a", "b")

		query := types.MsgQuery{
			Voting: &types.QueryVotingRequest{ID: votingID},
		}

		respBz, _, err := s.instance.Query(env, query)
		s.Require().NoError(err)

		s.Require().NotNil(respBz)
		var resp state.Voting
		s.Require().NoError(resp.UnmarshalJSON(respBz))

		s.Assert().Equal(votingID, resp.ID)
		s.Assert().Equal(s.creatorAddr, resp.CreatorAddr)
		s.Assert().Equal("Test", resp.Name)
		s.Assert().Equal(env.Block.Time, resp.StartTime)
		s.Assert().Equal(env.Block.Time+1000, resp.EndTime)

		s.Require().Len(resp.Tallies, 2)
		s.Assert().Equal("a", resp.Tallies[0].Option)
		s.Assert().Empty(resp.Tallies[0].YesAddrs)
		s.Assert().Empty(resp.Tallies[0].NoAddrs)
		s.Assert().Equal("b", resp.Tallies[1].Option)
		s.Assert().Empty(resp.Tallies[1].YesAddrs)
		s.Assert().Empty(resp.Tallies[1].NoAddrs)
	})
}

func (s *ContractTestSuite) TestQueryTally() {
	voter1Addr, voter2Addr := "Voter1Addr", "Voter2Addr"
	env := mocks.MockEnv()

	s.Run("Fail: non-existing", func() {
		query := types.MsgQuery{
			Tally: &types.QueryTallyRequest{ID: 0},
		}

		_, _, err := s.instance.Query(env, query)
		s.Assert().Error(err)
	})

	s.Run("OK", func() {
		// Add voting and votes
		votingID := s.AddVoting(env, s.creatorAddr, "Test", 1000, "a", "b")
		env.Block.Time++
		s.Vote(env, voter1Addr, votingID, "a", "yes")
		s.Vote(env, voter2Addr, votingID, "b", "no")

		query := types.MsgQuery{
			Tally: &types.QueryTallyRequest{ID: votingID},
		}

		respBz, _, err := s.instance.Query(env, query)
		s.Require().NoError(err)
		s.Require().NotNil(respBz)

		var resp types.QueryTallyResponse
		s.Require().NoError(resp.UnmarshalJSON(respBz))

		s.Assert().True(resp.Open)
		s.Require().Len(resp.Votes, 2)

		s.Assert().Equal("a", resp.Votes[0].Option)
		s.Assert().EqualValues(1, resp.Votes[0].TotalYes)
		s.Assert().EqualValues(0, resp.Votes[0].TotalNo)

		s.Assert().Equal("b", resp.Votes[1].Option)
		s.Assert().EqualValues(0, resp.Votes[1].TotalYes)
		s.Assert().EqualValues(1, resp.Votes[1].TotalNo)
	})
}

func (s *ContractTestSuite) TestQueryOpen() {
	env := mocks.MockEnv()
	env.Block.Time = 1

	runQuery := func() []uint64 {
		query := types.MsgQuery{
			Open: &EmptyStruct,
		}

		respBz, _, err := s.instance.Query(env, query)
		s.Require().NoError(err)

		var resp types.QueryOpenResponse
		s.Require().NoError(resp.UnmarshalJSON(respBz))

		return resp.Ids
	}

	s.Run("No votings", func() {
		s.Assert().Len(runQuery(), 0)
	})

	// Add votings with different durations
	votingID1 := s.AddVoting(env, s.creatorAddr, "Test1", 10, "a")
	votingID2 := s.AddVoting(env, s.creatorAddr, "Test2", 20, "a")
	votingID3 := s.AddVoting(env, s.creatorAddr, "Test3", 30, "a")

	s.Run("3 open", func() {
		env.Block.Time++

		idsExpected := []uint64{votingID1, votingID2, votingID3}
		idsReceived := runQuery()
		s.Assert().ElementsMatch(idsExpected, idsReceived)
	})

	s.Run("2 open", func() {
		env.Block.Time = 15

		idsExpected := []uint64{votingID2, votingID3}
		idsReceived := runQuery()
		s.Assert().ElementsMatch(idsExpected, idsReceived)
	})

	s.Run("1 open", func() {
		env.Block.Time = 25

		idsExpected := []uint64{votingID3}
		idsReceived := runQuery()
		s.Assert().ElementsMatch(idsExpected, idsReceived)
	})

	s.Run("0 open (again)", func() {
		env.Block.Time = 35

		idsReceived := runQuery()
		s.Assert().Empty(idsReceived)
	})
}

func (s *ContractTestSuite) TestQueryReleaseStats() {
	env := mocks.MockEnv()
	env.Block.Time = 1

	runQuery := func() state.ReleaseStats {
		query := types.MsgQuery{
			ReleaseStats: &EmptyStruct,
		}

		respBz, _, err := s.instance.Query(env, query)
		s.Require().NoError(err)

		var resp state.ReleaseStats
		s.Require().NoError(resp.UnmarshalJSON(respBz))

		return resp
	}

	s.Run("No releases", func() {
		stats := runQuery()
		s.Assert().EqualValues(0, stats.Count)
		s.Assert().Nil(stats.TotalAmount)
	})

	// Send Release msg and emulate Reply receive
	totalAmtExpected := stdTypes.NewCoinFromUint64(123, "uatom")
	{
		info := mocks.MockInfo(s.creatorAddr, nil)
		releaseMsg := types.MsgExecute{
			Release: &EmptyStruct,
		}

		_, _, err := s.instance.Execute(env, info, releaseMsg)
		s.Require().NoError(err)

		replyMsg := wasmVmTypes.Reply{
			ID: 0,
			Result: wasmVmTypes.SubMsgResult{
				Ok: &wasmVmTypes.SubMsgResponse{
					Events: wasmVmTypes.Events{
						{
							Type: "transfer",
							Attributes: wasmVmTypes.EventAttributes{
								{Key: "amount", Value: totalAmtExpected.String()},
							},
						},
					},
				},
			},
		}

		_, _, err = s.instance.Reply(env, replyMsg)
		s.Require().NoError(err)
	}

	s.Run("1 release", func() {
		stats := runQuery()
		s.Assert().EqualValues(1, stats.Count)
		s.Assert().ElementsMatch([]stdTypes.Coin{totalAmtExpected}, stats.TotalAmount)
	})
}

func (s *ContractTestSuite) TestQueryIBCStats() {
	env := mocks.MockEnv()
	senderAddr1, senderAddr2, senderAddr3 := "SenderAddr1", "SenderAddr2", "SenderAddr3"

	runQuery := func(senderAddr string) types.QueryIBCStatsResponse {
		query := types.MsgQuery{
			IBCStats: &types.QueryIBCStatsRequest{
				From: senderAddr,
			},
		}

		respBz, _, err := s.instance.Query(env, query)
		s.Require().NoError(err)

		var resp types.QueryIBCStatsResponse
		s.Require().NoError(resp.UnmarshalJSON(respBz))

		return resp
	}

	// Add stats
	ibcStats1Exp := types.QueryIBCStatsResponse{
		Stats: []state.IBCStats{
			{
				VotingID:  1,
				From:      senderAddr1,
				Status:    state.IBCPkgSentStatus,
				CreatedAt: env.Block.Time,
			},
			{
				VotingID:  2,
				From:      senderAddr1,
				Status:    state.IBCPkgSentStatus,
				CreatedAt: env.Block.Time,
			},
		},
	}

	ibcStats2Exp := types.QueryIBCStatsResponse{
		Stats: []state.IBCStats{
			{
				VotingID:  3,
				From:      senderAddr2,
				Status:    state.IBCPkgSentStatus,
				CreatedAt: env.Block.Time,
			},
		},
	}

	for _, stats := range ibcStats1Exp.Stats {
		s.IBCVote(env, stats.From, stats.VotingID, "a", "yes", "channel-1")
	}
	for _, stats := range ibcStats2Exp.Stats {
		s.IBCVote(env, stats.From, stats.VotingID, "b", "no", "channel-2")
	}

	s.Run("OK: non-existing", func() {
		resp := runQuery(senderAddr3)
		s.Assert().Empty(resp.Stats)
	})

	s.Run("OK: sender 1", func() {
		resp := runQuery(senderAddr1)
		s.Assert().ElementsMatch(ibcStats1Exp.Stats, resp.Stats)
	})

	s.Run("OK: sender 2", func() {
		resp := runQuery(senderAddr2)
		s.Assert().ElementsMatch(ibcStats2Exp.Stats, resp.Stats)
	})
}
