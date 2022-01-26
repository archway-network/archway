package gastracker

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"reflect"
)

const (
	QueryContractMetadata = "contract-metadata"
	QueryBlockGasTracking = "block-gas-tracking"
)

func NewLegacyQuerier(keeper GasTrackingKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			rsp interface{}
			err error
		)

		switch path[0] {
		case QueryContractMetadata:
			contractAddr, err := sdk.AccAddressFromBech32(path[1])
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
			}
			rsp, err = keeper.GetContractMetadata(ctx, contractAddr)
		case QueryBlockGasTracking:
			rsp, err = keeper.GetCurrentBlockTracking(ctx)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown data query endpoint")
		}

		if err != nil {
			return nil, err
		}

		if rsp == nil || reflect.ValueOf(rsp).IsNil() {
			return nil, nil
		}

		bz, err := json.MarshalIndent(rsp, "", "  ")
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil
	}
}
