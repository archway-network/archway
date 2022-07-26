package e2eTesting

import (
	"encoding/json"
	"io/ioutil"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/archway-network/archway/x/gastracker"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// UploadContract uploads a contract and returns the codeID.
func (chain *TestChain) UploadContract(sender Account, wasmPath string, instantiatePerms wasmdTypes.AccessConfig) (codeID uint64) {
	t := chain.t

	wasmBlob, err := ioutil.ReadFile(wasmPath)
	require.NoError(t, err)

	txMsg := wasmdTypes.MsgStoreCode{
		Sender:                sender.Address.String(),
		WASMByteCode:          wasmBlob,
		InstantiatePermission: &instantiatePerms,
	}

	_, res, err := chain.SendMsgs(sender, true, []sdk.Msg{&txMsg})
	require.NoError(t, err)

	txRes := chain.ParseSDKResultData(res)
	require.Len(t, txRes.Data, 1)

	var resp wasmdTypes.MsgStoreCodeResponse
	require.NoError(t, resp.Unmarshal(txRes.Data[0].Data))
	codeID = resp.CodeID

	return
}

// InstantiateContract instantiates a contract and returns the contract address.
func (chain *TestChain) InstantiateContract(sender Account, codeID uint64, adminAddr, label string, funds sdk.Coins, msg json.Marshaler) (contractAddr sdk.AccAddress, instResp []byte) {
	t := chain.t

	var msgBz []byte
	if msg != nil {
		bz, err := msg.MarshalJSON()
		require.NoError(t, err)
		msgBz = bz
	}

	txMsg := wasmdTypes.MsgInstantiateContract{
		Sender: sender.Address.String(),
		Admin:  adminAddr,
		CodeID: codeID,
		Label:  label,
		Msg:    msgBz,
		Funds:  funds,
	}

	_, res, err := chain.SendMsgs(sender, true, []sdk.Msg{&txMsg})
	require.NoError(t, err)

	txRes := chain.ParseSDKResultData(res)
	require.Len(t, txRes.Data, 1)

	var resp wasmdTypes.MsgInstantiateContractResponse
	require.NoError(t, resp.Unmarshal(txRes.Data[0].Data))

	contractAddr, err = sdk.AccAddressFromBech32(resp.Address)
	require.NoError(t, err)
	instResp = resp.Data

	return
}

// SmartQueryContract queries a contract and returns the result.
func (chain *TestChain) SmartQueryContract(contractAddr sdk.AccAddress, expPass bool, msg json.Marshaler) ([]byte, error) {
	t := chain.t

	require.NotNil(t, msg)
	reqBz, err := msg.MarshalJSON()
	require.NoError(t, err)

	resp, err := chain.app.WASMKeeper.QuerySmart(chain.GetContext(), contractAddr, reqBz)
	if expPass {
		require.NoError(t, err)
		return resp, nil
	}
	require.Error(t, err)

	return nil, err
}

// GetContractInfo returns a contract info.
func (chain *TestChain) GetContractInfo(contractAddr sdk.AccAddress) wasmdTypes.ContractInfo {
	t := chain.t

	info := chain.app.WASMKeeper.GetContractInfo(chain.GetContext(), contractAddr)
	require.NotNil(t, info)

	return *info
}

// GetContractMetadata returns a contract metadata.
func (chain *TestChain) GetContractMetadata(contractAddr sdk.AccAddress) gastracker.ContractInstanceMetadata {
	t := chain.t

	metadata, err := chain.app.GasTrackingKeeper.GetContractMetadata(chain.GetContext(), contractAddr)
	require.NoError(t, err)

	return metadata
}

// SetContractMetadata sets a contract metadata.
func (chain *TestChain) SetContractMetadata(sender Account, contractAddr sdk.AccAddress, metadata gastracker.ContractInstanceMetadata) {
	t := chain.t

	txMsg := gastracker.MsgSetContractMetadata{
		Sender:          sender.Address.String(),
		ContractAddress: contractAddr.String(),
		Metadata:        &metadata,
	}

	_, _, err := chain.SendMsgs(sender, true, []sdk.Msg{&txMsg})
	require.NoError(t, err)
}
