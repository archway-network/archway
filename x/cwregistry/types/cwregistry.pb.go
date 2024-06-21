// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: archway/cwregistry/v1/cwregistry.proto

package types

import (
	fmt "fmt"
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

// CodeMetadata defines the metadata of a contract code
type CodeMetadata struct {
	// The Code ID of the deployed contract
	CodeId uint64 `protobuf:"varint,1,opt,name=code_id,json=codeId,proto3" json:"code_id,omitempty"`
	// The information regarding the contract source codebase
	Source *SourceMetadata `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	// The information regarding the image used to build and optimize the contract binary
	SourceBuilder *SourceBuilder `protobuf:"bytes,3,opt,name=source_builder,json=sourceBuilder,proto3" json:"source_builder,omitempty"`
	// The JSON schema which specifies the interaction endpoints of the contract
	Schema string `protobuf:"bytes,4,opt,name=schema,proto3" json:"schema,omitempty"`
	// The contacts of the developers or security incidence handlers
	Contacts []string `protobuf:"bytes,5,rep,name=contacts,proto3" json:"contacts,omitempty"`
}

func (m *CodeMetadata) Reset()         { *m = CodeMetadata{} }
func (m *CodeMetadata) String() string { return proto.CompactTextString(m) }
func (*CodeMetadata) ProtoMessage()    {}
func (*CodeMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_2f25266653823114, []int{0}
}
func (m *CodeMetadata) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CodeMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CodeMetadata.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CodeMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CodeMetadata.Merge(m, src)
}
func (m *CodeMetadata) XXX_Size() int {
	return m.Size()
}
func (m *CodeMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_CodeMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_CodeMetadata proto.InternalMessageInfo

func (m *CodeMetadata) GetCodeId() uint64 {
	if m != nil {
		return m.CodeId
	}
	return 0
}

func (m *CodeMetadata) GetSource() *SourceMetadata {
	if m != nil {
		return m.Source
	}
	return nil
}

func (m *CodeMetadata) GetSourceBuilder() *SourceBuilder {
	if m != nil {
		return m.SourceBuilder
	}
	return nil
}

func (m *CodeMetadata) GetSchema() string {
	if m != nil {
		return m.Schema
	}
	return ""
}

func (m *CodeMetadata) GetContacts() []string {
	if m != nil {
		return m.Contacts
	}
	return nil
}

// SourceMetadata defines the metadata of the source code of a contract
type SourceMetadata struct {
	// The link to the code repository. e.g https://github.com/archway-network/archway
	Repository string `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// The tag of the commit message at which the binary was built and deployed. e.g v1.0.2
	Tag string `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	// The software license of the smart contract code. e.g Apache-2.0
	License string `protobuf:"bytes,3,opt,name=license,proto3" json:"license,omitempty"`
}

func (m *SourceMetadata) Reset()         { *m = SourceMetadata{} }
func (m *SourceMetadata) String() string { return proto.CompactTextString(m) }
func (*SourceMetadata) ProtoMessage()    {}
func (*SourceMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_2f25266653823114, []int{1}
}
func (m *SourceMetadata) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SourceMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SourceMetadata.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SourceMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SourceMetadata.Merge(m, src)
}
func (m *SourceMetadata) XXX_Size() int {
	return m.Size()
}
func (m *SourceMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_SourceMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_SourceMetadata proto.InternalMessageInfo

func (m *SourceMetadata) GetRepository() string {
	if m != nil {
		return m.Repository
	}
	return ""
}

func (m *SourceMetadata) GetTag() string {
	if m != nil {
		return m.Tag
	}
	return ""
}

func (m *SourceMetadata) GetLicense() string {
	if m != nil {
		return m.License
	}
	return ""
}

// SourceBuilder defines the metadata of the builder used to build the contract binary
type SourceBuilder struct {
	// Docker image. e.g cosmwasm/rust-optimizer
	Image string `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
	// Docker image tag. e.g 0.12.6
	Tag string `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	// Name of the generated contract binary. e.g counter.wasm
	ContractName string `protobuf:"bytes,3,opt,name=contract_name,json=contractName,proto3" json:"contract_name,omitempty"`
}

func (m *SourceBuilder) Reset()         { *m = SourceBuilder{} }
func (m *SourceBuilder) String() string { return proto.CompactTextString(m) }
func (*SourceBuilder) ProtoMessage()    {}
func (*SourceBuilder) Descriptor() ([]byte, []int) {
	return fileDescriptor_2f25266653823114, []int{2}
}
func (m *SourceBuilder) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SourceBuilder) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SourceBuilder.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SourceBuilder) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SourceBuilder.Merge(m, src)
}
func (m *SourceBuilder) XXX_Size() int {
	return m.Size()
}
func (m *SourceBuilder) XXX_DiscardUnknown() {
	xxx_messageInfo_SourceBuilder.DiscardUnknown(m)
}

var xxx_messageInfo_SourceBuilder proto.InternalMessageInfo

func (m *SourceBuilder) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *SourceBuilder) GetTag() string {
	if m != nil {
		return m.Tag
	}
	return ""
}

func (m *SourceBuilder) GetContractName() string {
	if m != nil {
		return m.ContractName
	}
	return ""
}

func init() {
	proto.RegisterType((*CodeMetadata)(nil), "archway.cwregistry.v1.CodeMetadata")
	proto.RegisterType((*SourceMetadata)(nil), "archway.cwregistry.v1.SourceMetadata")
	proto.RegisterType((*SourceBuilder)(nil), "archway.cwregistry.v1.SourceBuilder")
}

func init() {
	proto.RegisterFile("archway/cwregistry/v1/cwregistry.proto", fileDescriptor_2f25266653823114)
}

var fileDescriptor_2f25266653823114 = []byte{
	// 359 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0xcd, 0x6a, 0xea, 0x40,
	0x14, 0xc7, 0x9d, 0xab, 0xc6, 0x9b, 0x73, 0x55, 0x2e, 0x43, 0x3f, 0x42, 0x17, 0x21, 0xd8, 0x0f,
	0xb2, 0x69, 0x82, 0x2d, 0x5d, 0x76, 0x63, 0x57, 0xa5, 0xb4, 0x85, 0xe9, 0xae, 0x08, 0x32, 0x4e,
	0x86, 0x18, 0x6a, 0x32, 0x32, 0x33, 0x6a, 0xf3, 0x16, 0x7d, 0xac, 0x2e, 0x5d, 0x76, 0x59, 0x74,
	0xd1, 0xd7, 0x28, 0xc6, 0xa4, 0x44, 0x90, 0xee, 0xce, 0xef, 0xe4, 0xe4, 0xc7, 0xff, 0x0c, 0x07,
	0xce, 0xa8, 0x64, 0xa3, 0x39, 0x4d, 0x7d, 0x36, 0x97, 0x3c, 0x8c, 0x94, 0x96, 0xa9, 0x3f, 0xeb,
	0x96, 0xc8, 0x9b, 0x48, 0xa1, 0x05, 0xde, 0xcf, 0xe7, 0xbc, 0xd2, 0x97, 0x59, 0xb7, 0xf3, 0x85,
	0xa0, 0x79, 0x23, 0x02, 0x7e, 0xcf, 0x35, 0x0d, 0xa8, 0xa6, 0xf8, 0x10, 0x1a, 0x4c, 0x04, 0x7c,
	0x10, 0x05, 0x16, 0x72, 0x90, 0x5b, 0x23, 0xc6, 0x1a, 0x6f, 0x03, 0x7c, 0x0d, 0x86, 0x12, 0x53,
	0xc9, 0xb8, 0xf5, 0xc7, 0x41, 0xee, 0xbf, 0x8b, 0x53, 0x6f, 0xa7, 0xd1, 0x7b, 0xca, 0x86, 0x0a,
	0x1f, 0xc9, 0x7f, 0xc2, 0x77, 0xd0, 0xde, 0x54, 0x83, 0xe1, 0x34, 0x1a, 0x07, 0x5c, 0x5a, 0xd5,
	0x4c, 0x73, 0xf2, 0xab, 0xa6, 0xb7, 0x99, 0x25, 0x2d, 0x55, 0x46, 0x7c, 0x00, 0x86, 0x62, 0x23,
	0x1e, 0x53, 0xab, 0xe6, 0x20, 0xd7, 0x24, 0x39, 0xe1, 0x23, 0xf8, 0xcb, 0x44, 0xa2, 0x29, 0xd3,
	0xca, 0xaa, 0x3b, 0x55, 0xd7, 0x24, 0x3f, 0xdc, 0xe9, 0x43, 0x7b, 0x3b, 0x1a, 0xb6, 0x01, 0x24,
	0x9f, 0x08, 0x15, 0x69, 0x21, 0xd3, 0x6c, 0x5b, 0x93, 0x94, 0x3a, 0xf8, 0x3f, 0x54, 0x35, 0x0d,
	0xb3, 0x75, 0x4d, 0xb2, 0x2e, 0xb1, 0x05, 0x8d, 0x71, 0xc4, 0x78, 0xa2, 0x78, 0x96, 0xde, 0x24,
	0x05, 0x76, 0xfa, 0xd0, 0xda, 0x4a, 0x8c, 0xf7, 0xa0, 0x1e, 0xc5, 0x34, 0xe4, 0xb9, 0x77, 0x03,
	0x3b, 0x94, 0xc7, 0xd0, 0x5a, 0x47, 0x94, 0x94, 0xe9, 0x41, 0x42, 0xe3, 0x42, 0xdc, 0x2c, 0x9a,
	0x0f, 0x34, 0xe6, 0xbd, 0xc7, 0xf7, 0xa5, 0x8d, 0x16, 0x4b, 0x1b, 0x7d, 0x2e, 0x6d, 0xf4, 0xb6,
	0xb2, 0x2b, 0x8b, 0x95, 0x5d, 0xf9, 0x58, 0xd9, 0x95, 0xe7, 0xab, 0x30, 0xd2, 0xa3, 0xe9, 0xd0,
	0x63, 0x22, 0xf6, 0xf3, 0x87, 0x3c, 0x4f, 0xb8, 0x9e, 0x0b, 0xf9, 0x52, 0xb0, 0xff, 0x5a, 0xbe,
	0x0d, 0x9d, 0x4e, 0xb8, 0x1a, 0x1a, 0xd9, 0x51, 0x5c, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0xc0,
	0xbd, 0x47, 0x9a, 0x3e, 0x02, 0x00, 0x00,
}

func (m *CodeMetadata) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CodeMetadata) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CodeMetadata) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Contacts) > 0 {
		for iNdEx := len(m.Contacts) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Contacts[iNdEx])
			copy(dAtA[i:], m.Contacts[iNdEx])
			i = encodeVarintCwregistry(dAtA, i, uint64(len(m.Contacts[iNdEx])))
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.Schema) > 0 {
		i -= len(m.Schema)
		copy(dAtA[i:], m.Schema)
		i = encodeVarintCwregistry(dAtA, i, uint64(len(m.Schema)))
		i--
		dAtA[i] = 0x22
	}
	if m.SourceBuilder != nil {
		{
			size, err := m.SourceBuilder.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCwregistry(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if m.Source != nil {
		{
			size, err := m.Source.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCwregistry(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.CodeId != 0 {
		i = encodeVarintCwregistry(dAtA, i, uint64(m.CodeId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *SourceMetadata) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SourceMetadata) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SourceMetadata) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.License) > 0 {
		i -= len(m.License)
		copy(dAtA[i:], m.License)
		i = encodeVarintCwregistry(dAtA, i, uint64(len(m.License)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Tag) > 0 {
		i -= len(m.Tag)
		copy(dAtA[i:], m.Tag)
		i = encodeVarintCwregistry(dAtA, i, uint64(len(m.Tag)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Repository) > 0 {
		i -= len(m.Repository)
		copy(dAtA[i:], m.Repository)
		i = encodeVarintCwregistry(dAtA, i, uint64(len(m.Repository)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SourceBuilder) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SourceBuilder) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SourceBuilder) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ContractName) > 0 {
		i -= len(m.ContractName)
		copy(dAtA[i:], m.ContractName)
		i = encodeVarintCwregistry(dAtA, i, uint64(len(m.ContractName)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Tag) > 0 {
		i -= len(m.Tag)
		copy(dAtA[i:], m.Tag)
		i = encodeVarintCwregistry(dAtA, i, uint64(len(m.Tag)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Image) > 0 {
		i -= len(m.Image)
		copy(dAtA[i:], m.Image)
		i = encodeVarintCwregistry(dAtA, i, uint64(len(m.Image)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintCwregistry(dAtA []byte, offset int, v uint64) int {
	offset -= sovCwregistry(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *CodeMetadata) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.CodeId != 0 {
		n += 1 + sovCwregistry(uint64(m.CodeId))
	}
	if m.Source != nil {
		l = m.Source.Size()
		n += 1 + l + sovCwregistry(uint64(l))
	}
	if m.SourceBuilder != nil {
		l = m.SourceBuilder.Size()
		n += 1 + l + sovCwregistry(uint64(l))
	}
	l = len(m.Schema)
	if l > 0 {
		n += 1 + l + sovCwregistry(uint64(l))
	}
	if len(m.Contacts) > 0 {
		for _, s := range m.Contacts {
			l = len(s)
			n += 1 + l + sovCwregistry(uint64(l))
		}
	}
	return n
}

func (m *SourceMetadata) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Repository)
	if l > 0 {
		n += 1 + l + sovCwregistry(uint64(l))
	}
	l = len(m.Tag)
	if l > 0 {
		n += 1 + l + sovCwregistry(uint64(l))
	}
	l = len(m.License)
	if l > 0 {
		n += 1 + l + sovCwregistry(uint64(l))
	}
	return n
}

func (m *SourceBuilder) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Image)
	if l > 0 {
		n += 1 + l + sovCwregistry(uint64(l))
	}
	l = len(m.Tag)
	if l > 0 {
		n += 1 + l + sovCwregistry(uint64(l))
	}
	l = len(m.ContractName)
	if l > 0 {
		n += 1 + l + sovCwregistry(uint64(l))
	}
	return n
}

func sovCwregistry(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCwregistry(x uint64) (n int) {
	return sovCwregistry(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *CodeMetadata) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCwregistry
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
			return fmt.Errorf("proto: CodeMetadata: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CodeMetadata: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CodeId", wireType)
			}
			m.CodeId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CodeId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Source", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Source == nil {
				m.Source = &SourceMetadata{}
			}
			if err := m.Source.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceBuilder", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.SourceBuilder == nil {
				m.SourceBuilder = &SourceBuilder{}
			}
			if err := m.SourceBuilder.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Schema", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Schema = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Contacts", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Contacts = append(m.Contacts, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCwregistry(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCwregistry
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
func (m *SourceMetadata) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCwregistry
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
			return fmt.Errorf("proto: SourceMetadata: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SourceMetadata: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Repository", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Repository = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tag", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Tag = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field License", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.License = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCwregistry(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCwregistry
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
func (m *SourceBuilder) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCwregistry
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
			return fmt.Errorf("proto: SourceBuilder: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SourceBuilder: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Image", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Image = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tag", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Tag = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCwregistry
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
				return ErrInvalidLengthCwregistry
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCwregistry
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCwregistry(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCwregistry
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
func skipCwregistry(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCwregistry
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
					return 0, ErrIntOverflowCwregistry
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
					return 0, ErrIntOverflowCwregistry
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
				return 0, ErrInvalidLengthCwregistry
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCwregistry
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCwregistry
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCwregistry        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCwregistry          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCwregistry = fmt.Errorf("proto: unexpected end of group")
)
