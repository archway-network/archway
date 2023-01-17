package e2e

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	cwMath "github.com/CosmWasm/cosmwasm-go/std/math"
	cwSdkTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	channelTypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	voterCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
	voterState "github.com/archway-network/voter/src/state"
	voterTypes "github.com/archway-network/voter/src/types"

	"github.com/archway-network/archway/wasmbinding/pkg"
	rewardsWbTypes "github.com/archway-network/archway/wasmbinding/rewards/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// Voter contract related helpers.

// VoterUploadAndInstantiate creates a new Voter contract.
func (s *E2ETestSuite) VoterUploadAndInstantiate(chain *e2eTesting.TestChain, acc e2eTesting.Account) (contractAddr sdk.AccAddress) {
	codeID := chain.UploadContract(acc, VoterWasmPath, wasmdTypes.DefaultUploadAccess)

	instMsg := voterTypes.MsgInstantiate{
		Params: s.VoterDefaultParams(acc),
	}

	contractAddr, _ = chain.InstantiateContract(acc, codeID, acc.Address.String(), "voter", nil, instMsg)

	return
}

// VoterDefaultParams returns default parameters for the contract (used by VoterUploadAndInstantiate).
func (s *E2ETestSuite) VoterDefaultParams(acc e2eTesting.Account) voterTypes.Params {
	return voterTypes.Params{
		OwnerAddr: acc.Address.String(),
		NewVotingCost: cwSdkTypes.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: cwMath.NewUint128FromUint64(DefNewVotingCostAmt),
		}.String(),
		VoteCost: cwSdkTypes.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: cwMath.NewUint128FromUint64(DefNewVoteCostAmt),
		}.String(),
		IBCSendTimeout: 30000000000, // 30s‰
	}
}

// VoterNewVoting creates a new voting.
func (s *E2ETestSuite) VoterNewVoting(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, acc e2eTesting.Account, votingName string, voteOps []string, voteDur time.Duration) (votingID uint64) {
	req := voterTypes.MsgExecute{
		NewVoting: &voterTypes.NewVotingRequest{
			Name:        votingName,
			VoteOptions: voteOps,
			Duration:    uint64(voteDur),
		},
	}
	reqBz, err := req.MarshalJSON()
	s.Require().NoError(err)

	msg := wasmdTypes.MsgExecuteContract{
		Sender:   acc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      reqBz,
		Funds: sdk.NewCoins(sdk.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: sdk.NewIntFromUint64(DefNewVotingCostAmt),
		}),
	}

	_, res, _, _ := chain.SendMsgs(acc, true, []sdk.Msg{&msg})

	txRes := chain.ParseSDKResultData(res)
	s.Require().Len(txRes.Data, 1)

	var executeRes wasmdTypes.MsgExecuteContractResponse
	s.Require().NoError(executeRes.Unmarshal(txRes.Data[0].Data))

	var resp voterTypes.NewVotingResponse
	s.Require().NoError(resp.UnmarshalJSON(executeRes.Data))

	votingID = resp.VotingID

	return
}

// VoterVote adds a vote for an existing voting.
func (s *E2ETestSuite) VoterVote(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, acc e2eTesting.Account, votingID uint64, voteOpt string, voteYes bool) {
	vote := "yes"
	if !voteYes {
		vote = "no"
	}

	req := voterTypes.MsgExecute{
		Vote: &voterTypes.VoteRequest{
			ID:     votingID,
			Option: voteOpt,
			Vote:   vote,
		},
	}
	reqBz, err := req.MarshalJSON()
	s.Require().NoError(err)

	msg := wasmdTypes.MsgExecuteContract{
		Sender:   acc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      reqBz,
		Funds: sdk.NewCoins(sdk.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: sdk.NewIntFromUint64(DefNewVoteCostAmt),
		}),
	}

	chain.SendMsgs(acc, true, []sdk.Msg{&msg})
}

// VoterIBCVote adds a vote for an existing voting over IBC.
func (s *E2ETestSuite) VoterIBCVote(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, acc e2eTesting.Account, votingID uint64, voteOpt string, voteYes bool, channelID string) channelTypes.Packet {
	vote := "yes"
	if !voteYes {
		vote = "no"
	}

	req := voterTypes.MsgExecute{
		SendIBCVote: &voterTypes.SendIBCVoteRequest{
			VoteRequest: voterTypes.VoteRequest{
				ID:     votingID,
				Option: voteOpt,
				Vote:   vote,
			},
			ChannelID: channelID,
		},
	}
	reqBz, err := req.MarshalJSON()
	s.Require().NoError(err)

	msg := wasmdTypes.MsgExecuteContract{
		Sender:   acc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      reqBz,
		Funds: sdk.NewCoins(sdk.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: sdk.NewIntFromUint64(DefNewVoteCostAmt),
		}),
	}

	_, res, _, _ := chain.SendMsgs(acc, true, []sdk.Msg{&msg})

	// Assemble the IBC packet from the response
	var packet channelTypes.Packet

	pSeqRaw := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeySequence)
	s.Require().NotEmpty(pSeqRaw)
	packet.Sequence, err = strconv.ParseUint(pSeqRaw, 10, 64)
	s.Require().NoError(err)

	pSrcPort := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeySrcPort)
	s.Require().NotEmpty(pSrcPort)
	packet.SourcePort = pSrcPort

	pSrcChannel := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeySrcChannel)
	s.Require().NotEmpty(pSrcChannel)
	packet.SourceChannel = pSrcChannel

	pDstPort := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeyDstPort)
	s.Require().NotEmpty(pDstPort)
	packet.DestinationPort = pDstPort

	pDstChannel := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeyDstChannel)
	s.Require().NotEmpty(pDstChannel)
	packet.DestinationChannel = pDstChannel

	pData := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeyDataHex)
	s.Require().NotEmpty(pData)
	packet.Data, err = hex.DecodeString(pData)
	s.Require().NoError(err)

	pTimeoutHeightRaw := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeyTimeoutHeight)
	s.Require().NotEmpty(pTimeoutHeightRaw)
	pTimeoutHeightSplit := strings.Split(pTimeoutHeightRaw, "-")
	s.Require().Len(pTimeoutHeightSplit, 2)
	packet.TimeoutHeight.RevisionNumber, err = strconv.ParseUint(pTimeoutHeightSplit[0], 10, 64)
	s.Require().NoError(err)
	packet.TimeoutHeight.RevisionHeight, err = strconv.ParseUint(pTimeoutHeightSplit[1], 10, 64)
	s.Require().NoError(err)

	pTimeoutTSRaw := e2eTesting.GetStringEventAttribute(res.Events, channelTypes.EventTypeSendPacket, channelTypes.AttributeKeyTimeoutTimestamp)
	s.Require().NotEmpty(pTimeoutTSRaw)
	packet.TimeoutTimestamp, err = strconv.ParseUint(pTimeoutTSRaw, 10, 64)
	s.Require().NoError(err)

	return packet
}

// VoterRelease releases contract funds to the owner.
func (s *E2ETestSuite) VoterRelease(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, acc e2eTesting.Account) (releasedCoins sdk.Coins) {
	req := voterTypes.MsgExecute{
		Release: &struct{}{},
	}
	reqBz, err := req.MarshalJSON()
	s.Require().NoError(err)

	msg := wasmdTypes.MsgExecuteContract{
		Sender:   acc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      reqBz,
	}

	_, res, _, _ := chain.SendMsgs(acc, true, []sdk.Msg{&msg})

	txRes := chain.ParseSDKResultData(res)
	s.Require().Len(txRes.Data, 1)

	var executeRes wasmdTypes.MsgExecuteContractResponse
	s.Require().NoError(executeRes.Unmarshal(txRes.Data[0].Data))

	var resp voterTypes.ReleaseResponse
	s.Require().NoError(resp.UnmarshalJSON(executeRes.Data))

	releasedCoins = s.CosmWasmCoinsToSDK(resp.ReleasedAmount...)

	return
}

// VoterGetVoting returns the contract parameters.
func (s *E2ETestSuite) VoterGetParams(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress) voterTypes.Params {
	req := voterTypes.MsgQuery{
		Params: &struct{}{},
	}

	res, _ := chain.SmartQueryContract(contractAddr, true, req)

	var resp voterTypes.QueryParamsResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp.Params
}

// VoterGetVoting returns a voting.
func (s *E2ETestSuite) VoterGetVoting(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, votingID uint64) voterState.Voting {
	req := voterTypes.MsgQuery{
		Voting: &voterTypes.QueryVotingRequest{
			ID: votingID,
		},
	}

	res, _ := chain.SmartQueryContract(contractAddr, true, req)

	var resp voterTypes.QueryVotingResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp.Voting
}

// VoterGetTally returns the current voting state.
func (s *E2ETestSuite) VoterGetTally(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, votingID uint64) voterTypes.QueryTallyResponse {
	req := voterTypes.MsgQuery{
		Tally: &voterTypes.QueryTallyRequest{
			ID: votingID,
		},
	}

	res, _ := chain.SmartQueryContract(contractAddr, true, req)

	var resp voterTypes.QueryTallyResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp
}

// VoterGetReleaseStats returns the release stats (updated via Reply endpoint).
func (s *E2ETestSuite) VoterGetReleaseStats(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress) voterTypes.QueryReleaseStatsResponse {
	req := voterTypes.MsgQuery{
		ReleaseStats: &struct{}{},
	}

	res, _ := chain.SmartQueryContract(contractAddr, true, req)

	var resp voterTypes.QueryReleaseStatsResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp
}

// VoterGetIBCStats returns send IBC packages stats.
func (s *E2ETestSuite) VoterGetIBCStats(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, senderAddr e2eTesting.Account) []voterState.IBCStats {
	req := voterTypes.MsgQuery{
		IBCStats: &voterTypes.QueryIBCStatsRequest{
			From: senderAddr.Address.String(),
		},
	}

	res, _ := chain.SmartQueryContract(contractAddr, true, req)

	var resp voterTypes.QueryIBCStatsResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp.Stats
}

// VoterGetWithdrawStats returns the withdraw stats (updated via Reply endpoint).
func (s *E2ETestSuite) VoterGetWithdrawStats(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress) voterTypes.QueryWithdrawStatsResponse {
	req := voterTypes.MsgQuery{
		WithdrawStats: &struct{}{},
	}

	res, _ := chain.SmartQueryContract(contractAddr, true, req)

	var resp voterTypes.QueryWithdrawStatsResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp
}

// VoterGetMetadata returns the contract metadata queried via Custom querier plugin.
func (s *E2ETestSuite) VoterGetMetadata(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, useStargate, expPass bool) voterCustomTypes.ContractMetadataResponse {
	req := voterTypes.MsgQuery{
		CustomMetadata: &voterTypes.CustomMetadataRequest{
			UseStargateQuery: useStargate,
		},
	}

	res, _ := chain.SmartQueryContract(contractAddr, expPass, req)
	if !expPass {
		return voterCustomTypes.ContractMetadataResponse{}
	}

	var resp voterTypes.CustomMetadataResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp.ContractMetadataResponse
}

// VoterSendCustomMsg sends the Custom plugin message (should be serialized by the caller).
func (s *E2ETestSuite) VoterSendCustomMsg(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, acc e2eTesting.Account, customMsg []byte, expPass bool) error {
	req := voterTypes.MsgExecute{
		CustomCustom: customMsg,
	}
	reqBz, err := req.MarshalJSON()
	s.Require().NoError(err)

	msg := wasmdTypes.MsgExecuteContract{
		Sender:   acc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      reqBz,
	}

	_, _, _, err = chain.SendMsgs(acc, expPass, []sdk.Msg{&msg})
	if !expPass {
		s.Require().Error(err)
		return err
	}
	s.Require().NoError(err)

	return nil
}

// VoterUpdateMetadata sends the contract metadata update request via Custom message plugin.
func (s *E2ETestSuite) VoterUpdateMetadata(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, acc e2eTesting.Account, metaReq voterCustomTypes.UpdateContractMetadataRequest, expPass bool) error {
	req := voterTypes.MsgExecute{
		CustomUpdateMetadata: &metaReq,
	}
	reqBz, err := req.MarshalJSON()
	s.Require().NoError(err)

	msg := wasmdTypes.MsgExecuteContract{
		Sender:   acc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      reqBz,
	}

	_, _, _, err = chain.SendMsgs(acc, expPass, []sdk.Msg{&msg})

	return err
}

// VoterGetRewardsRecords returns the current contract rewards records (for the contractAddress as a rewardsAddress) paginated via Custom querier plugin.
func (s *E2ETestSuite) VoterGetRewardsRecords(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, pageReq *query.PageRequest, expPass bool) ([]rewardsTypes.RewardsRecord, query.PageResponse, int, error) {
	req := voterTypes.MsgQuery{
		CustomRewardsRecords: &voterTypes.CustomRewardsRecordsRequest{},
	}
	if pageReq != nil {
		r := pkg.NewPageRequestFromSDK(*pageReq)
		req.CustomRewardsRecords.Pagination = &voterCustomTypes.PageRequest{
			Key:        r.Key,
			Offset:     r.Offset,
			Limit:      r.Limit,
			CountTotal: r.CountTotal,
			Reverse:    r.Reverse,
		}
	}

	res, err := chain.SmartQueryContract(contractAddr, expPass, req)
	if !expPass {
		s.Require().Error(err)
		return nil, query.PageResponse{}, 0, err
	}
	s.Require().NoError(err)

	var resp rewardsWbTypes.RewardsRecordsResponse
	s.Require().NoError(json.Unmarshal(res, &resp))

	records := make([]rewardsTypes.RewardsRecord, 0, len(resp.Records))
	for _, record := range resp.Records {
		r, err := record.ToSDK()
		s.Require().NoError(err)

		records = append(records, r)
	}

	return records, resp.Pagination.ToSDK(), len(res), nil
}

// VoterWithdrawRewards sends the contract rewards withdrawal request via Custom message plugin.
func (s *E2ETestSuite) VoterWithdrawRewards(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, acc e2eTesting.Account, recordsLimit *uint64, recordIDs []uint64, expPass bool) error {
	req := voterTypes.MsgExecute{
		CustomWithdrawRewards: &voterCustomTypes.WithdrawRewardsRequest{
			RecordsLimit: recordsLimit,
			RecordIds:    recordIDs,
		},
	}
	reqBz, err := req.MarshalJSON()
	s.Require().NoError(err)

	msg := wasmdTypes.MsgExecuteContract{
		Sender:   acc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      reqBz,
	}

	_, _, _, err = chain.SendMsgs(acc, expPass, []sdk.Msg{&msg})
	if !expPass {
		s.Require().Error(err)
		return err
	}
	s.Require().NoError(err)

	return nil
}

// VoterGetCustomQuery returns the custom query result queried via Custom querier plugin.
func (s *E2ETestSuite) VoterGetCustomQuery(chain *e2eTesting.TestChain, contractAddr sdk.AccAddress, customQuery []byte, expPass bool) ([]byte, error) {
	req := voterTypes.MsgQuery{
		CustomCustom: customQuery,
	}

	res, err := chain.SmartQueryContract(contractAddr, expPass, req)
	if !expPass {
		return nil, err
	}

	var resp voterTypes.CustomCustomResponse
	s.Require().NoError(resp.UnmarshalJSON(res))

	return resp.Response, nil
}
