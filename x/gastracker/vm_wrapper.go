package gastracker

import (
	"encoding/base64"
	"encoding/json"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GasTrackingWasmEngine struct {
	vm wasmTypes.WasmerEngine
	wasmGasRegister wasmkeeper.WasmGasRegister
}

func (g GasTrackingWasmEngine) createCustomGasTrackingMessage(contractOperationInfo gstTypes.ContractOperationInfo) (*wasmvmtypes.SubMsg, error) {
	bz, err := json.Marshal(contractOperationInfo)
	if err != nil {
		return nil, err
	}
	return &wasmvmtypes.SubMsg{
		Msg: wasmvmtypes.CosmosMsg{Custom: bz},
		ReplyOn: wasmvmtypes.ReplyNever,
	}, nil
}

func (g GasTrackingWasmEngine) Create(code wasmvm.WasmCode) (wasmvm.Checksum, error) {
	return g.vm.Create(code)
}

func (g GasTrackingWasmEngine) AnalyzeCode(checksum wasmvm.Checksum) (*wasmvmtypes.AnalysisReport, error) {
	return g.vm.AnalyzeCode(checksum)
}

func (g GasTrackingWasmEngine) Instantiate(checksum wasmvm.Checksum, env wasmvmtypes.Env, info wasmvmtypes.MessageInfo, initMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	var contractInstantiationWrapper gstTypes.ContractInstantiationRequestWrapper
	if err := json.Unmarshal(initMsg, &contractInstantiationWrapper); err != nil {
		return nil, 0, err
	}
	_, err := sdk.AccAddressFromBech32(contractInstantiationWrapper.RewardAddress)
	if err != nil {
		return nil, 0, err
	}

	if contractInstantiationWrapper.CollectPremium {
		if contractInstantiationWrapper.GasRebateToUser {
			return nil, 0, gstTypes.ErrInvalidInitRequest1
		}

		if contractInstantiationWrapper.PremiumPercentageCharged > 200 {
			return nil, 0, gstTypes.ErrInvalidInitRequest2
		}
	}

	base64Req := []byte(contractInstantiationWrapper.InstantiationRequest)
	data := make([]byte, base64.StdEncoding.DecodedLen(len(base64Req)))
	bytesDecoded, err := base64.StdEncoding.Decode(data, base64Req)
	if err != nil {
		return nil, 0, err
	}
	data = data[:bytesDecoded]

	response, gasUsed, err := g.vm.Instantiate(checksum, env, info, data, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
		RewardAddress:      contractInstantiationWrapper.RewardAddress,
		GasRebateToEndUser: contractInstantiationWrapper.GasRebateToUser,
		CollectPremium: contractInstantiationWrapper.CollectPremium,
		PremiumPercentageCharged: contractInstantiationWrapper.PremiumPercentageCharged,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) Execute(code wasmvm.Checksum, env wasmvmtypes.Env, info wasmvmtypes.MessageInfo, executeMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	response, gasUsed, err := g.vm.Execute(code, env, info, executeMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) Query(code wasmvm.Checksum, env wasmvmtypes.Env, queryMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) ([]byte, uint64, error) {
	var gasTrackingQueryRequestWrapper gstTypes.GasTrackingQueryRequestWrapper
	injectGasConsumedIntoResponse := true
	err := json.Unmarshal(queryMsg, &gasTrackingQueryRequestWrapper)
	if err != nil || gasTrackingQueryRequestWrapper.MagicString != GasTrackingQueryRequestMagicString {
		injectGasConsumedIntoResponse = false
	} else {
		queryMsg = gasTrackingQueryRequestWrapper.QueryRequest
	}

	queryResult, gasUsed, err := g.vm.Query(code, env, queryMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil || !injectGasConsumedIntoResponse {
		return queryResult, gasUsed, err
	}

	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	queryResultWrapper := gstTypes.GasTrackingQueryResultWrapper{
		GasConsumed:   wasmGasUsed,
		QueryResponse: queryResult,
	}

	bz, err := json.Marshal(queryResultWrapper)
	if err != nil {
		return queryResult, gasUsed, err
	}

	return bz, gasUsed, err
}

func (g GasTrackingWasmEngine) Migrate(checksum wasmvm.Checksum, env wasmvmtypes.Env, migrateMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	response, gasUsed, err := g.vm.Migrate(checksum, env, migrateMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_MIGRATE,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) Sudo(checksum wasmvm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	response, gasUsed, err := g.vm.Sudo(checksum, env, sudoMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) Reply(checksum wasmvm.Checksum, env wasmvmtypes.Env, reply wasmvmtypes.Reply, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	response, gasUsed, err := g.vm.Reply(checksum, env, reply, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_REPLY,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) GetCode(code wasmvm.Checksum) (wasmvm.WasmCode, error) {
	return g.vm.GetCode(code)
}

func (g GasTrackingWasmEngine) Cleanup() {
	g.vm.Cleanup()
}

func (g GasTrackingWasmEngine) IBCChannelOpen(checksum wasmvm.Checksum, env wasmvmtypes.Env, channel wasmvmtypes.IBCChannelOpenMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (uint64, error) {
	 return g.vm.IBCChannelOpen(checksum, env, channel, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

func (g GasTrackingWasmEngine) IBCChannelConnect(checksum wasmvm.Checksum, env wasmvmtypes.Env, channel wasmvmtypes.IBCChannelConnectMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	response, gasUsed, err := g.vm.IBCChannelConnect(checksum, env, channel, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_IBC,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) IBCChannelClose(checksum wasmvm.Checksum, env wasmvmtypes.Env, channel wasmvmtypes.IBCChannelCloseMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	response, gasUsed, err := g.vm.IBCChannelClose(checksum, env, channel, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_IBC,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) IBCPacketReceive(checksum wasmvm.Checksum, env wasmvmtypes.Env, packet wasmvmtypes.IBCPacketReceiveMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCReceiveResponse, uint64, error) {
	response, gasUsed, err := g.vm.IBCPacketReceive(checksum, env, packet, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_IBC,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) IBCPacketAck(checksum wasmvm.Checksum, env wasmvmtypes.Env, ack wasmvmtypes.IBCPacketAckMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	response, gasUsed, err := g.vm.IBCPacketAck(checksum, env, ack, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_IBC,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) IBCPacketTimeout(checksum wasmvm.Checksum, env wasmvmtypes.Env, packet wasmvmtypes.IBCPacketTimeoutMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	response, gasUsed, err := g.vm.IBCPacketTimeout(checksum, env, packet, store, goapi, querier, gasMeter, gasLimit, deserCost)
	if err != nil {
		return response, gasUsed, err
	}
	wasmGasUsed := g.wasmGasRegister.FromWasmVMGas(gasUsed)
	subMsg, err := g.createCustomGasTrackingMessage(gstTypes.ContractOperationInfo{
		GasConsumed:        wasmGasUsed,
		Operation:          gstTypes.ContractOperation_CONTRACT_OPERATION_IBC,
	})
	if err != nil {
		return response, gasUsed, err
	}
	response.Messages = append(response.Messages, *subMsg)
	return response, gasUsed, err
}

func (g GasTrackingWasmEngine) Pin(checksum wasmvm.Checksum) error {
	return g.vm.Pin(checksum)
}

func (g GasTrackingWasmEngine) Unpin(checksum wasmvm.Checksum) error {
	return g.vm.Unpin(checksum)
}

func (g GasTrackingWasmEngine) GetMetrics() (*wasmvmtypes.Metrics, error) {
	return g.vm.GetMetrics()
}

var _ wasmTypes.WasmerEngine = GasTrackingWasmEngine{}

func NewGasTrackingWasmEngine(vm wasmTypes.WasmerEngine, gasRegister wasmkeeper.WasmGasRegister) GasTrackingWasmEngine {
	return GasTrackingWasmEngine{vm, gasRegister}
}
