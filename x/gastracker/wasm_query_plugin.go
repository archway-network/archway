package gastracker

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type WasmQuerier interface {
	QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
	QueryRaw(ctx sdk.Context, contractAddr sdk.AccAddress, key []byte) []byte
}

func NewGasTrackingWASMQueryPlugin(gasTrackingKeeper GasTrackingKeeper, wasmQuerier WasmQuerier) func(ctx sdk.Context, request *wasmvmtypes.WasmQuery) ([]byte, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.WasmQuery) ([]byte, error) {
		if request.Smart != nil && request.Raw != nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "only one WasmQuery variant can be replied to"}
		}

		if request.Smart == nil && request.Raw == nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown WasmQuery variant"}
		}

		if request.Smart != nil {
			addr, err := sdk.AccAddressFromBech32(request.Smart.ContractAddr)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, request.Smart.ContractAddr)
			}

			// Check if we are inside a tx or not
			_, err = gasTrackingKeeper.GetCurrentTxTrackingInfo(ctx)
			if err != nil {
				switch err {
				case gstTypes.ErrBlockTrackingDataNotFound:
					return wasmQuerier.QuerySmart(ctx, addr, request.Smart.Msg)
				default:
					return nil, err
				}
			}

			gasTrackingQueryRequestWrapper := gstTypes.GasTrackingQueryRequestWrapper{
				MagicString:  GasTrackingQueryRequestMagicString,
				QueryRequest: request.Smart.Msg,
			}
			wrappedMsg, err := json.Marshal(gasTrackingQueryRequestWrapper)
			if err != nil {
				return nil, err
			}

			resp, err := wasmQuerier.QuerySmart(ctx, addr, wrappedMsg)
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

			if contractInstanceMetadata.GasRebateToUser && gasTrackingKeeper.IsGasRebateToUserEnabled(ctx) {
				ctx.Logger().Info("Rebating gas to the user", "contractAddress", request.Smart.ContractAddr, "gasConsumed", gasTrackingQueryResultWrapper.GasConsumed)
				ctx.GasMeter().RefundGas(gasTrackingQueryResultWrapper.GasConsumed, gstTypes.GasRebateToUserDescriptor)
			}

			if contractInstanceMetadata.CollectPremium && gasTrackingKeeper.IsContractPremiumEnabled(ctx) {
				ctx.Logger().Info("Charging premium to user", "premiumPercentage", contractInstanceMetadata.PremiumPercentageCharged)
				premiumGas := (gasTrackingQueryResultWrapper.GasConsumed * contractInstanceMetadata.PremiumPercentageCharged) / 100
				ctx.GasMeter().ConsumeGas(premiumGas, gstTypes.PremiumGasDescriptor)
			}

			ctx.Logger().Info("Got the tracking for Query", "gasConsumed", gasTrackingQueryResultWrapper.GasConsumed, "Contract address", request.Smart.ContractAddr)

			err = gasTrackingKeeper.TrackContractGasUsage(ctx, request.Smart.ContractAddr, gasTrackingQueryResultWrapper.GasConsumed, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, !contractInstanceMetadata.GasRebateToUser)
			if err != nil {
				return nil, err
			}

			return gasTrackingQueryResultWrapper.QueryResponse, nil
		} else {
			addr, err := sdk.AccAddressFromBech32(request.Raw.ContractAddr)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, request.Raw.ContractAddr)
			}
			return wasmQuerier.QueryRaw(ctx, addr, request.Raw.Key), nil
		}
	}
}
