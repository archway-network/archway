package gastracker

import (
	"encoding/json"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewGasTrackingWASMQueryPlugin(gasTrackingKeeper GasTrackingKeeper, wasmKeeper *wasmkeeper.Keeper) func(ctx sdk.Context, request *wasmvmtypes.WasmQuery) ([]byte, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.WasmQuery) ([]byte, error) {
		if request.Smart != nil {
			gasTrackingQueryRequestWrapper := gstTypes.GasTrackingQueryRequestWrapper{
				MagicString: GasTrackingQueryRequestMagicString,
				QueryRequest: request.Smart.Msg,
			}
			wrappedMsg, err := json.Marshal(gasTrackingQueryRequestWrapper)
			if err != nil {
				return nil, err
			}

			addr, err := sdk.AccAddressFromBech32(request.Smart.ContractAddr)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, request.Smart.ContractAddr)
			}

			resp, err := wasmKeeper.QuerySmart(ctx, addr, wrappedMsg)
			if err != nil {
				return nil, err
			}

			var gasTrackingQueryResultWrapper gstTypes.GasTrackingQueryResultWrapper
			err = json.Unmarshal(resp, &gasTrackingQueryResultWrapper)
			if err != nil {
				return nil, err
			}

			contractInstanceMetadata, err := gasTrackingKeeper.GetNewContractMetadata(ctx, request.Smart.ContractAddr)
			if err != nil {
				return nil, err
			}

			if contractInstanceMetadata.GasRebateToUser {
				ctx.Logger().Info("Refunding gas to the user", "contractAddress", request.Smart.ContractAddr, "gasConsumed", gasTrackingQueryResultWrapper.GasConsumed)
				ctx.GasMeter().RefundGas(gasTrackingQueryResultWrapper.GasConsumed, "Gas Refund for smart contract execution")
			}

			ctx.Logger().Info("Got the tracking for Query", "gasConsumed", gasTrackingQueryResultWrapper.GasConsumed, "Contract address", request.Smart.ContractAddr)

			err = gasTrackingKeeper.TrackContractGasUsage(ctx, request.Smart.ContractAddr, gasTrackingQueryResultWrapper.GasConsumed, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, !contractInstanceMetadata.GasRebateToUser)
			if err != nil {
				return nil, err
			}

			return gasTrackingQueryResultWrapper.QueryResponse, nil
		}
		if request.Raw != nil {
			addr, err := sdk.AccAddressFromBech32(request.Raw.ContractAddr)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, request.Raw.ContractAddr)
			}
			return wasmKeeper.QueryRaw(ctx, addr, request.Raw.Key), nil
		}
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown WasmQuery variant"}
	}
}
