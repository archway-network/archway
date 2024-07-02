package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

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
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.ContractAddress)
	}

	if !k.sudoKeeper.HasContractInfo(ctx, senderAddr) {
		return nil, errors.Wrapf(types.ErrNotContract, "%s is not a contract address", msg.ContractAddress)
	}

	// Getting counterparty connection
	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, msg.ConnectionId)
	if !found {
		return nil, errors.Wrapf(types.ErrCounterpartyConnectionNotFoundID, "failed to get connection for counterparty %s", msg.ConnectionId)
	}
	icaMetadata := icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: msg.ConnectionId,
		HostConnectionId:       connectionEnd.Counterparty.ConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}
	icaMetadataBytes, err := icatypes.ModuleCdc.MarshalJSON(&icaMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "failed to MarshalJSON ica metadata")
	}
	version := string(icaMetadataBytes)

	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.ContractAddress, version); err != nil {
		return nil, errors.Wrap(err, "failed to RegisterInterchainAccount")
	}

	return &types.MsgRegisterInterchainAccountResponse{}, nil
}

// SendTx submits a transaction to the interchain account
func (k Keeper) SendTx(goCtx context.Context, msg *types.MsgSendTx) (*types.MsgSendTxResponse, error) {
	if msg == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "nil msg is prohibited")
	}

	if msg.Msgs == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "empty Msgs field is prohibited")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.ContractAddress)
	}

	if !k.sudoKeeper.HasContractInfo(ctx, senderAddr) {
		return nil, errors.Wrapf(types.ErrNotContract, "%s is not a contract address", msg.ContractAddress)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to GetParams")
	}
	if uint64(len(msg.Msgs)) > params.GetMsgSendTxMaxMessages() {
		return nil, fmt.Errorf(
			"MsgSubmitTx contains more messages than allowed, has=%d, max=%d",
			len(msg.Msgs),
			params.GetMsgSendTxMaxMessages(),
		)
	}

	portID, err := icatypes.NewControllerPortID(msg.ContractAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create NewControllerPortID")
	}

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		return nil, errors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to GetActiveChannelID for port %s", portID)
	}

	data, err := SerializeCosmosTxs(k.Codec, msg.Msgs)
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

	return &types.MsgSendTxResponse{
		SequenceId: sequence,
		Channel:    channelID,
	}, nil
}

// SerializeCosmosTxs serializes a slice of *types.Any messages using the CosmosTx type.
func SerializeCosmosTxs(cdc codec.BinaryCodec, msgs []*codectypes.Any) (bz []byte, err error) {
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

	if err := req.Params.Validate(); err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid parameters; %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
