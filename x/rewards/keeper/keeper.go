package keeper

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/archway-network/archway/pkg"
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

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetContractMetadata creates or updates the contract metadata verifying the ownership:
//   * Meta could be created by the contract admin (if set);
//   * Meta could be modified by the contract owner;
func (k Keeper) SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates types.ContractMetadata) error {
	state := k.state.ContractMetadataState(ctx)

	// Check if the contract exists
	contractInfo := k.contractInfoView.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return types.ErrContractNotFound
	}

	// Check ownership
	metaOld, metaExists := state.GetContractMetadata(contractAddr)
	if metaExists {
		if metaOld.OwnerAddress != senderAddr.String() {
			return sdkErrors.Wrap(types.ErrUnauthorized, "metadata can only be changed by the contract owner")
		}
	} else {
		if contractInfo.Admin != senderAddr.String() {
			return sdkErrors.Wrap(types.ErrUnauthorized, "metadata can only be created by the contract admin")
		}
	}

	// Build the updated meta
	metaNew := metaOld
	if !metaExists {
		metaNew.ContractAddress = contractAddr.String()
		metaNew.OwnerAddress = senderAddr.String()
	}
	if metaUpdates.HasOwnerAddress() {
		metaNew.OwnerAddress = metaUpdates.OwnerAddress
	}
	if metaUpdates.HasRewardsAddress() {
		metaNew.RewardsAddress = metaUpdates.RewardsAddress
	}

	// Set
	state.SetContractMetadata(contractAddr, metaNew)

	// Emit event
	types.EmitContractMetadataSetEvent(
		ctx,
		contractAddr,
		metaNew,
	)

	return nil
}

// GetContractMetadata returns the contract metadata for the given contract address (if found).
func (k Keeper) GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *types.ContractMetadata {
	meta, found := k.state.ContractMetadataState(ctx).GetContractMetadata(contractAddr)
	if !found {
		return nil
	}

	return &meta
}

// TrackFeeRebatesRewards creates a new transaction fee rebate reward record for the current transaction.
// Unique transaction ID is taken from the tracking module.
// CONTRACT: tracking Ante handler must be called before this module's Ante handler (tracking provides the primary key).
func (k Keeper) TrackFeeRebatesRewards(ctx sdk.Context, rewards sdk.Coins) {
	txID := k.trackingKeeper.GetCurrentTxID(ctx)
	k.state.TxRewardsState(ctx).CreateTxRewards(
		txID,
		ctx.BlockHeight(),
		rewards,
	)
}

// TrackInflationRewards creates a new inflation reward record for the current block.
func (k Keeper) TrackInflationRewards(ctx sdk.Context, rewards sdk.Coin) {
	k.state.BlockRewardsState(ctx).CreateBlockRewards(
		ctx.BlockHeight(),
		rewards,
		ctx.BlockGasMeter().Limit(),
	)
}

// UpdateMinConsensusFee calculates and updates the minimum consensus fee if eligible emitting an event.
func (k Keeper) UpdateMinConsensusFee(ctx sdk.Context, inflationRewards sdk.Coin) {
	// Prepare and verify inputs
	if inflationRewards.IsZero() {
		k.Logger(ctx).Info("Minimum consensus fee update skipped: inflation rewards are zero")
		return
	}
	inflationRewardsAmt := inflationRewards.Amount.ToDec()

	blockGasLimit := pkg.NewDecFromUint64(ctx.BlockGasMeter().Limit())
	if blockGasLimit.IsZero() {
		k.Logger(ctx).Info("Minimum consensus fee update skipped: block gas limit is not set")
		return
	}

	txFeeRebateRatio := k.TxFeeRebateRatio(ctx)

	// Calculate
	feeAmt := calculateMinConsensusFeeAmt(inflationRewardsAmt, blockGasLimit, txFeeRebateRatio)
	if feeAmt.IsZero() {
		k.Logger(ctx).Info("Minimum consensus fee update skipped: calculated amount is zero")
		return
	}
	feeCoin := sdk.DecCoin{
		Denom:  inflationRewards.Denom,
		Amount: feeAmt,
	}

	// Set and emit event
	k.state.MinConsensusFee(ctx).SetFee(feeCoin)
	k.Logger(ctx).Info("Minimum consensus fee update", "fee", feeCoin)

	types.EmitMinConsensusFeeSetEvent(ctx, feeCoin)
}

// GetMinConsensusFee returns the minimum consensus fee.
// Fee defines the minimum gas unit price for a transaction to be included in a block.
func (k Keeper) GetMinConsensusFee(ctx sdk.Context) *sdk.DecCoin {
	fee, found := k.state.MinConsensusFee(ctx).GetFee()
	if !found {
		return nil
	}

	return &fee
}

// UndistributedRewardsPool returns the current undistributed rewards leftovers.
func (k Keeper) UndistributedRewardsPool(ctx sdk.Context) sdk.Coins {
	poolAcc := k.authKeeper.GetModuleAccount(ctx, types.ContractRewardCollector)
	return k.bankKeeper.GetAllBalances(ctx, poolAcc.GetAddress())
}

// SetContractInfoViewer sets the contract info view dependency.
// Only for testing purposes.
func (k *Keeper) SetContractInfoViewer(viewer ContractInfoReaderExpected) {
	k.contractInfoView = viewer
}

// calculateMinConsensusFee calculates the minimum consensus fee amount using the formula:
//   -1 * ( BlockRewards / ( GasLimit * (TxFeeRatio - 1) ) )
// A simplified expression is used, original from specs: -1 * ( BlockRewards / ( GasLimit * TxFeeRatio - GasLimit ) )
func calculateMinConsensusFeeAmt(blockRewards, gasLimit, txFeeRatio sdk.Dec) sdk.Dec {
	return blockRewards.Quo(
		gasLimit.Mul(
			txFeeRatio.Sub(sdk.OneDec()),
		),
	).Neg()
}
