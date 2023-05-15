package genmsg

import (
	"fmt"
	v1 "github.com/archway-network/archway/x/genmsg/v1"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"log"
)

func anyToMsg(ir types.InterfaceRegistry, cdc codec.JSONCodec, anyMsg *types.Any) (sdk.Msg, error) {
	m, err := ir.Resolve(anyMsg.TypeUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve message: %s: %w", anyMsg.String(), err)
	}
	err = cdc.UnmarshalInterfaceJSON(anyMsg.Value, m)
	if err != nil {
		return nil, fmt.Errorf("unable to decode message: %s: %w", anyMsg.String(), err)
	}
	sdkMsg, ok := m.(sdk.Msg)
	if !ok {
		return nil, fmt.Errorf("message %s is not a sdk.Msg: %T", m.String(), m)
	}
	return sdkMsg, nil
}

func validateGenesis(cdc codec.JSONCodec, genesis *v1.GenesisState) error {
	interfaceRegistryProvider, ok := cdc.(interface {
		InterfaceRegistry() types.InterfaceRegistry
	})
	if !ok {
		return fmt.Errorf("codec does not implement InterfaceRegistry")
	}
	interfaceRegistry := interfaceRegistryProvider.InterfaceRegistry()
	// check if all messages are known by the codec
	for i, anyMsg := range genesis.Messages {
		if _, err := anyToMsg(interfaceRegistry, cdc, anyMsg); err != nil {
			return fmt.Errorf("at index %d: %w", i, err)
		}
	}
	return nil
}

func initGenesis(context sdk.Context, cdc codec.JSONCodec, router MessageRouter, genesis *v1.GenesisState) error {
	interfaceRegistryProvider, ok := cdc.(interface {
		InterfaceRegistry() types.InterfaceRegistry
	})
	if !ok {
		return fmt.Errorf("codec does not implement InterfaceRegistry")
	}
	interfaceRegistry := interfaceRegistryProvider.InterfaceRegistry()

	// execute all messages in order
	for i, anyMsg := range genesis.Messages {
		msg, err := anyToMsg(interfaceRegistry, cdc, anyMsg)
		if err != nil {
			return fmt.Errorf("at index %d: message decoding: %w", i, err)
		}
		handler := router.Handler(msg)
		resp, err := handler(context, msg)
		if err != nil {
			return fmt.Errorf("at index %d: message processing: %w", i, err)
		}
		log.Printf("message %d processed %s: %s", i, msg, resp.String())
	}
	return nil
}
