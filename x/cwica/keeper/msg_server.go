package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/archway-network/archway/x/cwica/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterInterchainAccount registers a new interchain account for the contract
func (k Keeper) RegisterInterchainAccount(goCtx context.Context, msg *types.MsgRegisterInterchainAccount) (*types.MsgRegisterInterchainAccountResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelRegisterInterchainAccount)
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.FromAddress)
	}

	if !k.sudoKeeper.HasContractInfo(ctx, senderAddr) {
		return nil, errors.Wrapf(types.ErrNotContract, "%s is not a contract address", msg.FromAddress)
	}

	// FIXME: empty version string doesn't look good
	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.FromAddress, ""); err != nil {
		return nil, errors.Wrap(err, "failed to RegisterInterchainAccount")
	}

	return &types.MsgRegisterInterchainAccountResponse{}, nil
}

// SubmitTx submits a transaction to the interchain account
func (k Keeper) SubmitTx(goCtx context.Context, msg *types.MsgSubmitTx) (*types.MsgSubmitTxResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelSubmitTx)

	if msg == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "nil msg is prohibited")
	}

	if msg.Msgs == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "empty Msgs field is prohibited")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.FromAddress)
	}

	if !k.sudoKeeper.HasContractInfo(ctx, senderAddr) {
		return nil, errors.Wrapf(types.ErrNotContract, "%s is not a contract address", msg.FromAddress)
	}

	params := k.GetParams(ctx)
	if uint64(len(msg.Msgs)) > params.GetMsgSubmitTxMaxMessages() {
		return nil, fmt.Errorf(
			"MsgSubmitTx contains more messages than allowed, has=%d, max=%d",
			len(msg.Msgs),
			params.GetMsgSubmitTxMaxMessages(),
		)
	}

	portID, err := icatypes.NewControllerPortID(msg.FromAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create NewControllerPortID")
	}

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		return nil, errors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to GetActiveChannelID for port %s", portID)
	}

	data, err := SerializeCosmosTx(k.Codec, msg.Msgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to SerializeCosmosTx")
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: msg.Memo,
	}

	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
	if !found {
		return nil, errors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", portID, channelID,
		)
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Duration(msg.Timeout) * time.Second).UnixNano()
	_, err = k.icaControllerKeeper.SendTx(ctx, nil, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return nil, errors.Wrap(err, "failed to SendTx")
	}

	return &types.MsgSubmitTxResponse{
		SequenceId: sequence,
		Channel:    channelID,
	}, nil
}

// SerializeCosmosTx serializes a slice of *types.Any messages using the CosmosTx type. The proto marshaled CosmosTx
// bytes are returned. This differs from icatypes.SerializeCosmosTx in that it does not serialize sdk.Msgs, but
// simply uses the already serialized values.
func SerializeCosmosTx(cdc codec.BinaryCodec, msgs []*codectypes.Any) (bz []byte, err error) {
	// only ProtoCodec is supported
	if _, ok := cdc.(*codec.ProtoCodec); !ok {
		return nil, errors.Wrap(icatypes.ErrInvalidCodec,
			"only ProtoCodec is supported for receiving messages on the host chain")
	}

	cosmosTx := &icatypes.CosmosTx{
		Messages: msgs,
	}

	bz, err = cdc.Marshal(cosmosTx)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// UpdateParams updates the module parameters
func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}
	authority := k.GetAuthority()
	if authority != req.Authority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
