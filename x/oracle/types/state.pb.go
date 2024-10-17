// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: archway/oracle/v1/state.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	github_com_archway_network_archway_x_oracle_asset "github.com/archway-network/archway/x/oracle/asset"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// a snapshot of the prices at a given point in time
type PriceSnapshot struct {
	Pair  github_com_archway_network_archway_x_oracle_asset.Pair `protobuf:"bytes,1,opt,name=pair,proto3,customtype=github.com/archway-network/archway/x/oracle/asset.Pair" json:"pair" yaml:"pair"`
	Price cosmossdk_io_math.LegacyDec                            `protobuf:"bytes,2,opt,name=price,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"price"`
	// milliseconds since unix epoch
	TimestampMs int64 `protobuf:"varint,3,opt,name=timestamp_ms,json=timestampMs,proto3" json:"timestamp_ms,omitempty"`
}

func (m *PriceSnapshot) Reset()         { *m = PriceSnapshot{} }
func (m *PriceSnapshot) String() string { return proto.CompactTextString(m) }
func (*PriceSnapshot) ProtoMessage()    {}
func (*PriceSnapshot) Descriptor() ([]byte, []int) {
	return fileDescriptor_f86ab5cb19be0ee7, []int{0}
}
func (m *PriceSnapshot) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PriceSnapshot) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PriceSnapshot.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PriceSnapshot) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PriceSnapshot.Merge(m, src)
}
func (m *PriceSnapshot) XXX_Size() int {
	return m.Size()
}
func (m *PriceSnapshot) XXX_DiscardUnknown() {
	xxx_messageInfo_PriceSnapshot.DiscardUnknown(m)
}

var xxx_messageInfo_PriceSnapshot proto.InternalMessageInfo

func (m *PriceSnapshot) GetTimestampMs() int64 {
	if m != nil {
		return m.TimestampMs
	}
	return 0
}

func init() {
	proto.RegisterType((*PriceSnapshot)(nil), "archway.oracle.v1.PriceSnapshot")
}

func init() { proto.RegisterFile("archway/oracle/v1/state.proto", fileDescriptor_f86ab5cb19be0ee7) }

var fileDescriptor_f86ab5cb19be0ee7 = []byte{
	// 327 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x90, 0xbd, 0x4e, 0xf3, 0x30,
	0x14, 0x86, 0x93, 0xaf, 0x1f, 0x48, 0xa4, 0x30, 0x10, 0x31, 0x54, 0x05, 0xd2, 0x52, 0x96, 0x2e,
	0xc4, 0x8a, 0x90, 0x90, 0x60, 0xac, 0xd8, 0xa0, 0x52, 0x55, 0x36, 0x16, 0x74, 0x62, 0xac, 0xc4,
	0x6a, 0x9d, 0x63, 0xd9, 0x87, 0x96, 0xdc, 0x05, 0x97, 0xd5, 0xb1, 0x23, 0x62, 0xa8, 0x50, 0x7b,
	0x07, 0x5c, 0x01, 0xca, 0x4f, 0x59, 0x98, 0xd8, 0x6c, 0x3f, 0xf6, 0xfb, 0xbc, 0x3e, 0xde, 0x29,
	0x18, 0x9e, 0xce, 0x21, 0x67, 0x68, 0x80, 0x4f, 0x05, 0x9b, 0x45, 0xcc, 0x12, 0x90, 0x08, 0xb5,
	0x41, 0x42, 0xff, 0xb0, 0xc6, 0x61, 0x85, 0xc3, 0x59, 0xd4, 0x3e, 0x4a, 0x30, 0xc1, 0x92, 0xb2,
	0x62, 0x55, 0x5d, 0x6c, 0x9f, 0x24, 0x88, 0xc9, 0x54, 0x30, 0xd0, 0x92, 0x41, 0x96, 0x21, 0x01,
	0x49, 0xcc, 0x6c, 0x4d, 0x83, 0xdf, 0x96, 0x3a, 0xb0, 0xe6, 0x1c, 0xad, 0x42, 0xcb, 0x62, 0xb0,
	0x05, 0x8c, 0x05, 0x41, 0xc4, 0x38, 0xca, 0xac, 0xe2, 0xbd, 0xa5, 0xeb, 0x1d, 0x8c, 0x8c, 0xe4,
	0xe2, 0x21, 0x03, 0x6d, 0x53, 0x24, 0x1f, 0xbc, 0xff, 0x1a, 0xa4, 0x69, 0xb9, 0x5d, 0xb7, 0xbf,
	0x37, 0x18, 0x2e, 0x56, 0x1d, 0xe7, 0x63, 0xd5, 0xb9, 0x4a, 0x24, 0xa5, 0x2f, 0x71, 0xc8, 0x51,
	0xb1, 0x5a, 0x79, 0x91, 0x09, 0x9a, 0xa3, 0x99, 0x6c, 0xf7, 0xec, 0x75, 0x5b, 0x02, 0xac, 0x15,
	0x14, 0x8e, 0x40, 0x9a, 0xaf, 0x55, 0xa7, 0x99, 0x83, 0x9a, 0xde, 0xf4, 0x8a, 0xcc, 0xde, 0xb8,
	0x8c, 0xf6, 0xaf, 0xbd, 0x1d, 0x5d, 0x38, 0x5b, 0xff, 0x4a, 0xc7, 0x79, 0xed, 0x38, 0xae, 0xba,
	0xda, 0xe7, 0x49, 0x28, 0x91, 0x29, 0xa0, 0x34, 0xbc, 0x17, 0x09, 0xf0, 0xfc, 0x56, 0xf0, 0x71,
	0xf5, 0xc2, 0x3f, 0xf3, 0xf6, 0x49, 0x2a, 0x61, 0x09, 0x94, 0x7e, 0x52, 0xb6, 0xd5, 0xe8, 0xba,
	0xfd, 0xc6, 0xb8, 0xf9, 0x73, 0x36, 0xb4, 0x83, 0xbb, 0xc5, 0x3a, 0x70, 0x97, 0xeb, 0xc0, 0xfd,
	0x5c, 0x07, 0xee, 0xdb, 0x26, 0x70, 0x96, 0x9b, 0xc0, 0x79, 0xdf, 0x04, 0xce, 0x63, 0xf4, 0x97,
	0x4f, 0x50, 0xae, 0x85, 0x8d, 0x77, 0xcb, 0x31, 0x5d, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0x3c,
	0x7a, 0xa3, 0xe3, 0xce, 0x01, 0x00, 0x00,
}

func (m *PriceSnapshot) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PriceSnapshot) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PriceSnapshot) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.TimestampMs != 0 {
		i = encodeVarintState(dAtA, i, uint64(m.TimestampMs))
		i--
		dAtA[i] = 0x18
	}
	{
		size := m.Price.Size()
		i -= size
		if _, err := m.Price.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintState(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size := m.Pair.Size()
		i -= size
		if _, err := m.Pair.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintState(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintState(dAtA []byte, offset int, v uint64) int {
	offset -= sovState(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PriceSnapshot) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Pair.Size()
	n += 1 + l + sovState(uint64(l))
	l = m.Price.Size()
	n += 1 + l + sovState(uint64(l))
	if m.TimestampMs != 0 {
		n += 1 + sovState(uint64(m.TimestampMs))
	}
	return n
}

func sovState(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozState(x uint64) (n int) {
	return sovState(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PriceSnapshot) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: PriceSnapshot: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PriceSnapshot: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pair", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Pair.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Price", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Price.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimestampMs", wireType)
			}
			m.TimestampMs = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TimestampMs |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthState
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
func skipState(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
				return 0, ErrInvalidLengthState
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupState
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthState
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthState        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowState          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupState = fmt.Errorf("proto: unexpected end of group")
)
