// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: indexer/info/query.proto

package types

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// QueryVersionRequest is the request type for the Query/Versions RPC method
type QueryVersionRequest struct {
}

func (m *QueryVersionRequest) Reset()         { *m = QueryVersionRequest{} }
func (m *QueryVersionRequest) String() string { return proto.CompactTextString(m) }
func (*QueryVersionRequest) ProtoMessage()    {}
func (*QueryVersionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_81019926f3a532d0, []int{0}
}
func (m *QueryVersionRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryVersionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVersionRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryVersionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVersionRequest.Merge(m, src)
}
func (m *QueryVersionRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryVersionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVersionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVersionRequest proto.InternalMessageInfo

// QueryVersionResponse is the response type for the Query/Versions RPC method
type QueryVersionResponse struct {
	Versions []*SubmoduleVersion `protobuf:"bytes,1,rep,name=versions,proto3" json:"versions,omitempty"`
}

func (m *QueryVersionResponse) Reset()         { *m = QueryVersionResponse{} }
func (m *QueryVersionResponse) String() string { return proto.CompactTextString(m) }
func (*QueryVersionResponse) ProtoMessage()    {}
func (*QueryVersionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_81019926f3a532d0, []int{1}
}
func (m *QueryVersionResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryVersionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVersionResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryVersionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVersionResponse.Merge(m, src)
}
func (m *QueryVersionResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryVersionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVersionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVersionResponse proto.InternalMessageInfo

func (m *QueryVersionResponse) GetVersions() []*SubmoduleVersion {
	if m != nil {
		return m.Versions
	}
	return nil
}

// QueryVMTypeRequest is the request type for the Query/VMType RPC method
type QueryVMTypeRequest struct {
}

func (m *QueryVMTypeRequest) Reset()         { *m = QueryVMTypeRequest{} }
func (m *QueryVMTypeRequest) String() string { return proto.CompactTextString(m) }
func (*QueryVMTypeRequest) ProtoMessage()    {}
func (*QueryVMTypeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_81019926f3a532d0, []int{2}
}
func (m *QueryVMTypeRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryVMTypeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVMTypeRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryVMTypeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVMTypeRequest.Merge(m, src)
}
func (m *QueryVMTypeRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryVMTypeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVMTypeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVMTypeRequest proto.InternalMessageInfo

// QueryVMTypeResponse is the response type for the Query/VMType RPC method
type QueryVMTypeResponse struct {
	Vmtype string `protobuf:"bytes,1,opt,name=vmtype,proto3" json:"vmtype,omitempty"`
}

func (m *QueryVMTypeResponse) Reset()         { *m = QueryVMTypeResponse{} }
func (m *QueryVMTypeResponse) String() string { return proto.CompactTextString(m) }
func (*QueryVMTypeResponse) ProtoMessage()    {}
func (*QueryVMTypeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_81019926f3a532d0, []int{3}
}
func (m *QueryVMTypeResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryVMTypeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVMTypeResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryVMTypeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVMTypeResponse.Merge(m, src)
}
func (m *QueryVMTypeResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryVMTypeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVMTypeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVMTypeResponse proto.InternalMessageInfo

func (m *QueryVMTypeResponse) GetVmtype() string {
	if m != nil {
		return m.Vmtype
	}
	return ""
}

func init() {
	proto.RegisterType((*QueryVersionRequest)(nil), "indexer.info.QueryVersionRequest")
	proto.RegisterType((*QueryVersionResponse)(nil), "indexer.info.QueryVersionResponse")
	proto.RegisterType((*QueryVMTypeRequest)(nil), "indexer.info.QueryVMTypeRequest")
	proto.RegisterType((*QueryVMTypeResponse)(nil), "indexer.info.QueryVMTypeResponse")
}

func init() { proto.RegisterFile("indexer/info/query.proto", fileDescriptor_81019926f3a532d0) }

var fileDescriptor_81019926f3a532d0 = []byte{
	// 332 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x6b, 0x10, 0x55, 0x31, 0x48, 0x80, 0x29, 0x10, 0x45, 0xc8, 0x6a, 0x33, 0x75, 0x69,
	0x2c, 0x95, 0x8d, 0x91, 0x19, 0x06, 0x0a, 0x62, 0x60, 0x4b, 0xa8, 0x5b, 0xac, 0x26, 0x76, 0x1a,
	0x3b, 0x51, 0xb3, 0xf2, 0x04, 0x48, 0xbc, 0x14, 0x63, 0x25, 0x16, 0x06, 0x06, 0x94, 0xf0, 0x20,
	0x28, 0x89, 0x03, 0x89, 0x54, 0x75, 0xb3, 0xef, 0x7e, 0xff, 0xdf, 0x7f, 0x3e, 0x68, 0x30, 0x3e,
	0xa1, 0x4b, 0x1a, 0x12, 0xc6, 0xa7, 0x82, 0x2c, 0x22, 0x1a, 0x26, 0x76, 0x10, 0x0a, 0x25, 0xd0,
	0xbe, 0xee, 0xd8, 0x79, 0xc7, 0x3c, 0x9f, 0x09, 0x31, 0xf3, 0x28, 0x71, 0x02, 0x46, 0x1c, 0xce,
	0x85, 0x72, 0x14, 0x13, 0x5c, 0x96, 0x5a, 0xb3, 0xe9, 0xa2, 0x92, 0x80, 0xea, 0x8e, 0x75, 0x02,
	0x8f, 0x6f, 0x73, 0xd3, 0x07, 0x1a, 0x4a, 0x26, 0xf8, 0x98, 0x2e, 0x22, 0x2a, 0x95, 0x35, 0x86,
	0xdd, 0x66, 0x59, 0x06, 0x82, 0x4b, 0x8a, 0x2e, 0x61, 0x27, 0x2e, 0x4b, 0xd2, 0x00, 0xbd, 0xed,
	0xc1, 0xde, 0x08, 0xdb, 0xf5, 0x1c, 0xf6, 0x5d, 0xe4, 0xfa, 0x62, 0x12, 0x79, 0xb4, 0x7a, 0xf9,
	0xa7, 0xb7, 0xba, 0x10, 0x95, 0x9e, 0x37, 0xf7, 0x49, 0x40, 0x2b, 0xd2, 0xb0, 0x0a, 0xa0, 0xab,
	0x1a, 0x74, 0x0a, 0xdb, 0xb1, 0x9f, 0x07, 0x35, 0x40, 0x0f, 0x0c, 0x76, 0xc7, 0xfa, 0x36, 0xfa,
	0x02, 0x70, 0xa7, 0xd0, 0xa3, 0x39, 0xec, 0x68, 0x86, 0x44, 0xfd, 0x66, 0x88, 0x35, 0x13, 0x99,
	0xd6, 0x26, 0x49, 0x09, 0xb5, 0x8c, 0x97, 0x8f, 0x9f, 0xb7, 0x2d, 0x84, 0x0e, 0x49, 0xf5, 0x5f,
	0x3a, 0x3c, 0x9a, 0xc2, 0x76, 0x19, 0x10, 0xf5, 0xd6, 0xf9, 0xd4, 0x27, 0x32, 0xfb, 0x1b, 0x14,
	0x1a, 0x74, 0x56, 0x80, 0x8e, 0xd0, 0xc1, 0x3f, 0xa8, 0x18, 0xef, 0xea, 0xfa, 0x3d, 0xc5, 0x60,
	0x95, 0x62, 0xf0, 0x9d, 0x62, 0xf0, 0x9a, 0xe1, 0xd6, 0x2a, 0xc3, 0xad, 0xcf, 0x0c, 0xb7, 0x1e,
	0x47, 0x33, 0xa6, 0x9e, 0x23, 0xd7, 0x7e, 0x12, 0x3e, 0x61, 0x9c, 0x29, 0xe6, 0x0c, 0x3d, 0xc7,
	0x95, 0x64, 0x1e, 0x57, 0x16, 0xcb, 0xda, 0xb9, 0x58, 0xb1, 0xdb, 0x2e, 0x76, 0x7c, 0xf1, 0x1b,
	0x00, 0x00, 0xff, 0xff, 0x25, 0x57, 0xdf, 0x02, 0x45, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// Version queries all the versions of the submodules
	Versions(ctx context.Context, in *QueryVersionRequest, opts ...grpc.CallOption) (*QueryVersionResponse, error)
	// VMType queries the type of the Minitia's VM
	VMType(ctx context.Context, in *QueryVMTypeRequest, opts ...grpc.CallOption) (*QueryVMTypeResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Versions(ctx context.Context, in *QueryVersionRequest, opts ...grpc.CallOption) (*QueryVersionResponse, error) {
	out := new(QueryVersionResponse)
	err := c.cc.Invoke(ctx, "/indexer.info.Query/Versions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) VMType(ctx context.Context, in *QueryVMTypeRequest, opts ...grpc.CallOption) (*QueryVMTypeResponse, error) {
	out := new(QueryVMTypeResponse)
	err := c.cc.Invoke(ctx, "/indexer.info.Query/VMType", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Version queries all the versions of the submodules
	Versions(context.Context, *QueryVersionRequest) (*QueryVersionResponse, error)
	// VMType queries the type of the Minitia's VM
	VMType(context.Context, *QueryVMTypeRequest) (*QueryVMTypeResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Versions(ctx context.Context, req *QueryVersionRequest) (*QueryVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Versions not implemented")
}
func (*UnimplementedQueryServer) VMType(ctx context.Context, req *QueryVMTypeRequest) (*QueryVMTypeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VMType not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Versions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Versions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/indexer.info.Query/Versions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Versions(ctx, req.(*QueryVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_VMType_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryVMTypeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).VMType(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/indexer.info.Query/VMType",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).VMType(ctx, req.(*QueryVMTypeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "indexer.info.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Versions",
			Handler:    _Query_Versions_Handler,
		},
		{
			MethodName: "VMType",
			Handler:    _Query_VMType_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "indexer/info/query.proto",
}

func (m *QueryVersionRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVersionRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVersionRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryVersionResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVersionResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVersionResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Versions) > 0 {
		for iNdEx := len(m.Versions) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Versions[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryVMTypeRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVMTypeRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVMTypeRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryVMTypeResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVMTypeResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVMTypeResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Vmtype) > 0 {
		i -= len(m.Vmtype)
		copy(dAtA[i:], m.Vmtype)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Vmtype)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryVersionRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryVersionResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Versions) > 0 {
		for _, e := range m.Versions {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryVMTypeRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryVMTypeResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Vmtype)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryVersionRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryVersionRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVersionRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryVersionResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryVersionResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVersionResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Versions", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Versions = append(m.Versions, &SubmoduleVersion{})
			if err := m.Versions[len(m.Versions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryVMTypeRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryVMTypeRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVMTypeRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryVMTypeResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryVMTypeResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVMTypeResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Vmtype", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Vmtype = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)