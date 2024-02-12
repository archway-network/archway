package keeper

import (
	"fmt"

	"cosmossdk.io/errors"

	"github.com/archway-network/archway/x/interchaintxs/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	LabelSubmitTx                  = "submit_tx"
	LabelHandleAcknowledgment      = "handle_ack"
	LabelLabelHandleChanOpenAck    = "handle_chan_open_ack"
	LabelRegisterInterchainAccount = "register_interchain_account"
	LabelHandleTimeout             = "handle_timeout"
)

type (
	Keeper struct {
		Codec               codec.BinaryCodec
		storeKey            storetypes.StoreKey
		memKey              storetypes.StoreKey
		channelKeeper       types.ChannelKeeper
		icaControllerKeeper types.ICAControllerKeeper
		sudoKeeper          types.WasmKeeper
		bankKeeper          types.BankKeeper
		feeCollectorAddr    string
		authority           string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	channelKeeper types.ChannelKeeper,
	icaControllerKeeper types.ICAControllerKeeper,
	sudoKeeper types.WasmKeeper,
	bankKeeper types.BankKeeper,
	feeCollectorAddr string,
	authority string,
) *Keeper {
	return &Keeper{
		Codec:               cdc,
		storeKey:            storeKey,
		memKey:              memKey,
		channelKeeper:       channelKeeper,
		icaControllerKeeper: icaControllerKeeper,
		sudoKeeper:          sudoKeeper,
		bankKeeper:          bankKeeper,
		feeCollectorAddr:    feeCollectorAddr,
		authority:           authority,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ChargeFee(ctx sdk.Context, payer sdk.AccAddress, fee sdk.Coins) error {
	k.Logger(ctx).Debug("Trying to change fees", "payer", payer, "fee", fee)

	params := k.GetParams(ctx)

	if !fee.IsAnyGTE(params.RegisterFee) {
		return errors.Wrapf(sdkerrors.ErrInsufficientFee, "provided fee is less than min governance set ack fee: %s < %s", fee, params.RegisterFee)
	}

	feeCollectorAddress, err := sdk.AccAddressFromBech32(k.feeCollectorAddr)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to convert fee collector, bech32 to AccAddress: %s: %s", k.feeCollectorAddr, err.Error())
	}

	err = k.bankKeeper.SendCoins(ctx, payer, feeCollectorAddress, fee)
	if err != nil {
		return errors.Wrapf(err, "failed send fee(%s) from %s to %s", fee, payer, feeCollectorAddress)
	}
	return nil
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetICARegistrationFeeFirstCodeID returns code id, starting from which we charge fee for ICA registration
func (k Keeper) GetICARegistrationFeeFirstCodeID(ctx sdk.Context) (codeID uint64) {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.ICARegistrationFeeFirstCodeID)
	if bytes == nil {
		k.Logger(ctx).Debug("Fee register ICA code id key don't exists, GetLastCodeID returns 0")
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

// SetWasmKeeper sets the given wasm keeper.
// Only for testing purposes
func (k *Keeper) SetWasmKeeper(wk types.WasmKeeper) {
	k.sudoKeeper = wk
}

// SetICAControllerKeeper sets the given ica controller keeper.
// Only for testing purposes
func (k *Keeper) SetICAControllerKeeper(icak types.ICAControllerKeeper) {
	k.icaControllerKeeper = icak
}

// SetChannelKeeper sets the given channel keeper.
// Only for testing purposes
func (k *Keeper) SetChannelKeeper(ck types.ChannelKeeper) {
	k.channelKeeper = ck
}
