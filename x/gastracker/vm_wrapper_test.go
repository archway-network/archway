package gastracker

import (
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	cosmwasm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/store"
	stTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"
	"math/rand"
	"testing"
)

type testError struct {}

func (t *testError) Error() string {
	return "Fail"
}

var errTestFail = &testError{}

type loggingVMLog struct {
	MethodName string
	Message []byte
}

type loggingVMLogs []loggingVMLog

type loggingVM struct {
	logs    loggingVMLogs
	GasUsed uint64
	Fail bool
}

func (l *loggingVM) Create(code cosmwasm.WasmCode) (cosmwasm.Checksum, error) {
	panic("Not implemented")
}

func (l *loggingVM) AnalyzeCode(checksum cosmwasm.Checksum) (*wasmvmtypes.AnalysisReport, error) {
	panic("Not implemented")
}

func (l *loggingVM) Instantiate(checksum cosmwasm.Checksum, env wasmvmtypes.Env, info wasmvmtypes.MessageInfo, initMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.Response{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "Instantiate",
		Message:    initMsg,
	})
	return &wasmvmtypes.Response{}, l.GasUsed, nil
}

func (l *loggingVM) Execute(code cosmwasm.Checksum, env wasmvmtypes.Env, info wasmvmtypes.MessageInfo, executeMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.Response{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "Execute",
		Message:    executeMsg,
	})
	return &wasmvmtypes.Response{}, l.GasUsed, nil
}

func (l *loggingVM) Query(code cosmwasm.Checksum, env wasmvmtypes.Env, queryMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) ([]byte, uint64, error) {
	if l.Fail {
		return []byte{}, l.GasUsed, errTestFail
	}
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "Query",
		Message:    queryMsg,
	})
	return []byte{1}, l.GasUsed, nil
}

func (l *loggingVM) Migrate(checksum cosmwasm.Checksum, env wasmvmtypes.Env, migrateMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.Response{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "Migrate",
		Message:    migrateMsg,
	})
	return &wasmvmtypes.Response{}, l.GasUsed, nil
}

func (l *loggingVM) Sudo(checksum cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.Response{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "Sudo",
		Message:    sudoMsg,
	})
	return &wasmvmtypes.Response{}, l.GasUsed, nil
}

func (l *loggingVM) Reply(checksum cosmwasm.Checksum, env wasmvmtypes.Env, reply wasmvmtypes.Reply, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.Response{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "Reply",
		Message:    nil,
	})
	return &wasmvmtypes.Response{}, l.GasUsed, nil
}

func (l *loggingVM) GetCode(code cosmwasm.Checksum) (cosmwasm.WasmCode, error) {
	panic("not implemented in test")
}

func (l *loggingVM) Cleanup() {
	panic("not implemented in test")
}

func (l *loggingVM) IBCChannelOpen(checksum cosmwasm.Checksum, env wasmvmtypes.Env, channel wasmvmtypes.IBCChannelOpenMsg, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (uint64, error) {
	if l.Fail {
		return l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "IBCChannelOpen",
		Message:    nil,
	})
	return l.GasUsed, nil
}

func (l *loggingVM) IBCChannelConnect(checksum cosmwasm.Checksum, env wasmvmtypes.Env, channel wasmvmtypes.IBCChannelConnectMsg, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "IBCChannelConnect",
		Message:    nil,
	})
	return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, nil
}

func (l *loggingVM) IBCChannelClose(checksum cosmwasm.Checksum, env wasmvmtypes.Env, channel wasmvmtypes.IBCChannelCloseMsg, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "IBCChannelClose",
		Message:    nil,
	})
	return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, nil
}

func (l *loggingVM) IBCPacketReceive(checksum cosmwasm.Checksum, env wasmvmtypes.Env, packet wasmvmtypes.IBCPacketReceiveMsg, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCReceiveResponse, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.IBCReceiveResponse{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "IBCPacketReceive",
		Message:    nil,
	})
	return &wasmvmtypes.IBCReceiveResponse{}, l.GasUsed, nil
}

func (l *loggingVM) IBCPacketAck(checksum cosmwasm.Checksum, env wasmvmtypes.Env, ack wasmvmtypes.IBCPacketAckMsg, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "IBCPacketAck",
		Message:    nil,
	})
	return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, nil
}

func (l *loggingVM) IBCPacketTimeout(checksum cosmwasm.Checksum, env wasmvmtypes.Env, packet wasmvmtypes.IBCPacketTimeoutMsg, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	if l.Fail {
		return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, errTestFail
	}
	l.GasUsed = rand.Uint64() % 50000
	l.logs = append(l.logs, loggingVMLog{
		MethodName: "IBCPacketTimeout",
		Message:    nil,
	})
	return &wasmvmtypes.IBCBasicResponse{}, l.GasUsed, nil
}

func (l *loggingVM) Pin(checksum cosmwasm.Checksum) error {
	panic("not implemented in test")
}

func (l *loggingVM) Unpin(checksum cosmwasm.Checksum) error {
	panic("not implemented in test")
}

func (l *loggingVM) GetMetrics() (*wasmvmtypes.Metrics, error) {
	panic("not implemented in test")
}

func (l *loggingVM) Reset() {
	l.logs = nil
	l.Fail = false
}



func TestVMWrapper(t *testing.T) {

	defaultGasRegister := keeper.NewDefaultWasmGasRegister()
	var loggingVm = &loggingVM{
		GasUsed: defaultGasRegister.ToWasmVMGas(234),
	}

	vmWrapper := GasTrackingWasmEngine{
		vm: loggingVm,
		wasmGasRegister: defaultGasRegister,
	}

	kvStore := cosmwasm.KVStore(store.NewCommitMultiStore(db.NewMemDB()).GetCommitKVStore(stTypes.NewKVStoreKey("test")))

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	// Test 1: Everything is passed correctly
	request := gstTypes.ContractInstantiationRequestWrapper{
		RewardAddress:            "archway14hj2tavq8fpesdwxxcu44rty3hh90vhudldltd",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 205,
		InstantiationRequest:     "e30=",
	}
	msg, err := json.Marshal(request)
	require.NoError(t, err)

	response, gasUsed, err := vmWrapper.Instantiate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, msg, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Should succeed")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, len(response.Messages), 1)

	var contractOperationInfo gstTypes.ContractOperationInfo
	msg = response.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)
	require.Equal(t, request.GasRebateToUser, contractOperationInfo.GasRebateToEndUser)
	require.Equal(t, request.CollectPremium, contractOperationInfo.CollectPremium)
	require.Equal(t, request.PremiumPercentageCharged, contractOperationInfo.PremiumPercentageCharged)

	loggingVm.Reset()

	// Test 2: Everything is passed correctly
	request = gstTypes.ContractInstantiationRequestWrapper{
		RewardAddress:            "archway14hj2tavq8fpesdwxxcu44rty3hh90vhudldltd",
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 200,
		InstantiationRequest:     "e30=",
	}
	msg, err = json.Marshal(request)
	require.NoError(t, err)

	response, gasUsed, err = vmWrapper.Instantiate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, msg, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Should succeed")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "Instantiate", loggingVm.logs[0].MethodName)

	require.Equal(t, len(response.Messages), 1)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = response.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)
	require.Equal(t, request.GasRebateToUser, contractOperationInfo.GasRebateToEndUser)
	require.Equal(t, request.CollectPremium, contractOperationInfo.CollectPremium)
	require.Equal(t, request.PremiumPercentageCharged, contractOperationInfo.PremiumPercentageCharged)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.Instantiate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, msg, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	// Test 3: Invalid base64 string
	request = gstTypes.ContractInstantiationRequestWrapper{
		RewardAddress:            "archway14hj2tavq8fpesdwxxcu44rty3hh90vhudldltd",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 205,
		InstantiationRequest:     "a",
	}
	msg, err = json.Marshal(request)
	require.NoError(t, err)

	response, gasUsed, err = vmWrapper.Instantiate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, msg, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, "illegal base64 data at input byte 0","Should give an error about invalid base64")

	// Test 4: Both GasRebateToUser and CollectPremium is turned on
	request = gstTypes.ContractInstantiationRequestWrapper{
		RewardAddress:            "archway14hj2tavq8fpesdwxxcu44rty3hh90vhudldltd",
		GasRebateToUser:          true,
		CollectPremium:           true,
		PremiumPercentageCharged: 205,
		InstantiationRequest:     "e30=",
	}
	msg, err = json.Marshal(request)
	require.NoError(t, err)

	response, gasUsed, err = vmWrapper.Instantiate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, msg, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, gstTypes.ErrInvalidInitRequest1.Error(),"Should give an error about invalid base64")

	//Test 5: Premium percentage is greater than 200
	request = gstTypes.ContractInstantiationRequestWrapper{
		RewardAddress:            "archway14hj2tavq8fpesdwxxcu44rty3hh90vhudldltd",
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 205,
		InstantiationRequest:     "e30=",
	}
	msg, err = json.Marshal(request)
	require.NoError(t, err)

	response, gasUsed, err = vmWrapper.Instantiate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, msg, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, gstTypes.ErrInvalidInitRequest2.Error(),"Should give an error about invalid base64")

	//Test 6: Invalid bech32 string
	request = gstTypes.ContractInstantiationRequestWrapper{
		RewardAddress:            "1",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 205,
		InstantiationRequest:     "e30=",
	}
	msg, err = json.Marshal(request)
	require.NoError(t, err)

	response, gasUsed, err = vmWrapper.Instantiate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, msg, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, "decoding bech32 failed: invalid bech32 string length 1","Should give an error about invalid base64")


	response, gasUsed, err = vmWrapper.Execute(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, []byte{1}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "Execute", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = response.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	response, gasUsed, err = vmWrapper.Execute(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.MessageInfo{}, []byte{1}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	response, gasUsed, err = vmWrapper.Sudo(cosmwasm.Checksum{}, wasmvmtypes.Env{}, []byte{1}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "Sudo", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = response.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.Sudo(cosmwasm.Checksum{}, wasmvmtypes.Env{}, []byte{1}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	response, gasUsed, err = vmWrapper.Migrate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, []byte{1}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "Migrate", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = response.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_MIGRATE, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.Migrate(cosmwasm.Checksum{}, wasmvmtypes.Env{}, []byte{1}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	response, gasUsed, err = vmWrapper.Reply(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.Reply{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "Reply", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = response.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_REPLY, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.Reply(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.Reply{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	ibcResponse, gasUsed, err := vmWrapper.IBCPacketTimeout(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCPacketTimeoutMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "IBCPacketTimeout", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = ibcResponse.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_IBC, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.IBCPacketTimeout(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCPacketTimeoutMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	ibcResponse, gasUsed, err = vmWrapper.IBCPacketAck(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCPacketAckMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "IBCPacketAck", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = ibcResponse.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_IBC, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.IBCPacketAck(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCPacketAckMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	ibcReceiveResponse, gasUsed, err := vmWrapper.IBCPacketReceive(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCPacketReceiveMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "IBCPacketReceive", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = ibcReceiveResponse.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_IBC, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.IBCPacketReceive(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCPacketReceiveMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	ibcChannelCloseResp, gasUsed, err := vmWrapper.IBCChannelClose(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCChannelCloseMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "IBCChannelClose", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = ibcChannelCloseResp.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_IBC, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.IBCChannelClose(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCChannelCloseMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	ibcResponse, gasUsed, err = vmWrapper.IBCChannelConnect(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCChannelConnectMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "IBCChannelConnect", loggingVm.logs[0].MethodName)

	contractOperationInfo = gstTypes.ContractOperationInfo{}
	msg = ibcResponse.Messages[0].Msg.Custom
	err = json.Unmarshal(msg, &contractOperationInfo)
	require.NoError(t, err, "JSON unmarshalling should succeed")

	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_IBC, contractOperationInfo.Operation)
	require.Equal(t, defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), contractOperationInfo.GasConsumed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, _, err = vmWrapper.IBCChannelConnect(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCChannelConnectMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()

	gasUsed, err = vmWrapper.IBCChannelOpen(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCChannelOpenMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "IBCChannelOpen", loggingVm.logs[0].MethodName)

	require.Equal(t, loggingVm.GasUsed, gasUsed)

	loggingVm.Reset()

	loggingVm.Fail = true
	_, err = vmWrapper.IBCChannelOpen(cosmwasm.Checksum{}, wasmvmtypes.Env{}, wasmvmtypes.IBCChannelOpenMsg{}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()



	queryRequestWrapper := gstTypes.GasTrackingQueryRequestWrapper{
		MagicString:  gstTypes.MagicString,
		QueryRequest: []byte{1},
	}
	bz, err := json.Marshal(queryRequestWrapper)
	require.NoError(t, err, "Marshalling should not fail")

	queryResponse, gasUsed, err := vmWrapper.Query(cosmwasm.Checksum{}, wasmvmtypes.Env{}, bz, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "Query", loggingVm.logs[0].MethodName)

	queryResultWrapper := gstTypes.GasTrackingQueryResultWrapper{}
	err = json.Unmarshal(queryResponse, &queryResultWrapper)
	require.NoError(t, err, "JSON unmarshalling should succeed")
	require.Equal(t,defaultGasRegister.FromWasmVMGas(loggingVm.GasUsed), queryResultWrapper.GasConsumed)
	require.Equal(t, []byte{1}, queryResultWrapper.QueryResponse)

	loggingVm.Fail = true
	_, _, err = vmWrapper.Query(cosmwasm.Checksum{}, wasmvmtypes.Env{}, []byte{1}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.EqualError(t, err, errTestFail.Error(), "Should Fail")
	loggingVm.Reset()


	queryResponse, gasUsed, err = vmWrapper.Query(cosmwasm.Checksum{}, wasmvmtypes.Env{}, []byte{2}, kvStore, cosmwasm.GoAPI{}, nil, sdk.NewInfiniteGasMeter(), 50000, wasmvmtypes.UFraction{})
	require.NoError(t, err, "Contract should be executed successfully")
	require.Equal(t, loggingVm.GasUsed, gasUsed)
	require.Equal(t, 1, len(loggingVm.logs))
	require.Equal(t, "Query", loggingVm.logs[0].MethodName)

	require.Equal(t, []byte{1}, queryResponse)
}
