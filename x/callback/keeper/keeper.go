package keeper

import (
	"cosmossdk.io/collections"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/archway-network/archway/internal/collcompat"
	"github.com/archway-network/archway/x/callback/types"
)

// Keeper provides module state operations.
type Keeper struct {
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
	wasmKeeper    types.WasmKeeperExpected
	rewardsKeeper types.RewardsKeeperExpected
	bankKeeper    types.BankKeeperExpected
	authority     string // this should be the x/gov module account

	Schema collections.Schema

	// Params key: ParamsKeyPrefix | value: Params
	Params collections.Item[types.Params]
	// Callbacks key: CallbackKeyPrefix | value: []Callback
	Callbacks collections.Map[collections.Triple[int64, []byte, uint64], types.Callback]
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, wk types.WasmKeeperExpected, rk types.RewardsKeeperExpected, bk types.BankKeeperExpected, authority string) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))
	k := Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		wasmKeeper:    wk,
		rewardsKeeper: rk,
		bankKeeper:    bk,
		authority:     authority,
		Params: collections.NewItem(
			sb,
			types.ParamsKeyPrefix,
			"params",
			collcompat.ProtoValue[types.Params](cdc),
		),
		Callbacks: collections.NewMap(
			sb,
			types.CallbackKeyPrefix,
			"callbacks",
			collections.TripleKeyCodec(collections.Int64Key, collections.BytesKey, collections.Uint64Key),
			collcompat.ProtoValue[types.Callback](cdc),
		),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the x/callback module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetWasmKeeper sets the given wasm keeper.
// Only for testing purposes
func (k *Keeper) SetWasmKeeper(wk types.WasmKeeperExpected) {
	k.wasmKeeper = wk
}

// SendToCallbackModule sends coins from the sender to the x/callback module account.
func (k Keeper) SendToCallbackModule(ctx sdk.Context, sender string, amount sdk.Coin) error {
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(amount))
}

// RefundFromCallbackModule sends coins from the x/callback module account to the recipient.
func (k Keeper) RefundFromCallbackModule(ctx sdk.Context, recipient string, amount sdk.Coin) error {
	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return err
	}
	if k.bankKeeper.BlockedAddr(recipientAddr) { // blocked accounts cant receive funds. so in that case we send to fee collector
		return k.SendToFeeCollector(ctx, amount)
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, sdk.NewCoins(amount))
}

// SendToFeeCollector sends coins from the x/callback module account to the fee collector account.
func (k Keeper) SendToFeeCollector(ctx sdk.Context, amount sdk.Coin) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, authTypes.FeeCollectorName, sdk.NewCoins(amount))
}
