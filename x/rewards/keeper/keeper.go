package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/archway-network/archway/x/rewards/types"
)

// Keeper provides module state operations.
type Keeper struct {
	cdc              codec.Codec
	paramStore       paramTypes.Subspace
	state            State
	contractInfoView types.ContractInfoReader
	trackingKeeper   types.TrackingKeeper
	authKeeper       types.AuthKeeper
	bankKeeper       types.BankKeeper
	mintKeeper       types.MintKeeper
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, key sdk.StoreKey,
	contractInfoReader types.ContractInfoReader,
	trackingKeeper types.TrackingKeeper,
	ak types.AuthKeeper,
	bk types.BankKeeper,
	mk types.MintKeeper,
	ps paramTypes.Subspace) Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:              cdc,
		paramStore:       ps,
		state:            NewState(cdc, key),
		contractInfoView: contractInfoReader,
		trackingKeeper:   trackingKeeper,
		authKeeper:       ak,
		bankKeeper:       bk,
		mintKeeper:       mk,
	}
}

// SetContractInfoViewer sets the contract info view dependency.
// Only for testing purposes.
func (k *Keeper) SetContractInfoViewer(viewer types.ContractInfoReader) {
	k.contractInfoView = viewer
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// UndistributedRewardsPool returns the current undistributed rewards (yet to be withdrawn).
func (k Keeper) UndistributedRewardsPool(ctx sdk.Context) sdk.Coins {
	poolAcc := k.authKeeper.GetModuleAccount(ctx, types.ContractRewardCollector)
	return k.bankKeeper.GetAllBalances(ctx, poolAcc.GetAddress())
}

// TreasuryPool returns the current undistributed treasury rewards.
func (k Keeper) TreasuryPool(ctx sdk.Context) sdk.Coins {
	poolAcc := k.authKeeper.GetModuleAccount(ctx, types.TreasuryCollector)
	return k.bankKeeper.GetAllBalances(ctx, poolAcc.GetAddress())
}

// GetRewardsRecords returns all the rewards records for a given rewards address paginated.
// Query checks the page limit and uses the default limit if not provided.
func (k Keeper) GetRewardsRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, pageReq *query.PageRequest) ([]types.RewardsRecord, *query.PageResponse, error) {
	if pageReq == nil {
		pageReq = &query.PageRequest{
			Limit: types.MaxRecordsQueryLimit,
		}
	}
	if pageReq.Limit > types.MaxRecordsQueryLimit {
		return nil, nil, sdkErrors.Wrapf(types.ErrInvalidRequest, "max records (%d) query limit exceeded", types.MaxRecordsQueryLimit)
	}

	return k.state.RewardsRecord(ctx).GetRewardsRecordByRewardsAddressPaginated(rewardsAddr, pageReq)
}

// GetInflationaryRewards gets the sdk.Coin distributed to the x/rewards in the current block
func (k Keeper) GetInflationaryRewards(ctx sdk.Context) (sdk.Coin, bool) {
	return k.mintKeeper.GetInflationForRecipient(ctx, types.ModuleName)
}
