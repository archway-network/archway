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
	stakingKeeper types.StakingKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, ps paramTypes.Subspace, sk types.StakingKeeper, bk types.BankKeeper) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramStore:    ps,
		stakingKeeper: sk,
		bankKeeper:    bk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	return k.stakingKeeper.BondedRatio(ctx)
}

func (k Keeper) BondDenom(ctx sdk.Context) string {
	return k.stakingKeeper.BondDenom(ctx)
}

// MintCoins creates new coins from thin air and adds it to the given module account.
func (k Keeper) MintCoin(ctx sdk.Context, name string, amt sdk.Coin) error {
	return k.bankKeeper.MintCoins(ctx, name, sdk.NewCoins(amt))
}

func (k Keeper) GetBondedTokenSupply(ctx sdk.Context) sdk.Coin {
	denom := k.BondDenom(ctx)
	return k.bankKeeper.GetSupply(ctx, denom)
}
