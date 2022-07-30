package keeper

import (
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/archway-network/archway/x/tracking/types"
)

// Keeper provides module state operations.
type Keeper struct {
	WasmGasRegister wasmKeeper.GasRegister

	cdc        codec.Codec
	paramStore paramTypes.Subspace
	state      State
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, key sdk.StoreKey, gasRegister wasmKeeper.GasRegister) Keeper {
	return Keeper{
		cdc:             cdc,
		WasmGasRegister: gasRegister,
		state:           NewState(cdc, key),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// TrackNewTx creates a new transaction tracking info with a unique ID that is used to link new contract operations to.
// TxInfo object itself is created later during the EndBlocker.
func (k Keeper) TrackNewTx(ctx sdk.Context) {
	k.state.TxInfoState(ctx).CreateEmptyTxInfo()
}

// GetCurrentTxID returns the current transaction ID being tracked.
// That ID is used to link new contract operations and rewards tracking to the current transaction.
func (k Keeper) GetCurrentTxID(ctx sdk.Context) uint64 {
	return k.state.TxInfoState(ctx).GetCurrentTxID()
}

// TrackNewContractOperation creates a new contract operation tracking entry with a unique ID using the current transaction ID.
func (k Keeper) TrackNewContractOperation(ctx sdk.Context, contractAddr sdk.AccAddress, opType types.ContractOperation, vmGasConsumed, sdkGasConsumed uint64) {
	curTxID := k.GetCurrentTxID(ctx)
	k.state.ContractOpInfoState(ctx).CreateContractOpInfo(
		curTxID,
		contractAddr,
		opType,
		vmGasConsumed,
		sdkGasConsumed,
	)
}

// FinalizeBlockTxTracking updates block transactions total gas consumed value using tracked contract operations.
func (k Keeper) FinalizeBlockTxTracking(ctx sdk.Context) {
	txState := k.state.TxInfoState(ctx)
	contractOpState := k.state.ContractOpInfoState(ctx)

	for _, txInfo := range txState.GetTxInfosByBlock(ctx.BlockHeight()) {
		for _, contractOp := range contractOpState.GetContractOpInfoByTxID(txInfo.Id) {
			txInfo.TotalGas += contractOp.VmGas + contractOp.SdkGas
		}
		txState.SetTxInfo(txInfo)
	}
}

// GetBlockTrackingInfo returns block gas tracking info containing all transactions and contract operations.
func (k Keeper) GetBlockTrackingInfo(ctx sdk.Context, height int64) types.BlockTracking {
	txState := k.state.TxInfoState(ctx)
	contractOpState := k.state.ContractOpInfoState(ctx)

	var resp types.BlockTracking

	txInfos := txState.GetTxInfosByBlock(height)
	resp.Txs = make([]types.TxTracking, 0, len(txInfos))
	for _, txInfo := range txInfos {
		contractOps := contractOpState.GetContractOpInfoByTxID(txInfo.Id)
		resp.Txs = append(
			resp.Txs, types.TxTracking{
				Info:               txInfo,
				ContractOperations: contractOps,
			},
		)
	}

	return resp
}

// RemoveBlockTrackingInfo removes gas tracking entries for the given height.
func (k Keeper) RemoveBlockTrackingInfo(ctx sdk.Context, height int64) {
	k.state.DeleteTxInfosCascade(ctx, height)
}
