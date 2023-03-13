package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/archway-network/archway/x/mint/types"
)

// Keeper provides module state operations.
type Keeper struct {
	cdc           codec.Codec
	paramStore    paramTypes.Subspace
	storeKey      sdk.StoreKey
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, ps paramTypes.Subspace, bk types.BankKeeper, sk types.StakingKeeper) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramStore:    ps,
		bankKeeper:    bk,
		stakingKeeper: sk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// MintCoins creates new coins from thin air and adds it to the given module account.
func (k Keeper) MintCoins(ctx sdk.Context, amt sdk.Coins) error {
	return k.bankKeeper.MintCoins(ctx, types.ModuleName, amt)
}

// GetBondedTokenSupply retrieves the bond token supply from store
func (k Keeper) GetBondedTokenSupply(ctx sdk.Context) sdk.Coin {
	denom := k.BondDenom(ctx)
	return k.bankKeeper.GetSupply(ctx, denom)
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	return k.stakingKeeper.BondedRatio(ctx)
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) string {
	return k.stakingKeeper.BondDenom(ctx)
}

// SendCoinsFromModuleToModule sends the given number of coins from one module account to another module account
func (k Keeper) SendCoinsToModule(ctx sdk.Context, recipientModule string, amt sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipientModule, amt)
}
