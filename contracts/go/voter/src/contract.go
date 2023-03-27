package src

import (
	"errors"
	"strconv"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

// Instantiate performs the contract state initialization.
func Instantiate(deps *std.Deps, env stdTypes.Env, info stdTypes.MessageInfo, msgBz []byte) (*stdTypes.Response, error) {
	deps.Api.Debug("Instantiate called")

	var msg types.MsgInstantiate
	if err := msg.UnmarshalJSON(msgBz); err != nil {
		return nil, types.NewErrInvalidRequest("msg JSON unmarshal: " + err.Error())
	}

	return handleMsgInstantiate(deps, info, msg)
}

// Migrate performs the contract state upgrade and can only be called by the contract admin.
// If the admin field is not set for a contract, contract is immutable.
func Migrate(deps *std.Deps, env stdTypes.Env, msgBz []byte) (*stdTypes.Response, error) {
	return nil, types.NewErrUnimplemented("Migrate")
}

// Execute performs the contract state change.
func Execute(deps *std.Deps, env stdTypes.Env, info stdTypes.MessageInfo, msgBz []byte) (*stdTypes.Response, error) {
	deps.Api.Debug("Execute called")

	var msg types.MsgExecute
	if err := msg.UnmarshalJSON(msgBz); err != nil {
		return nil, types.NewErrInvalidRequest("msg JSON unmarshal: " + err.Error())
	}

	switch {
	case msg.Release != nil:
		return handleMsgRelease(deps, env, info)
	case msg.NewVoting != nil:
		return handleMsgNewVoting(deps, env, info, *msg.NewVoting)
	case msg.Vote != nil:
		return handleMsgVote(deps, env, info, *msg.Vote)
	case msg.SendIBCVote != nil:
		return handleMsgSendIBCVote(deps, env, info, *msg.SendIBCVote)
	case msg.CustomCustom != nil:
		return handleMsgCustomCustom(msg.CustomCustom)
	case msg.CustomUpdateMetadata != nil:
		return handleMsgUpdateMetadata(*msg.CustomUpdateMetadata)
	case msg.CustomWithdrawRewards != nil:
		return handleMsgWithdrawRewards(deps, *msg.CustomWithdrawRewards)
	case msg.Fail != nil:
		return nil, errors.New("this call fails")
	case msg.ReplyOnError != nil:
		// this path sends a message that fails to the contract address provided.
		wasmMsg := stdTypes.ExecuteMsg{
			ContractAddr: *msg.ReplyOnError,
			Msg:          []byte(`{"fail":{}}`),
			Funds:        nil,
		}
		return &stdTypes.Response{
			Messages: []stdTypes.SubMsg{
				stdTypes.ReplyOnError(wasmMsg, 3),
			},
			Data:       nil,
			Attributes: nil,
			Events:     nil,
		}, nil
	}

	return nil, types.NewErrInvalidRequest("unknown execute request")
}

// Sudo performs the contract state change and can only be called by a native Cosmos module (like x/gov).
func Sudo(deps *std.Deps, env stdTypes.Env, msgBz []byte) (*stdTypes.Response, error) {
	deps.Api.Debug("Sudo called")

	var msg types.MsgSudo
	if err := msg.UnmarshalJSON(msgBz); err != nil {
		return nil, types.NewErrInvalidRequest("msg JSON unmarshal: " + err.Error())
	}

	switch {
	case msg.ChangeNewVotingCost != nil:
		return handleSudoChangeNewVotingCost(deps, *msg.ChangeNewVotingCost)
	case msg.ChangeVoteCost != nil:
		return handleSudoChangeVoteCost(deps, *msg.ChangeVoteCost)
	}

	return nil, types.NewErrInvalidRequest("unknown sudo request")
}

// Reply performs an optional contract state change.
// Endpoint is called when stdTypes.SubMsg was sent with (always/success/error) ReplyOn policy
// on other endpoint invocation (instantiate/execute/migrate/sudo/reply).
// SubMsg identification is done via stdTypes.SubMsg.ID field.
func Reply(deps *std.Deps, env stdTypes.Env, reply stdTypes.Reply) (*stdTypes.Response, error) {
	deps.Api.Debug("Reply called")

	if reply.ID == 3 {
		return new(stdTypes.Response), nil
	}

	replyType, found, err := state.GetReplyMsgType(deps.Storage, reply.ID)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}
	if !found {
		return nil, types.NewErrInternal("replyID (" + strconv.FormatUint(reply.ID, 10) + "): not found")
	}

	switch replyType {
	case state.ReplyMsgTypeBank:
		return handleReplyBankMsg(deps, reply.Result)
	case state.ReplyMsgTypeWithdraw:
		return handleReplyCustomWithdrawMsg(deps, reply.Result)
	}

	return nil, types.NewErrInternal("unknown replyMsgType: " + strconv.Itoa(int(replyType)))
}

// Query performs the contract state read.
func Query(deps *std.Deps, env stdTypes.Env, msgBz []byte) ([]byte, error) {
	deps.Api.Debug("Query called")

	var msg types.MsgQuery
	if err := msg.UnmarshalJSON(msgBz); err != nil {
		return nil, types.NewErrInvalidRequest("msg JSON unmarshal: " + err.Error())
	}

	var handlerRes std.JSONType
	var handlerErr error
	switch {
	case msg.Params != nil:
		handlerRes, handlerErr = queryParams(deps)
	case msg.Voting != nil:
		handlerRes, handlerErr = queryVoting(deps, *msg.Voting)
	case msg.Tally != nil:
		handlerRes, handlerErr = queryTally(deps, env, *msg.Tally)
	case msg.Open != nil:
		handlerRes, handlerErr = queryOpen(deps, env)
	case msg.ReleaseStats != nil:
		handlerRes, handlerErr = queryReleaseStats(deps)
	case msg.IBCStats != nil:
		handlerRes, handlerErr = queryIBCStats(deps, *msg.IBCStats)
	case msg.WithdrawStats != nil:
		handlerRes, handlerErr = queryWithdrawStats(deps)
	case msg.APIVerifySecp256k1Signature != nil:
		handlerRes, handlerErr = queryAPIVerifySecp256k1Signature(deps, *msg.APIVerifySecp256k1Signature)
	case msg.APIRecoverSecp256k1PubKey != nil:
		handlerRes, handlerErr = queryAPIRecoverSecp256k1PubKey(deps, *msg.APIRecoverSecp256k1PubKey)
	case msg.APIVerifyEd25519Signature != nil:
		handlerRes, handlerErr = queryAPIVerifyEd25519Signature(deps, *msg.APIVerifyEd25519Signature)
	case msg.APIVerifyEd25519Signatures != nil:
		handlerRes, handlerErr = queryAPIVerifyEd25519Signatures(deps, *msg.APIVerifyEd25519Signatures)
	case msg.CustomCustom != nil:
		handlerRes, handlerErr = queryCustomCustom(deps, msg.CustomCustom)
	case msg.CustomMetadata != nil:
		handlerRes, handlerErr = queryCustomMetadata(deps, env, *msg.CustomMetadata)
	case msg.CustomRewardsRecords != nil:
		handlerRes, handlerErr = queryCustomRewardsRecords(deps, env, *msg.CustomRewardsRecords)
	case msg.CustomGovVoteRequest != nil:
		handlerRes, handlerErr = queryCustomGovVote(deps, env, *msg.CustomGovVoteRequest)
	default:
		handlerErr = types.NewErrInvalidRequest("unknown query")
	}
	if handlerErr != nil {
		return nil, handlerErr
	}

	resBz, err := handlerRes.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("query result JSON marshal: " + err.Error())
	}

	return resBz, nil
}

// IBCChannelOpen performs the IBC handshake checks.
// Endpoint is  called when the contract to participating in the IBC channel handshake step.
// IBC protocol wise, this is either the "Channel Open Init" event on the initiating chain or the
// "Channel Open Try" on the counterparty chain.
// Protocol version and channel ordering should be verified for example.
func IBCChannelOpen(deps *std.Deps, env stdTypes.Env, openMsg stdTypes.IBCChannelOpenMsg) error {
	deps.Api.Debug("IBCChannelOpen called")

	return handleIBCChannelOpen(openMsg)
}

// IBCChannelConnect performs the IBC handshake checks.
// Endpoint is called when an IBC channel connection was established.
// IBC protocol wise, this is either the "Channel Open Ack" event on the initiating chain or the "Channel Open Confirm"
// on the counterparty chain).
func IBCChannelConnect(deps *std.Deps, env stdTypes.Env, connectMsg stdTypes.IBCChannelConnectMsg) (*stdTypes.IBCBasicResponse, error) {
	deps.Api.Debug("IBCChannelConnect called")

	return handleIBCChannelConnect(connectMsg)
}

// IBCChannelClose informs the contract that an IBC channel was closed.
// Endpoint is called when an IBC channel connection was closed.
// Once closed, channels cannot be reopened and identifiers cannot be reused.
func IBCChannelClose(deps *std.Deps, env stdTypes.Env, closeMsg stdTypes.IBCChannelCloseMsg) (*stdTypes.IBCBasicResponse, error) {
	deps.Api.Debug("IBCChannelClose called")

	return nil, types.NewErrUnimplemented("IBCChannelClose")
}

// IBCPacketReceive performs the contract state change on IBC received packet.
// Endpoint is called when an incoming IBC packet is received by a counterparty chain and should be processed.
func IBCPacketReceive(deps *std.Deps, env stdTypes.Env, receiveMsg stdTypes.IBCPacketReceiveMsg) (*stdTypes.IBCReceiveResponse, error) {
	deps.Api.Debug("IBCPacketReceive called")

	var msg types.MsgIBC
	if err := msg.UnmarshalJSON(receiveMsg.Packet.Data); err != nil {
		return nil, types.NewErrInvalidRequest("msg JSON unmarshal: " + err.Error())
	}

	switch {
	case msg.Vote != nil:
		return handleIBCMsgVote(deps, env, *msg.Vote)
	}

	return nil, types.NewErrInvalidRequest("unknown IBC packet")
}

// IBCPacketAck performs an optional contract state change on IBC sent packet acknowledgement.
// Endpoint is called when an outgoing IBC packet is acknowledged by a counterparty chain (packet
// processing success or failure).
func IBCPacketAck(deps *std.Deps, env stdTypes.Env, ackMsg stdTypes.IBCPacketAckMsg) (*stdTypes.IBCBasicResponse, error) {
	deps.Api.Debug("IBCPacketAck called")

	var origMsg types.MsgIBC
	if err := origMsg.UnmarshalJSON(ackMsg.OriginalPacket.Data); err != nil {
		return nil, types.NewErrInvalidRequest("original msg JSON unmarshal: " + err.Error())
	}

	switch {
	case origMsg.Vote != nil:
		return handleIBCAckVote(deps, *origMsg.Vote, ackMsg.Acknowledgement)
	}

	return nil, types.NewErrInvalidRequest("unknown IBC original packet")
}

// IBCPacketTimeout inform the contract that an outgoing IBC packet has timed out.
// Endpoint is called when an outgoing IBC packet wasn't received by a counterparty chain within timeout boundaries.
func IBCPacketTimeout(deps *std.Deps, env stdTypes.Env, timeoutMsg stdTypes.IBCPacketTimeoutMsg) (*stdTypes.IBCBasicResponse, error) {
	deps.Api.Debug("IBCPacketTimeout called♂️")

	var msg types.MsgIBC
	if err := msg.UnmarshalJSON(timeoutMsg.Packet.Data); err != nil {
		return nil, types.NewErrInvalidRequest("msg JSON unmarshal: " + err.Error())
	}

	switch {
	case msg.Vote != nil:
		return handleIBCTimeoutVote(deps, *msg.Vote)
	}

	return nil, types.NewErrInvalidRequest("unknown IBC packet")
}
