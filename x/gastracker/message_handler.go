package gastracker

import (
	"encoding/json"
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GasConsumptionMsgHandler struct {
	gastrackingKeeper GasTrackingKeeper
}

func (g GasConsumptionMsgHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	if msg.Custom == nil {
		return events, data, wasmTypes.ErrUnknownMsg
	}

	var contractOperationInfo gstTypes.ContractOperationInfo
	err = json.Unmarshal(msg.Custom, &contractOperationInfo)
	if err != nil {
		return events, data, err
	}

	// Checking if block tracking and tx tracking already in place
	_, err = g.gastrackingKeeper.GetCurrentTxTrackingInfo(ctx)
	if err != nil {
		return events, data, err
	}

	var contractInstanceMetadata gstTypes.ContractInstanceMetadata
	if contractOperationInfo.Operation == gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION {
		contractInstanceMetadata = gstTypes.ContractInstanceMetadata{
			RewardAddress: contractOperationInfo.RewardAddress,
			GasRebateToUser: contractOperationInfo.GasRebateToEndUser,
			CollectPremium: contractOperationInfo.CollectPremium,
			PremiumPercentageCharged: contractOperationInfo.PremiumPercentageCharged,
		}
		err = g.gastrackingKeeper.AddNewContractMetadata(ctx, contractAddr.String(), contractInstanceMetadata)
		if err != nil {
			return events, data, err
		}
	} else {
		contractInstanceMetadata, err = g.gastrackingKeeper.GetNewContractMetadata(ctx, contractAddr.String())
		if err != nil {
			return events, data, err
		}
	}

	if contractInstanceMetadata.GasRebateToUser {
		ctx.Logger().Info("Refunding gas to the user", "contractAddress", contractAddr.String(), "gasConsumed", contractOperationInfo.GasConsumed)
		ctx.GasMeter().RefundGas(contractOperationInfo.GasConsumed, gstTypes.GasRebateToUserDescriptor)
	}

	if contractInstanceMetadata.CollectPremium {
		ctx.Logger().Info("Charging premium to user", "premiumPercentage", contractInstanceMetadata.PremiumPercentageCharged)
		premiumGas := (contractOperationInfo.GasConsumed * contractInstanceMetadata.PremiumPercentageCharged) / 100
		ctx.GasMeter().ConsumeGas(premiumGas, gstTypes.PremiumGasDescriptor)
	}

	err = g.gastrackingKeeper.TrackContractGasUsage(ctx, contractAddr.String(), contractOperationInfo.GasConsumed, contractOperationInfo.Operation, !contractInstanceMetadata.GasRebateToUser)
	if err != nil {
		return events, data, err
	}

	return events, data, nil
}

func newGasConsumptionMsgHandler(gasTrackingKeeper GasTrackingKeeper) GasConsumptionMsgHandler {
	return GasConsumptionMsgHandler{gasTrackingKeeper}
}

func NewGasTrackingMessageHandler(
	router sdk.Router,
	channelKeeper wasmTypes.ChannelKeeper,
	capabilityKeeper wasmTypes.CapabilityKeeper,
	bankKeeper wasmTypes.Burner,
	unpacker codectypes.AnyUnpacker,
	portSource wasmTypes.ICS20TransferPortSource,
	gasTrackingKeeper GasTrackingKeeper,
	customEncoders ...*wasmKeeper.MessageEncoders,
) wasmKeeper.Messenger {
	encoders := wasmKeeper.DefaultEncoders(unpacker, portSource)
	for _, e := range customEncoders {
		encoders = encoders.Merge(e)
	}
	return wasmKeeper.NewMessageHandlerChain(
		newGasConsumptionMsgHandler(gasTrackingKeeper),
		wasmKeeper.NewSDKMessageHandler(router, encoders),
		wasmKeeper.NewIBCRawPacketHandler(channelKeeper, capabilityKeeper),
		wasmKeeper.NewBurnCoinMessageHandler(bankKeeper),
	)
}



