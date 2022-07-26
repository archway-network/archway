package wasmbinding

import (
	"github.com/archway-network/archway/x/gastracker"
	wasmBindingTypes "github.com/archway-network/archway/x/gastracker/wasmbinding/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ContractMetadataReader defines the GasTrackerKeeper expected operations.
type ContractMetadataReader interface {
	GetContractMetadata(ctx sdk.Context, address sdk.AccAddress) (gastracker.ContractInstanceMetadata, error)
}

// ContractMetadataWriter defines the GasTrackerKeeper expected operations.
type ContractMetadataWriter interface {
	AddPendingChangeForContractMetadataByContract(ctx sdk.Context, contractAddr sdk.AccAddress, req wasmBindingTypes.UpdateMetadataRequest) error
}

type ContractMetadataReaderWriter interface {
	ContractMetadataReader
	ContractMetadataWriter
}
