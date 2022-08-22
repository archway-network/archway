package integration

import (
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	mocks "github.com/CosmWasm/wasmvm/api"

	"github.com/archway-network/voter/src/types"
)

func (s *ContractTestSuite) TestSudoChangeAddVotingCost() {
	env := mocks.MockEnv()
	expectedCoin := s.ParamsNewVotingCoin()
	expectedCoin.Amount = expectedCoin.Amount.Sub64(1)

	s.Run("OK", func() {
		msg := types.MsgSudo{
			ChangeNewVotingCost: &types.ChangeCostRequest{
				NewCost: expectedCoin,
			},
		}

		res, _, err := s.instance.Sudo(env, msg)
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Assert().Empty(res.Data)
		s.Assert().Empty(res.Messages)
		s.Assert().Empty(res.Attributes)

		// Verify events
		s.Require().Len(res.Events, 1)
		rcvEvent := res.Events[0]
		s.Assert().Equal(types.EventTypeNewVotingCostChanged, rcvEvent.Type)

		s.Require().Len(rcvEvent.Attributes, 2)
		s.Assert().Equal(types.EventAttrKeyOldCost, rcvEvent.Attributes[0].Key)
		s.Assert().Equal(s.genParams.NewVotingCost, rcvEvent.Attributes[0].Value)
		s.Assert().Equal(types.EventAttrKeyNewCost, rcvEvent.Attributes[1].Key)
		s.Assert().Equal(expectedCoin.String(), rcvEvent.Attributes[1].Value)

		// Verify state change
		query := types.MsgQuery{Params: &EmptyStruct}
		paramsBz, _, err := s.instance.Query(env, query)
		s.Require().NoError(err)
		s.Require().NotNil(paramsBz)

		var params types.QueryParamsResponse
		s.Require().NoError(params.UnmarshalJSON(paramsBz))
		s.Assert().Equal(expectedCoin.String(), params.NewVotingCost)
	})

	s.Run("Fail: invalid input", func() {
		msg := types.MsgSudo{
			ChangeNewVotingCost: &types.ChangeCostRequest{
				NewCost: stdTypes.Coin{
					Denom:  "1uatom",
					Amount: expectedCoin.Amount,
				},
			},
		}

		_, _, err := s.instance.Sudo(env, msg)
		s.Require().Error(err)
	})
}

func (s *ContractTestSuite) TestSudoChangeVoteCost() {
	env := mocks.MockEnv()
	expectedCoin := s.ParamsVoteCoin()
	expectedCoin.Amount = expectedCoin.Amount.Sub64(1)

	s.Run("OK", func() {
		msg := types.MsgSudo{
			ChangeVoteCost: &types.ChangeCostRequest{
				NewCost: expectedCoin,
			},
		}

		res, _, err := s.instance.Sudo(env, msg)
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Assert().Empty(res.Data)
		s.Assert().Empty(res.Messages)
		s.Assert().Empty(res.Attributes)

		// Verify events
		s.Require().Len(res.Events, 1)
		rcvEvent := res.Events[0]
		s.Assert().Equal(types.EventTypeVoteCostChanged, rcvEvent.Type)

		s.Require().Len(rcvEvent.Attributes, 2)
		s.Assert().Equal(types.EventAttrKeyOldCost, rcvEvent.Attributes[0].Key)
		s.Assert().Equal(s.genParams.VoteCost, rcvEvent.Attributes[0].Value)
		s.Assert().Equal(types.EventAttrKeyNewCost, rcvEvent.Attributes[1].Key)
		s.Assert().Equal(expectedCoin.String(), rcvEvent.Attributes[1].Value)

		// Verify state change
		query := types.MsgQuery{Params: &EmptyStruct}
		paramsBz, _, err := s.instance.Query(env, query)
		s.Require().NoError(err)
		s.Require().NotNil(paramsBz)

		var params types.QueryParamsResponse
		s.Require().NoError(params.UnmarshalJSON(paramsBz))
		s.Assert().Equal(expectedCoin.String(), params.VoteCost)
	})

	s.Run("Fail: invalid input", func() {
		msg := types.MsgSudo{
			ChangeVoteCost: &types.ChangeCostRequest{
				NewCost: stdTypes.Coin{
					Denom:  "1uatom",
					Amount: expectedCoin.Amount,
				},
			},
		}

		_, _, err := s.instance.Sudo(env, msg)
		s.Require().Error(err)
	})
}
