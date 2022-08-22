package src

import (
	"bytes"
	"errors"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

// handleIBCChannelOpen handles IBC channel open event.
// Handles "Channel Open Init" / "Channel Open Try" and validates an IBC channel params.
func handleIBCChannelOpen(msg stdTypes.IBCChannelOpenMsg) error {
	switch {
	case msg.OpenInit != nil:
		// Do not check counterparty here, since it is not created yet (check on OpenTry later)
		if err := types.ValidateIBCChannelParams(msg.OpenInit.Channel, false); err != nil {
			return errors.New("openInit: " + err.Error())
		}
	case msg.OpenTry != nil:
		if err := types.ValidateIBCChannelParams(msg.OpenTry.Channel, true); err != nil {
			return errors.New("openTry: " + err.Error())
		}
		if err := types.ValidateIBCVersion(msg.OpenTry.CounterpartyVersion); err != nil {
			return errors.New("openTry: counterparty: " + err.Error())
		}
	default:
		return errors.New("unknown IBC channel open request")
	}

	return nil
}

// handleIBCChannelOpen handles IBC channel connect event.
// Handles "Channel Open Ack" validating an IBC channel params.
func handleIBCChannelConnect(msg stdTypes.IBCChannelConnectMsg) (*stdTypes.IBCBasicResponse, error) {
	switch {
	case msg.OpenAck != nil:
		if err := types.ValidateIBCVersion(msg.OpenAck.CounterpartyVersion); err != nil {
			return nil, errors.New("openAck: counterparty: " + err.Error())
		}
	case msg.OpenConfirm != nil:
	default:
		return nil, errors.New("unknown IBC channel connect request")
	}

	return &stdTypes.IBCBasicResponse{}, nil
}

// handleIBCMsgVote handles MsgIBC.Vote msg.
// Failure error is ignored since we are using standard "0x01" and "0x00" acknowledgements.
func handleIBCMsgVote(deps *std.Deps, env stdTypes.Env, req types.IBCVoteRequest) (retResp *stdTypes.IBCReceiveResponse, retErr error) {
	defer func() {
		if retErr != nil {
			retResp.Acknowledgement = types.IBCAckDataFailure
			retErr = nil
		}
	}()

	// Prepare OK ack
	retResp = &stdTypes.IBCReceiveResponse{
		Acknowledgement: types.IBCAckDataOK,
	}

	// Input check
	if err := req.Validate(); err != nil {
		retErr = types.NewErrInvalidRequest("req validation: " + err.Error())
		return
	}

	voting, err := state.GetVoting(deps.Storage, req.ID)
	if err != nil {
		retErr = types.NewErrInternal(err.Error())
		return
	}
	if voting == nil {
		retErr = types.NewErrInvalidRequest("voting with requested ID not found")
		return
	}

	if voting.IsClosed(env.Block.Time) {
		retErr = types.ErrVotingClosed
		return
	}
	if voting.HasVote(req.From) {
		retErr = types.ErrAlreadyVoted
		return
	}

	// Append vote and update contract state
	var voteErr error
	switch req.Vote {
	case "yes":
		voteErr = voting.AddYesVote(req.Option, req.From)
	case "no":
		voteErr = voting.AddNoVote(req.Option, req.From)
	}
	if voteErr != nil {
		retErr = types.NewErrInvalidRequest(voteErr.Error())
		return
	}

	if err := state.SetVoting(deps.Storage, *voting); err != nil {
		retErr = types.NewErrInternal(err.Error())
		return
	}

	retResp.Events = []stdTypes.Event{
		types.NewEventIBCVoteAdded(req.From, req.ID, req.Option, req.Vote),
	}

	return
}

// handleIBCAckVote handles MsgIBC.Vote ack msg.
func handleIBCAckVote(deps *std.Deps, origReq types.IBCVoteRequest, ackReq stdTypes.IBCAcknowledgement) (*stdTypes.IBCBasicResponse, error) {
	newIBCStatus := state.IBCPkgAckedStatus
	if !bytes.Equal(ackReq.Data, types.IBCAckDataOK) {
		newIBCStatus = state.IBCPkgRejectedStatus
	}

	ibcStats, err := state.GetIBCStats(deps.Storage, origReq.From, origReq.ID)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	ibcStats.Status = newIBCStatus
	if err := state.SetIBCStats(deps.Storage, ibcStats); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.IBCBasicResponse{}, nil
}

// handleIBCTimeoutVote handles MsgIBC.Vote timeout.
func handleIBCTimeoutVote(deps *std.Deps, req types.IBCVoteRequest) (*stdTypes.IBCBasicResponse, error) {
	ibcStats, err := state.GetIBCStats(deps.Storage, req.From, req.ID)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	ibcStats.Status = state.IBCPkgTimedOutStatus
	if err := state.SetIBCStats(deps.Storage, ibcStats); err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &stdTypes.IBCBasicResponse{}, nil
}
