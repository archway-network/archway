// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: archway/callback/v1/callback.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Callback defines the callback structure.
type Callback struct {
	// contract_address is the address of the contract which is requesting the callback (bech32 encoded).
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	// job_id is an identifier the callback requestor can pass in to identify the callback when it happens.
	JobId uint64 `protobuf:"varint,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	// callback_height is the height at which the callback is executed.
	CallbackHeight int64 `protobuf:"varint,3,opt,name=callback_height,json=callbackHeight,proto3" json:"callback_height,omitempty"`
	// fee_split is the breakdown of the fees paid by the contract to reserve the callback
	FeeSplit *CallbackFeesFeeSplit `protobuf:"bytes,4,opt,name=fee_split,json=feeSplit,proto3" json:"fee_split,omitempty"`
	// reserved_by is the address which reserved the callback (bech32 encoded).
	ReservedBy string `protobuf:"bytes,5,opt,name=reserved_by,json=reservedBy,proto3" json:"reserved_by,omitempty"`
	// callback_gas_limit is the maximum gas that can be consumed by this callback.
	MaxGasLimit uint64 `protobuf:"varint,6,opt,name=max_gas_limit,json=maxGasLimit,proto3" json:"max_gas_limit,omitempty"`
}

func (m *Callback) Reset()         { *m = Callback{} }
func (m *Callback) String() string { return proto.CompactTextString(m) }
func (*Callback) ProtoMessage()    {}
func (*Callback) Descriptor() ([]byte, []int) {
	return fileDescriptor_91c209d2fabf62aa, []int{0}
}
func (m *Callback) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Callback) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Callback.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Callback) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Callback.Merge(m, src)
}
func (m *Callback) XXX_Size() int {
	return m.Size()
}
func (m *Callback) XXX_DiscardUnknown() {
	xxx_messageInfo_Callback.DiscardUnknown(m)
}

var xxx_messageInfo_Callback proto.InternalMessageInfo

func (m *Callback) GetContractAddress() string {
	if m != nil {
		return m.ContractAddress
	}
	return ""
}

func (m *Callback) GetJobId() uint64 {
	if m != nil {
		return m.JobId
	}
	return 0
}

func (m *Callback) GetCallbackHeight() int64 {
	if m != nil {
		return m.CallbackHeight
	}
	return 0
}

func (m *Callback) GetFeeSplit() *CallbackFeesFeeSplit {
	if m != nil {
		return m.FeeSplit
	}
	return nil
}

func (m *Callback) GetReservedBy() string {
	if m != nil {
		return m.ReservedBy
	}
	return ""
}

func (m *Callback) GetMaxGasLimit() uint64 {
	if m != nil {
		return m.MaxGasLimit
	}
	return 0
}

// CallbackFeesFeeSplit is the breakdown of all the fees that need to be paid by the contract to reserve a callback
type CallbackFeesFeeSplit struct {
	// transaction_fees is the transaction fees for the callback based on its gas consumption
	TransactionFees *types.Coin `protobuf:"bytes,1,opt,name=transaction_fees,json=transactionFees,proto3" json:"transaction_fees,omitempty"`
	// block_reservation_fees is the block reservation fees portion of the callback reservation fees
	BlockReservationFees *types.Coin `protobuf:"bytes,2,opt,name=block_reservation_fees,json=blockReservationFees,proto3" json:"block_reservation_fees,omitempty"`
	// future_reservation_fees is the future reservation fees portion of the callback reservation fees
	FutureReservationFees *types.Coin `protobuf:"bytes,3,opt,name=future_reservation_fees,json=futureReservationFees,proto3" json:"future_reservation_fees,omitempty"`
	// surplus_fees is any extra fees passed in for the registration of the callback
	SurplusFees *types.Coin `protobuf:"bytes,4,opt,name=surplus_fees,json=surplusFees,proto3" json:"surplus_fees,omitempty"`
}

func (m *CallbackFeesFeeSplit) Reset()         { *m = CallbackFeesFeeSplit{} }
func (m *CallbackFeesFeeSplit) String() string { return proto.CompactTextString(m) }
func (*CallbackFeesFeeSplit) ProtoMessage()    {}
func (*CallbackFeesFeeSplit) Descriptor() ([]byte, []int) {
	return fileDescriptor_91c209d2fabf62aa, []int{1}
}
func (m *CallbackFeesFeeSplit) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CallbackFeesFeeSplit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CallbackFeesFeeSplit.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CallbackFeesFeeSplit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CallbackFeesFeeSplit.Merge(m, src)
}
func (m *CallbackFeesFeeSplit) XXX_Size() int {
	return m.Size()
}
func (m *CallbackFeesFeeSplit) XXX_DiscardUnknown() {
	xxx_messageInfo_CallbackFeesFeeSplit.DiscardUnknown(m)
}

var xxx_messageInfo_CallbackFeesFeeSplit proto.InternalMessageInfo

func (m *CallbackFeesFeeSplit) GetTransactionFees() *types.Coin {
	if m != nil {
		return m.TransactionFees
	}
	return nil
}

func (m *CallbackFeesFeeSplit) GetBlockReservationFees() *types.Coin {
	if m != nil {
		return m.BlockReservationFees
	}
	return nil
}

func (m *CallbackFeesFeeSplit) GetFutureReservationFees() *types.Coin {
	if m != nil {
		return m.FutureReservationFees
	}
	return nil
}

func (m *CallbackFeesFeeSplit) GetSurplusFees() *types.Coin {
	if m != nil {
		return m.SurplusFees
	}
	return nil
}

// Params defines the module parameters.
type Params struct {
	// callback_gas_limit is the maximum gas that can be consumed by a callback.
	CallbackGasLimit uint64 `protobuf:"varint,1,opt,name=callback_gas_limit,json=callbackGasLimit,proto3" json:"callback_gas_limit,omitempty"`
	// max_block_reservation_limit is the maximum number of callbacks which can be registered in a given block.
	MaxBlockReservationLimit uint64 `protobuf:"varint,2,opt,name=max_block_reservation_limit,json=maxBlockReservationLimit,proto3" json:"max_block_reservation_limit,omitempty"`
	// max_future_reservation_limit is the maximum number of blocks in the future that a contract can request a callback in.
	MaxFutureReservationLimit uint64 `protobuf:"varint,3,opt,name=max_future_reservation_limit,json=maxFutureReservationLimit,proto3" json:"max_future_reservation_limit,omitempty"`
	// block_reservation_fee_multiplier is used to calculate a part of the reservation fees which will need to be paid when requesting the callback.
	BlockReservationFeeMultiplier cosmossdk_io_math.LegacyDec `protobuf:"bytes,4,opt,name=block_reservation_fee_multiplier,json=blockReservationFeeMultiplier,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"block_reservation_fee_multiplier"`
	// future_reservation_fee_multiplier is used to calculate a part of the reservation fees which will need to be paid while requesting the callback.
	FutureReservationFeeMultiplier cosmossdk_io_math.LegacyDec `protobuf:"bytes,5,opt,name=future_reservation_fee_multiplier,json=futureReservationFeeMultiplier,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"future_reservation_fee_multiplier"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_91c209d2fabf62aa, []int{2}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetCallbackGasLimit() uint64 {
	if m != nil {
		return m.CallbackGasLimit
	}
	return 0
}

func (m *Params) GetMaxBlockReservationLimit() uint64 {
	if m != nil {
		return m.MaxBlockReservationLimit
	}
	return 0
}

func (m *Params) GetMaxFutureReservationLimit() uint64 {
	if m != nil {
		return m.MaxFutureReservationLimit
	}
	return 0
}

func init() {
	proto.RegisterType((*Callback)(nil), "archway.callback.v1.Callback")
	proto.RegisterType((*CallbackFeesFeeSplit)(nil), "archway.callback.v1.CallbackFeesFeeSplit")
	proto.RegisterType((*Params)(nil), "archway.callback.v1.Params")
}

func init() {
	proto.RegisterFile("archway/callback/v1/callback.proto", fileDescriptor_91c209d2fabf62aa)
}

var fileDescriptor_91c209d2fabf62aa = []byte{
	// 624 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0x41, 0x6f, 0xd3, 0x30,
	0x14, 0x6e, 0xda, 0xad, 0xda, 0x5c, 0x60, 0x93, 0xd9, 0x20, 0xdd, 0x20, 0x2b, 0xbd, 0xd0, 0x49,
	0x2c, 0x51, 0xb7, 0x2b, 0x08, 0xd1, 0x4d, 0x05, 0xa4, 0x4d, 0x40, 0xb8, 0x71, 0x89, 0x1c, 0xc7,
	0x4d, 0xbd, 0x26, 0x71, 0x65, 0x3b, 0x5d, 0x83, 0x38, 0xf1, 0x0b, 0xf8, 0x29, 0x1c, 0x38, 0x73,
	0xde, 0x71, 0xe2, 0x84, 0x38, 0x4c, 0x68, 0xfb, 0x23, 0x28, 0x71, 0xd2, 0x4e, 0x5b, 0xa4, 0x49,
	0xdc, 0x5e, 0xbe, 0xf7, 0xbe, 0xef, 0x3d, 0x7f, 0xcf, 0x31, 0x68, 0x23, 0x8e, 0x87, 0x27, 0x28,
	0xb1, 0x30, 0x0a, 0x02, 0x17, 0xe1, 0x91, 0x35, 0xe9, 0xce, 0x62, 0x73, 0xcc, 0x99, 0x64, 0xf0,
	0x7e, 0x5e, 0x63, 0xce, 0xf0, 0x49, 0x77, 0xa3, 0xe9, 0x33, 0xe6, 0x07, 0xc4, 0xca, 0x4a, 0xdc,
	0x78, 0x60, 0xa1, 0x28, 0x51, 0xf5, 0x1b, 0x6b, 0x3e, 0xf3, 0x59, 0x16, 0x5a, 0x69, 0x94, 0xa3,
	0x06, 0x66, 0x22, 0x64, 0xc2, 0x72, 0x91, 0x20, 0xd6, 0xa4, 0xeb, 0x12, 0x89, 0xba, 0x16, 0x66,
	0x34, 0xca, 0xf3, 0x4d, 0x95, 0x77, 0x14, 0x51, 0x7d, 0xa8, 0x54, 0xfb, 0x6b, 0x15, 0x2c, 0xed,
	0xe7, 0xbd, 0xe1, 0x36, 0x58, 0xc5, 0x2c, 0x92, 0x1c, 0x61, 0xe9, 0x20, 0xcf, 0xe3, 0x44, 0x08,
	0x5d, 0x6b, 0x69, 0x9d, 0x65, 0x7b, 0xa5, 0xc0, 0x5f, 0x29, 0x18, 0xae, 0x83, 0xfa, 0x31, 0x73,
	0x1d, 0xea, 0xe9, 0xd5, 0x96, 0xd6, 0x59, 0xb0, 0x17, 0x8f, 0x99, 0xfb, 0xd6, 0x83, 0x4f, 0xc1,
	0x4a, 0x71, 0x12, 0x67, 0x48, 0xa8, 0x3f, 0x94, 0x7a, 0xad, 0xa5, 0x75, 0x6a, 0xf6, 0xbd, 0x02,
	0x7e, 0x93, 0xa1, 0xb0, 0x0f, 0x96, 0x07, 0x84, 0x38, 0x62, 0x1c, 0x50, 0xa9, 0x2f, 0xb4, 0xb4,
	0x4e, 0x63, 0x77, 0xdb, 0x2c, 0x31, 0xc3, 0x2c, 0x86, 0xeb, 0x13, 0x22, 0xfa, 0x84, 0x7c, 0x4c,
	0x09, 0xf6, 0xd2, 0x20, 0x8f, 0xe0, 0x16, 0x68, 0x70, 0x22, 0x08, 0x9f, 0x10, 0xcf, 0x71, 0x13,
	0x7d, 0x31, 0x9b, 0x16, 0x14, 0x50, 0x2f, 0x81, 0x6d, 0x70, 0x37, 0x44, 0x53, 0xc7, 0x47, 0xc2,
	0x09, 0x68, 0x48, 0xa5, 0x5e, 0xcf, 0xe6, 0x6d, 0x84, 0x68, 0xfa, 0x1a, 0x89, 0xc3, 0x14, 0x6a,
	0xff, 0xac, 0x82, 0xb5, 0xb2, 0x3e, 0xf0, 0x00, 0xac, 0x4a, 0x8e, 0x22, 0x81, 0xb0, 0xa4, 0x2c,
	0x72, 0x06, 0x84, 0x28, 0x43, 0x1a, 0xbb, 0x4d, 0x33, 0xb7, 0x31, 0xf5, 0xdc, 0xcc, 0x3d, 0x37,
	0xf7, 0x19, 0x8d, 0xec, 0x95, 0x2b, 0x94, 0x54, 0x0d, 0xbe, 0x03, 0x0f, 0xdc, 0x80, 0xe1, 0x91,
	0xa3, 0xc6, 0x42, 0x73, 0xad, 0xea, 0x6d, 0x5a, 0x6b, 0x19, 0xd1, 0x9e, 0xf3, 0x32, 0xc1, 0x0f,
	0xe0, 0xe1, 0x20, 0x96, 0x31, 0x27, 0x37, 0x15, 0x6b, 0xb7, 0x29, 0xae, 0x2b, 0xe6, 0x75, 0xc9,
	0xe7, 0xe0, 0x8e, 0x88, 0xf9, 0x38, 0x88, 0x85, 0xd2, 0x59, 0xb8, 0x4d, 0xa7, 0x91, 0x97, 0xa7,
	0xec, 0xf6, 0xf7, 0x1a, 0xa8, 0xbf, 0x47, 0x1c, 0x85, 0x02, 0x3e, 0x03, 0x70, 0x76, 0x03, 0xe6,
	0xa6, 0x6b, 0x99, 0xe9, 0xab, 0x45, 0xa6, 0x70, 0x1e, 0xbe, 0x00, 0x9b, 0xe9, 0x76, 0x6e, 0xda,
	0xa3, 0x68, 0xea, 0x6e, 0xe9, 0x21, 0x9a, 0xf6, 0xae, 0xf9, 0xa0, 0xe8, 0x2f, 0xc1, 0xa3, 0x94,
	0x5e, 0x62, 0x86, 0xe2, 0xd7, 0x32, 0x7e, 0x33, 0x44, 0xd3, 0xfe, 0xf5, 0x53, 0x2b, 0x81, 0xcf,
	0xa0, 0x55, 0xba, 0x1a, 0x27, 0x8c, 0x03, 0x49, 0xc7, 0x01, 0x25, 0x3c, 0xb3, 0x62, 0xb9, 0xd7,
	0x3d, 0x3d, 0xdf, 0xaa, 0xfc, 0x39, 0xdf, 0xda, 0x54, 0x8e, 0x08, 0x6f, 0x64, 0x52, 0x66, 0x85,
	0x48, 0x0e, 0xcd, 0x43, 0xe2, 0x23, 0x9c, 0x1c, 0x10, 0xfc, 0xeb, 0xc7, 0x0e, 0xc8, 0x0d, 0x3b,
	0x20, 0xd8, 0x7e, 0x5c, 0xb2, 0xbc, 0xa3, 0x99, 0x2e, 0xfc, 0x02, 0x9e, 0x94, 0x6f, 0xf1, 0x6a,
	0xf3, 0xc5, 0xff, 0x6d, 0x6e, 0x94, 0xed, 0x79, 0xde, 0xbd, 0x77, 0x74, 0x7a, 0x61, 0x68, 0x67,
	0x17, 0x86, 0xf6, 0xf7, 0xc2, 0xd0, 0xbe, 0x5d, 0x1a, 0x95, 0xb3, 0x4b, 0xa3, 0xf2, 0xfb, 0xd2,
	0xa8, 0x7c, 0xda, 0xf3, 0xa9, 0x1c, 0xc6, 0xae, 0x89, 0x59, 0x68, 0xe5, 0x7f, 0xe4, 0x4e, 0x44,
	0xe4, 0x09, 0xe3, 0xa3, 0xe2, 0xdb, 0x9a, 0xce, 0x1f, 0x35, 0x99, 0x8c, 0x89, 0x70, 0xeb, 0xd9,
	0x73, 0xb2, 0xf7, 0x2f, 0x00, 0x00, 0xff, 0xff, 0x71, 0xd4, 0xb7, 0x78, 0xf5, 0x04, 0x00, 0x00,
}

func (m *Callback) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Callback) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Callback) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.MaxGasLimit != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.MaxGasLimit))
		i--
		dAtA[i] = 0x30
	}
	if len(m.ReservedBy) > 0 {
		i -= len(m.ReservedBy)
		copy(dAtA[i:], m.ReservedBy)
		i = encodeVarintCallback(dAtA, i, uint64(len(m.ReservedBy)))
		i--
		dAtA[i] = 0x2a
	}
	if m.FeeSplit != nil {
		{
			size, err := m.FeeSplit.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCallback(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	if m.CallbackHeight != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.CallbackHeight))
		i--
		dAtA[i] = 0x18
	}
	if m.JobId != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.JobId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintCallback(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *CallbackFeesFeeSplit) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CallbackFeesFeeSplit) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CallbackFeesFeeSplit) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.SurplusFees != nil {
		{
			size, err := m.SurplusFees.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCallback(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	if m.FutureReservationFees != nil {
		{
			size, err := m.FutureReservationFees.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCallback(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if m.BlockReservationFees != nil {
		{
			size, err := m.BlockReservationFees.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCallback(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.TransactionFees != nil {
		{
			size, err := m.TransactionFees.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCallback(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.FutureReservationFeeMultiplier.Size()
		i -= size
		if _, err := m.FutureReservationFeeMultiplier.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintCallback(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	{
		size := m.BlockReservationFeeMultiplier.Size()
		i -= size
		if _, err := m.BlockReservationFeeMultiplier.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintCallback(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if m.MaxFutureReservationLimit != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.MaxFutureReservationLimit))
		i--
		dAtA[i] = 0x18
	}
	if m.MaxBlockReservationLimit != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.MaxBlockReservationLimit))
		i--
		dAtA[i] = 0x10
	}
	if m.CallbackGasLimit != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.CallbackGasLimit))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintCallback(dAtA []byte, offset int, v uint64) int {
	offset -= sovCallback(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Callback) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovCallback(uint64(l))
	}
	if m.JobId != 0 {
		n += 1 + sovCallback(uint64(m.JobId))
	}
	if m.CallbackHeight != 0 {
		n += 1 + sovCallback(uint64(m.CallbackHeight))
	}
	if m.FeeSplit != nil {
		l = m.FeeSplit.Size()
		n += 1 + l + sovCallback(uint64(l))
	}
	l = len(m.ReservedBy)
	if l > 0 {
		n += 1 + l + sovCallback(uint64(l))
	}
	if m.MaxGasLimit != 0 {
		n += 1 + sovCallback(uint64(m.MaxGasLimit))
	}
	return n
}

func (m *CallbackFeesFeeSplit) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.TransactionFees != nil {
		l = m.TransactionFees.Size()
		n += 1 + l + sovCallback(uint64(l))
	}
	if m.BlockReservationFees != nil {
		l = m.BlockReservationFees.Size()
		n += 1 + l + sovCallback(uint64(l))
	}
	if m.FutureReservationFees != nil {
		l = m.FutureReservationFees.Size()
		n += 1 + l + sovCallback(uint64(l))
	}
	if m.SurplusFees != nil {
		l = m.SurplusFees.Size()
		n += 1 + l + sovCallback(uint64(l))
	}
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.CallbackGasLimit != 0 {
		n += 1 + sovCallback(uint64(m.CallbackGasLimit))
	}
	if m.MaxBlockReservationLimit != 0 {
		n += 1 + sovCallback(uint64(m.MaxBlockReservationLimit))
	}
	if m.MaxFutureReservationLimit != 0 {
		n += 1 + sovCallback(uint64(m.MaxFutureReservationLimit))
	}
	l = m.BlockReservationFeeMultiplier.Size()
	n += 1 + l + sovCallback(uint64(l))
	l = m.FutureReservationFeeMultiplier.Size()
	n += 1 + l + sovCallback(uint64(l))
	return n
}

func sovCallback(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCallback(x uint64) (n int) {
	return sovCallback(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Callback) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCallback
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Callback: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Callback: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field JobId", wireType)
			}
			m.JobId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.JobId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CallbackHeight", wireType)
			}
			m.CallbackHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CallbackHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeeSplit", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.FeeSplit == nil {
				m.FeeSplit = &CallbackFeesFeeSplit{}
			}
			if err := m.FeeSplit.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReservedBy", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ReservedBy = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxGasLimit", wireType)
			}
			m.MaxGasLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxGasLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipCallback(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCallback
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *CallbackFeesFeeSplit) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCallback
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CallbackFeesFeeSplit: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CallbackFeesFeeSplit: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TransactionFees", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.TransactionFees == nil {
				m.TransactionFees = &types.Coin{}
			}
			if err := m.TransactionFees.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockReservationFees", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.BlockReservationFees == nil {
				m.BlockReservationFees = &types.Coin{}
			}
			if err := m.BlockReservationFees.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FutureReservationFees", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.FutureReservationFees == nil {
				m.FutureReservationFees = &types.Coin{}
			}
			if err := m.FutureReservationFees.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SurplusFees", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.SurplusFees == nil {
				m.SurplusFees = &types.Coin{}
			}
			if err := m.SurplusFees.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCallback(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCallback
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCallback
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CallbackGasLimit", wireType)
			}
			m.CallbackGasLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CallbackGasLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxBlockReservationLimit", wireType)
			}
			m.MaxBlockReservationLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxBlockReservationLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxFutureReservationLimit", wireType)
			}
			m.MaxFutureReservationLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxFutureReservationLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockReservationFeeMultiplier", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.BlockReservationFeeMultiplier.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FutureReservationFeeMultiplier", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.FutureReservationFeeMultiplier.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCallback(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCallback
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipCallback(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCallback
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthCallback
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCallback
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCallback
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCallback        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCallback          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCallback = fmt.Errorf("proto: unexpected end of group")
)
