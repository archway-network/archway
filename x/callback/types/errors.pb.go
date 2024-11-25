// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: archway/callback/v1/errors.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	math "math"
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

// ModuleErrors defines the module level error codes
type ModuleErrors int32

const (
	// ERR_UNKNOWN is the default error code
	ModuleErrors_ERR_UNKNOWN ModuleErrors = 0
	// ERR_OUT_OF_GAS is the error code when the contract callback exceeds the gas
	// limit allowed by the module
	ModuleErrors_ERR_OUT_OF_GAS ModuleErrors = 1
	// ERR_CONTRACT_EXECUTION_FAILED is the error code when the contract callback
	// execution fails
	ModuleErrors_ERR_CONTRACT_EXECUTION_FAILED ModuleErrors = 2
)

var ModuleErrors_name = map[int32]string{
	0: "ERR_UNKNOWN",
	1: "ERR_OUT_OF_GAS",
	2: "ERR_CONTRACT_EXECUTION_FAILED",
}

var ModuleErrors_value = map[string]int32{
	"ERR_UNKNOWN":                   0,
	"ERR_OUT_OF_GAS":                1,
	"ERR_CONTRACT_EXECUTION_FAILED": 2,
}

func (x ModuleErrors) String() string {
	return proto.EnumName(ModuleErrors_name, int32(x))
}

func (ModuleErrors) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_f0078bfce91cddb8, []int{0}
}

func init() {
	proto.RegisterEnum("archway.callback.v1.ModuleErrors", ModuleErrors_name, ModuleErrors_value)
}

func init() { proto.RegisterFile("archway/callback/v1/errors.proto", fileDescriptor_f0078bfce91cddb8) }

var fileDescriptor_f0078bfce91cddb8 = []byte{
	// 229 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x48, 0x2c, 0x4a, 0xce,
	0x28, 0x4f, 0xac, 0xd4, 0x4f, 0x4e, 0xcc, 0xc9, 0x49, 0x4a, 0x4c, 0xce, 0xd6, 0x2f, 0x33, 0xd4,
	0x4f, 0x2d, 0x2a, 0xca, 0x2f, 0x2a, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x86, 0xaa,
	0xd0, 0x83, 0xa9, 0xd0, 0x2b, 0x33, 0x94, 0x12, 0x49, 0xcf, 0x4f, 0xcf, 0x07, 0xcb, 0xeb, 0x83,
	0x58, 0x10, 0xa5, 0x5a, 0x61, 0x5c, 0x3c, 0xbe, 0xf9, 0x29, 0xa5, 0x39, 0xa9, 0xae, 0x60, 0x03,
	0x84, 0xf8, 0xb9, 0xb8, 0x5d, 0x83, 0x82, 0xe2, 0x43, 0xfd, 0xbc, 0xfd, 0xfc, 0xc3, 0xfd, 0x04,
	0x18, 0x84, 0x84, 0xb8, 0xf8, 0x40, 0x02, 0xfe, 0xa1, 0x21, 0xf1, 0xfe, 0x6e, 0xf1, 0xee, 0x8e,
	0xc1, 0x02, 0x8c, 0x42, 0x8a, 0x5c, 0xb2, 0x20, 0x31, 0x67, 0x7f, 0xbf, 0x90, 0x20, 0x47, 0xe7,
	0x90, 0x78, 0xd7, 0x08, 0x57, 0xe7, 0xd0, 0x10, 0x4f, 0x7f, 0xbf, 0x78, 0x37, 0x47, 0x4f, 0x1f,
	0x57, 0x17, 0x01, 0x26, 0x27, 0xdf, 0x13, 0x8f, 0xe4, 0x18, 0x2f, 0x3c, 0x92, 0x63, 0x7c, 0xf0,
	0x48, 0x8e, 0x71, 0xc2, 0x63, 0x39, 0x86, 0x0b, 0x8f, 0xe5, 0x18, 0x6e, 0x3c, 0x96, 0x63, 0x88,
	0x32, 0x4e, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x87, 0xba, 0x53, 0x37,
	0x2f, 0xb5, 0xa4, 0x3c, 0xbf, 0x28, 0x1b, 0xc6, 0xd7, 0xaf, 0x40, 0xf8, 0xad, 0xa4, 0xb2, 0x20,
	0xb5, 0x38, 0x89, 0x0d, 0xec, 0x5a, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb0, 0x21, 0x6b,
	0x03, 0xfc, 0x00, 0x00, 0x00,
}
