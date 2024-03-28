package keeper

import (
	"fmt"

	"github.com/archway-network/archway/x/cwica/types"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		Codec               codec.BinaryCodec
		storeKey            storetypes.StoreKey
		channelKeeper       types.ChannelKeeper
		connectionKeeper    types.ConnectionKeeper
		errorsKeeper        types.ErrorsKeeper
		icaControllerKeeper types.ICAControllerKeeper
		sudoKeeper          types.WasmKeeper
		authority           string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	channelKeeper types.ChannelKeeper,
	connectionKeeper types.ConnectionKeeper,
	errorsKeeper types.ErrorsKeeper,
	icaControllerKeeper types.ICAControllerKeeper,
	sudoKeeper types.WasmKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		Codec:               cdc,
		storeKey:            storeKey,
		channelKeeper:       channelKeeper,
		connectionKeeper:    connectionKeeper,
		errorsKeeper:        errorsKeeper,
		icaControllerKeeper: icaControllerKeeper,
		sudoKeeper:          sudoKeeper,
		authority:           authority,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the authority of the keeper. Should be the governance module address.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetWasmKeeper sets the given wasm keeper.
// NOTE: Only for testing purposes
func (k *Keeper) SetWasmKeeper(wk types.WasmKeeper) {
	k.sudoKeeper = wk
}

// SetICAControllerKeeper sets the given ica controller keeper.
// NOTE: Only for testing purposes
func (k *Keeper) SetICAControllerKeeper(icak types.ICAControllerKeeper) {
	k.icaControllerKeeper = icak
}

// SetChannelKeeper sets the given channel keeper.
// NOTE: Only for testing purposes
func (k *Keeper) SetChannelKeeper(ck types.ChannelKeeper) {
	k.channelKeeper = ck
}

// SetConnectionKeeper sets the given connection keeper.
// NOTE: Only for testing purposes
func (k *Keeper) SetConnectionKeeper(ck types.ConnectionKeeper) {
	k.connectionKeeper = ck
}

// SetErrorsKeeper sets the given errors keeper.
// NOTE: Only for testing purposes
func (k *Keeper) SetErrorsKeeper(ek types.ErrorsKeeper) {
	k.errorsKeeper = ek
}
