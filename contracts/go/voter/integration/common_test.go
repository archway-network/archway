package integration

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/CosmWasm/cosmwasm-go/std/math"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	"github.com/CosmWasm/cosmwasm-go/systest"
	mocks "github.com/CosmWasm/wasmvm/api"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	"github.com/stretchr/testify/suite"

	"github.com/archway-network/voter/src/pkg"
	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

const (
	ContractWasmFileName = "code.wasm"
	ValidAddr            = "01234567890abcdefghijklmnopqrstu"
)

var EmptyStruct = struct{}{}

type ContractTestSuite struct {
	suite.Suite

	instance systest.Instance

	creatorAddr string
	genFunds    []stdTypes.Coin
	genParams   types.Params
}

func (s *ContractTestSuite) SetupTest() {
	contractPath := filepath.Join("..", ContractWasmFileName)
	creatorAddr := ValidAddr
	contractFundsCoin := stdTypes.NewCoinFromUint64(1200, "uatom")

	// Load
	instance := systest.NewInstance(s.T(),
		contractPath,
		15_000_000_000_000,
		[]wasmVmTypes.Coin{contractFundsCoin.ToWasmVMCoin()},
	)

	params := types.Params{
		OwnerAddr: creatorAddr,
		NewVotingCost: stdTypes.Coin{
			Denom:  "uatom",
			Amount: math.NewUint128FromUint64(100),
		}.String(),
		VoteCost: stdTypes.Coin{
			Denom:  "uatom",
			Amount: math.NewUint128FromUint64(10),
		}.String(),
		IBCSendTimeout: 10000000000,
	}

	env := mocks.MockEnv()
	info := mocks.MockInfo(creatorAddr, nil)
	msg := types.MsgInstantiate{
		Params: params,
	}

	// Instantiate
	res, _, err := instance.Instantiate(env, info, msg)
	s.Require().NoError(err)

	// Verify response
	s.Require().NotNil(res)
	s.Assert().Empty(res.Messages)
	s.Assert().Empty(res.Attributes)
	s.Assert().Empty(res.Events)

	// Setup
	s.instance = instance
	s.creatorAddr = creatorAddr
	s.genFunds = []stdTypes.Coin{contractFundsCoin}
	s.genParams = msg.Params
}

func (s *ContractTestSuite) ParamsNewVotingCoin() stdTypes.Coin {
	coin, err := pkg.ParseCoinFromString(s.genParams.NewVotingCost)
	s.Require().NoError(err)

	return coin
}

func (s *ContractTestSuite) ParamsVoteCoin() stdTypes.Coin {
	coin, err := pkg.ParseCoinFromString(s.genParams.VoteCost)
	s.Require().NoError(err)

	return coin
}

func (s *ContractTestSuite) AddVoting(env wasmVmTypes.Env, creatorAddr, name string, dur uint64, opts ...string) uint64 {
	info := mocks.MockInfo(creatorAddr, []wasmVmTypes.Coin{s.ParamsNewVotingCoin().ToWasmVMCoin()})
	msg := types.MsgExecute{
		NewVoting: &types.NewVotingRequest{
			Name:        name,
			VoteOptions: opts,
			Duration:    dur,
		},
	}

	res, _, err := s.instance.Execute(env, info, msg)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Empty(res.Messages)
	s.Require().Empty(res.Attributes)

	// Verify data
	var resp types.NewVotingResponse
	s.Require().NoError(resp.UnmarshalJSON(res.Data))

	// Verify events
	s.Require().Len(res.Events, 1)
	event := res.Events[0]
	s.Require().Equal(types.EventTypeNewVoting, event.Type)
	s.Require().Len(event.Attributes, 2)

	s.Require().Equal(types.EventAttrKeySender, event.Attributes[0].Key)
	s.Require().Equal(creatorAddr, event.Attributes[0].Value)

	s.Require().Equal(types.EventAttrKeyVotingID, event.Attributes[1].Key)
	s.Require().Equal(strconv.FormatUint(resp.VotingID, 10), event.Attributes[1].Value)

	return resp.VotingID
}

func (s *ContractTestSuite) Vote(env wasmVmTypes.Env, voterAddr string, votingID uint64, opt, vote string) {
	info := mocks.MockInfo(voterAddr, []wasmVmTypes.Coin{s.ParamsVoteCoin().ToWasmVMCoin()})
	msg := types.MsgExecute{
		Vote: &types.VoteRequest{
			ID:     votingID,
			Option: opt,
			Vote:   vote,
		},
	}

	res, _, err := s.instance.Execute(env, info, msg)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Empty(res.Data)
	s.Require().Empty(res.Messages)
	s.Require().Empty(res.Attributes)

	// Verify events
	s.Require().Len(res.Events, 1)
	event := res.Events[0]
	s.Require().Equal(types.EventTypeVote, event.Type)
	s.Require().Len(event.Attributes, 4)

	s.Require().Equal(types.EventAttrKeySender, event.Attributes[0].Key)
	s.Require().Equal(voterAddr, event.Attributes[0].Value)

	s.Require().Equal(types.EventAttrKeyVotingID, event.Attributes[1].Key)
	s.Require().Equal(strconv.FormatUint(votingID, 10), event.Attributes[1].Value)

	s.Require().Equal(types.EventAttrKeyVoteOption, event.Attributes[2].Key)
	s.Require().Equal(opt, event.Attributes[2].Value)

	s.Require().Equal(types.EventAttrKeyVoteDecision, event.Attributes[3].Key)
	s.Require().Equal(vote, event.Attributes[3].Value)
}

func (s *ContractTestSuite) IBCVote(env wasmVmTypes.Env, voterAddr string, votingID uint64, opt, vote, channelID string) types.MsgIBC {
	info := mocks.MockInfo(voterAddr, nil)
	msg := types.MsgExecute{
		SendIBCVote: &types.SendIBCVoteRequest{
			VoteRequest: types.VoteRequest{
				ID:     votingID,
				Option: opt,
				Vote:   vote,
			},
			ChannelID: channelID,
		},
	}

	res, _, err := s.instance.Execute(env, info, msg)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Len(res.Messages, 1)

	// Validate events
	s.Require().Len(res.Events, 1)
	s.Assert().Equal(types.EventTypeIBCVoteSent, res.Events[0].Type)
	s.Require().Len(res.Events[0].Attributes, 4)
	s.Assert().Equal(types.EventAttrKeySender, res.Events[0].Attributes[0].Key)
	s.Assert().Equal(voterAddr, res.Events[0].Attributes[0].Value)
	s.Assert().Equal(types.EventAttrKeyVotingID, res.Events[0].Attributes[1].Key)
	s.Assert().Equal(strconv.FormatUint(votingID, 10), res.Events[0].Attributes[1].Value)
	s.Assert().Equal(types.EventAttrKeyVoteOption, res.Events[0].Attributes[2].Key)
	s.Assert().Equal(opt, res.Events[0].Attributes[2].Value)
	s.Assert().Equal(types.EventAttrKeyVoteDecision, res.Events[0].Attributes[3].Key)
	s.Assert().Equal(vote, res.Events[0].Attributes[3].Value)

	// Build and return IBC message
	return types.MsgIBC{
		Vote: &types.IBCVoteRequest{
			VoteRequest: types.VoteRequest{
				ID:     votingID,
				Option: opt,
				Vote:   vote,
			},
			From: voterAddr,
		},
	}
}

func (s *ContractTestSuite) GetIBCStats(env wasmVmTypes.Env, senderAddr string, votingID uint64) state.IBCStats {
	query := types.MsgQuery{
		IBCStats: &types.QueryIBCStatsRequest{
			From: senderAddr,
		},
	}

	respBz, _, err := s.instance.Query(env, query)
	s.Require().NoError(err)

	var resp types.QueryIBCStatsResponse
	s.Require().NoError(resp.UnmarshalJSON(respBz))

	for _, ibcStats := range resp.Stats {
		if ibcStats.VotingID == votingID {
			return ibcStats
		}
	}
	s.Failf("No ibcStats found", "senderAddr: %s, votingID: %d", senderAddr, votingID)

	return state.IBCStats{}
}

func TestContract(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}
