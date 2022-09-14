package keeper

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/archway-network/archway/x/rewards/types"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
)

// ContractInfoReaderExpected defines the interface for the x/wasmd module dependency.
type ContractInfoReaderExpected interface {
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

// TrackingKeeperExpected defines the interface for the x/tracking module dependency.
type TrackingKeeperExpected interface {
	GetCurrentTxID(ctx sdk.Context) uint64
	GetBlockTrackingInfo(ctx sdk.Context, height int64) trackingTypes.BlockTracking
	RemoveBlockTrackingInfo(ctx sdk.Context, height int64)
}

// AuthKeeperExpected defines the interface for the x/auth module dependency.
type AuthKeeperExpected interface {
	GetModuleAccount(ctx sdk.Context, name string) authTypes.ModuleAccountI
}

// BankKeeperExpected defines the interface for the x/bank module dependency.
type BankKeeperExpected interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
}

// Keeper provides module state operations.
type Keeper struct {
	cdc              codec.Codec
	paramStore       paramTypes.Subspace
	state            State
	contractInfoView ContractInfoReaderExpected
	trackingKeeper   TrackingKeeperExpected
	authKeeper       AuthKeeperExpected
	bankKeeper       BankKeeperExpected
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, key sdk.StoreKey, contractInfoReader ContractInfoReaderExpected, trackingKeeper TrackingKeeperExpected, ak AuthKeeperExpected, bk BankKeeperExpected, ps paramTypes.Subspace) Keeper {
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
	}
}

// SetContractInfoViewer sets the contract info view dependency.
// Only for testing purposes.
func (k *Keeper) SetContractInfoViewer(viewer ContractInfoReaderExpected) {
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
