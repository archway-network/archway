package src

import (
	"bytes"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/pkg"
	archwayCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

// handleMsgInstantiate handles types.MsgInstantiate msg.
func handleMsgInstantiate(deps *std.Deps, info stdTypes.MessageInfo, msg types.MsgInstantiate) (*stdTypes.Response, error) {
	// Input check
	params, err := msg.Params.ValidateAndConvert(deps.Api, info)
	if err != nil {
		return nil, types.NewErrInvalidRequest("msg validation: params: " + err.Error())
	}

	// Set initial contract state
	if err := state.SetParams(deps.Storage, params); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{}, nil
}

// handleMsgRelease handles MsgExecute.Release msg.
func handleMsgRelease(deps *std.Deps, env stdTypes.Env, info stdTypes.MessageInfo) (*stdTypes.Response, error) {
	// Input check
	senderAddr, err := deps.Api.CanonicalAddress(info.Sender)
	if err != nil {
		return nil, types.NewErrInvalidRequest("sender address: canonical convert: " + err.Error())
	}

	params, err := state.GetParams(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	if !bytes.Equal(senderAddr, params.OwnerAddr) {
		return nil, types.NewErrInvalidRequest("release can be done only by the contract owner")
	}

	// Transfer
	queryClient := std.QuerierWrapper{Querier: deps.Querier}
	contractFunds, err := queryClient.QueryAllBalances(env.Contract.Address)
	if err != nil {
		return nil, types.NewErrInternal("bank balance query: " + err.Error())
	}

	bankMsg := stdTypes.SendMsg{
		ToAddress: info.Sender,
		Amount:    contractFunds,
	}

	replyID, err := state.SetReplyMsgType(deps.Storage, state.ReplyMsgTypeBank)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	// Result
	res := types.ReleaseResponse{
		ReleasedAmount: contractFunds,
	}

	resBz, err := res.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("result JSON marshal: " + err.Error())
	}

	return &stdTypes.Response{
		Data: resBz,
		Messages: []stdTypes.SubMsg{
			stdTypes.ReplyOnSuccess(bankMsg, replyID),
		},
		Events: []stdTypes.Event{
			types.NewEventRelease(info.Sender),
		},
	}, nil
}

// handleMsgNewVoting handles MsgExecute.NewVoting msg.
func handleMsgNewVoting(deps *std.Deps, env stdTypes.Env, info stdTypes.MessageInfo, req types.NewVotingRequest) (*stdTypes.Response, error) {
	// Input check
	if err := req.Validate(); err != nil {
		return nil, types.NewErrInvalidRequest("req validation: " + err.Error())
	}

	params, err := state.GetParams(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	if err := pkg.CoinsContainMinAmount(info.Funds, params.NewVotingCost); err != nil {
		return nil, types.NewErrInvalidRequest(err.Error())
	}

	// Create a new voting
	votingID, err := state.NextVotingID(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	voting := state.NewVoting(votingID, req.Name, info.Sender, env.Block.Time, req.Duration, req.VoteOptions)

	// Update contract state
	state.SetLastVotingID(deps.Storage, votingID)
	if err := state.SetVoting(deps.Storage, voting); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	// Result
	res := types.NewVotingResponse{
		VotingID: votingID,
	}

	resBz, err := res.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("result JSON marshal: " + err.Error())
	}

	return &stdTypes.Response{
		Data: resBz,
		Events: []stdTypes.Event{
			types.NewEventVotingCreated(info.Sender, votingID),
		},
	}, nil
}

// handleMsgVote handles MsgExecute.Vote msg.
func handleMsgVote(deps *std.Deps, env stdTypes.Env, info stdTypes.MessageInfo, req types.VoteRequest) (*stdTypes.Response, error) {
	// Input check
	if err := req.Validate(); err != nil {
		return nil, types.NewErrInvalidRequest("req validation: " + err.Error())
	}

	params, err := state.GetParams(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	if err := pkg.CoinsContainMinAmount(info.Funds, params.VoteCost); err != nil {
		return nil, types.NewErrInvalidRequest(err.Error())
	}

	voting, err := state.GetVoting(deps.Storage, req.ID)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}
	if voting == nil {
		return nil, types.NewErrInvalidRequest("voting with requested ID not found")
	}

	if voting.IsClosed(env.Block.Time) {
		return nil, types.ErrVotingClosed
	}
	if voting.HasVote(info.Sender) {
		return nil, types.ErrAlreadyVoted
	}

	// Append vote and update contract state
	var voteErr error
	switch req.Vote {
	case "yes":
		voteErr = voting.AddYesVote(req.Option, info.Sender)
	case "no":
		voteErr = voting.AddNoVote(req.Option, info.Sender)
	}
	if voteErr != nil {
		return nil, types.NewErrInvalidRequest(voteErr.Error())
	}

	if err := state.SetVoting(deps.Storage, *voting); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{
		Events: []stdTypes.Event{
			types.NewEventVoteAdded(info.Sender, req.ID, req.Option, req.Vote),
		},
	}, nil
}

// handleMsgSendIBCVote handles MsgExecute.SendIBCVote msg.
func handleMsgSendIBCVote(deps *std.Deps, env stdTypes.Env, info stdTypes.MessageInfo, req types.SendIBCVoteRequest) (*stdTypes.Response, error) {
	// Input check
	if err := req.Validate(); err != nil {
		return nil, types.NewErrInvalidRequest("req validation: " + err.Error())
	}

	// Build IBC message
	params, err := state.GetParams(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	ibcData := types.MsgIBC{
		Vote: &types.IBCVoteRequest{
			VoteRequest: req.VoteRequest,
			From:        info.Sender,
		},
	}
	ibcDataBz, err := ibcData.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("ibcData JSON marshal: " + err.Error())
	}

	ibcTimeout := env.Block.Time + params.IBCSendTimeout
	ibcMsg := stdTypes.IBCMsg{
		SendPacket: &stdTypes.SendPacketMsg{
			ChannelID: req.ChannelID,
			Data:      ibcDataBz,
			Timeout: stdTypes.IBCTimeout{
				Timestamp: ibcTimeout,
			},
		},
	}

	// Save IBC stats
	ibcStats := state.NewIBCStats(info.Sender, req.ID, env)
	if err := state.SetIBCStats(deps.Storage, ibcStats); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{
		Messages: []stdTypes.SubMsg{
			stdTypes.NewSubMsg(ibcMsg),
		},
		Events: []stdTypes.Event{
			types.NewEventIBCVoteSent(info.Sender, req.ID, req.Option, req.Vote),
		},
	}, nil
}

// handleSudoChangeNewVotingCost handles MsgSudo.NewVotingCost msg.
func handleSudoChangeNewVotingCost(deps *std.Deps, req types.ChangeCostRequest) (*stdTypes.Response, error) {
	// Input check
	if err := req.Validate(); err != nil {
		return nil, types.NewErrInvalidRequest("req validation: " + err.Error())
	}

	// Update params state
	params, err := state.GetParams(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}
	oldCost := params.NewVotingCost

	params.NewVotingCost = req.NewCost
	if err := state.SetParams(deps.Storage, params); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{
		Events: []stdTypes.Event{
			types.NewEventNewVotingCostChanged(oldCost, req.NewCost),
		},
	}, nil
}

// handleSudoChangeVoteCost handles MsgSudo.VoteCost msg.
func handleSudoChangeVoteCost(deps *std.Deps, req types.ChangeCostRequest) (*stdTypes.Response, error) {
	// Input check
	if err := req.Validate(); err != nil {
		return nil, types.NewErrInvalidRequest("req validation: " + err.Error())
	}

	// Update params state
	params, err := state.GetParams(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}
	oldCost := params.VoteCost

	params.VoteCost = req.NewCost
	if err := state.SetParams(deps.Storage, params); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{
		Events: []stdTypes.Event{
			types.NewEventVoteCostChanged(oldCost, req.NewCost),
		},
	}, nil
}

// handleReplyBankMsg handles a Reply from the x/bank Send sub call.
// Handler adjusts the contract release stats.
func handleReplyBankMsg(deps *std.Deps, reply stdTypes.SubcallResult) (*stdTypes.Response, error) {
	// Input check
	if reply.Err != "" {
		return nil, types.NewErrInvalidRequest("x/bank reply: error received")
	}
	if reply.Ok == nil {
		return nil, types.NewErrInvalidRequest("x/bank reply: Ok is nil")
	}

	var releasedAmt []stdTypes.Coin
out:
	for _, event := range reply.Ok.Events {
		if event.Type != "transfer" {
			continue
		}

		for _, attr := range event.Attributes {
			if attr.Key != "amount" {
				continue
			}

			coins, err := pkg.ParseCoinsFromString(attr.Value)
			if err != nil {
				return nil, types.NewErrInvalidRequest("x/bank reply: parsing transfer.amount value attribute: " + err.Error())
			}
			releasedAmt = coins
			break out
		}
	}
	// The following check is disabled because the x/wasmd v0.29.X has disabled SDK events pass through.
	//if len(releasedAmt) == 0 {
	//	return nil, types.NewErrInvalidRequest("x/bank reply: transfer.amount attribute: not found")
	//}

	// Update release stats
	stats, err := state.GetReleaseStats(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	stats.AddRelease(releasedAmt)
	if err := state.SetReleaseStats(deps.Storage, stats); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{}, nil
}

// handleMsgCustomCustom handles MsgExecute.CustomCustom msg creating a custom Cosmos msg for the Custom WASM handler.
func handleMsgCustomCustom(req stdTypes.RawMessage) (*stdTypes.Response, error) {
	msg := stdTypes.CosmosMsg{
		Custom: req,
	}

	return &stdTypes.Response{
		Messages: []stdTypes.SubMsg{
			stdTypes.NewSubMsg(msg),
		},
	}, nil
}

// handleMsgUpdateMetadata handles MsgExecute.CustomUpdateMetadata msg creating a custom Cosmos msg for the Custom WASM handler.
func handleMsgUpdateMetadata(req archwayCustomTypes.UpdateContractMetadataRequest) (*stdTypes.Response, error) {
	// Build msg
	customMsg := archwayCustomTypes.CustomMsg{
		UpdateContractMetadata: &req,
	}
	customMsgBz, err := customMsg.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("customMsg JSON marshal: " + err.Error())
	}

	msg := stdTypes.CosmosMsg{
		Custom: customMsgBz,
	}

	return &stdTypes.Response{
		Messages: []stdTypes.SubMsg{
			stdTypes.NewSubMsg(msg),
		},
	}, nil
}

// handleMsgWithdrawRewards handles MsgExecute.CustomWithdrawRewards msg creating a custom Cosmos msg for the Custom WASM handler with Reply.
func handleMsgWithdrawRewards(deps *std.Deps, req archwayCustomTypes.WithdrawRewardsRequest) (*stdTypes.Response, error) {
	// Build msg
	customMsg := archwayCustomTypes.CustomMsg{
		WithdrawRewards: &req,
	}
	customMsgBz, err := customMsg.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("customMsg JSON marshal: " + err.Error())
	}

	msg := stdTypes.CosmosMsg{
		Custom: customMsgBz,
	}

	replyID, err := state.SetReplyMsgType(deps.Storage, state.ReplyMsgTypeWithdraw)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{
		Messages: []stdTypes.SubMsg{
			stdTypes.ReplyOnSuccess(msg, replyID),
		},
	}, nil
}

// handleReplyCustomWithdrawMsg handles a Reply from the x/rewards CustomMsg sub call.
// Handler adjusts the contract release stats.
func handleReplyCustomWithdrawMsg(deps *std.Deps, reply stdTypes.SubcallResult) (*stdTypes.Response, error) {
	// Input check
	if reply.Err != "" {
		return nil, types.NewErrInvalidRequest("x/rewards CustomMsg reply: error received")
	}
	if reply.Ok == nil {
		return nil, types.NewErrInvalidRequest("x/rewards CustomMsg reply: Ok is nil")
	}

	// Parse reply data
	var data archwayCustomTypes.WithdrawRewardsResponse
	if err := data.UnmarshalJSON(reply.Ok.Data); err != nil {
		return nil, types.NewErrInternal("x/rewards CustomMsg reply: JSON unmarshal: " + err.Error())
	}

	// Update release stats
	stats, err := state.GetWithdrawStats(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	stats.AddWithdraw(data.TotalRewards, data.RecordsNum)
	if err := state.SetWithdrawStats(deps.Storage, stats); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.Response{}, nil
}
