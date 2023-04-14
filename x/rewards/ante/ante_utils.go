package ante

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// RewardsKeeperExpected defines the expected interface for the x/rewards keeper.
type RewardsKeeperExpected interface {
	// Used in MinFeeDecorator
	GetMinConsensusFee(ctx sdk.Context) (sdk.DecCoin, bool)
	GetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress) (sdk.Coin, bool)
	CreateFlatFeeRewardsRecords(ctx sdk.Context, contractAddress sdk.AccAddress, flatfee sdk.Coin)

	// Used in DeductFeeDecorator
	TxFeeRebateRatio(ctx sdk.Context) sdk.Dec
	TrackFeeRebatesRewards(ctx sdk.Context, rewards sdk.Coins)
}

func GetContractFlatFees(ctx sdk.Context, rk RewardsKeeperExpected, codec codec.BinaryCodec, m sdk.Msg) (fees sdk.Coins, hasWasmMsgs bool, err error) {
	switch msg := m.(type) {
	case *wasmTypes.MsgMigrateContract:
		{
			return nil, true, nil
		}
	case *wasmTypes.MsgExecuteContract: // if msg is contract execute, fetch flatfee for msg.Contract address
		{
			ca, err := sdk.AccAddressFromBech32(msg.Contract)
			if err != nil {
				return nil, true, err
			}
			fee, found := rk.GetFlatFee(ctx, ca)
			if found {
				rk.CreateFlatFeeRewardsRecords(ctx, ca, fee)
				return sdk.NewCoins(fee), true, nil
			}
		}
	case *authz.MsgExec: // if msg is authz msg, unwrap the msg and check if any are wasmTypes.MsgExecuteContract
		{
			var flatfees sdk.Coins
			for _, v := range msg.Msgs {
				var wrappedMsg sdk.Msg
				err := codec.UnpackAny(v, &wrappedMsg)
				if err != nil {
					return nil, false, sdkErrors.Wrapf(sdkErrors.ErrUnauthorized, "error decoding authz messages")
				}
				fees, hasWasmMsgs, err := GetContractFlatFees(ctx, rk, codec, wrappedMsg)
				if err != nil {
					return nil, hasWasmMsgs, err
				}
				flatfees = flatfees.Add(fees...)
			}
			return flatfees, hasWasmMsgs, nil
		}
	}
	return nil, false, nil
}
