// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.19.4
// source: service_dns.proto

package pb

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// 查找问题
type FindAllDNSIssuesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeClusterId int64 `protobuf:"varint,1,opt,name=nodeClusterId,proto3" json:"nodeClusterId,omitempty"`
}

func (x *FindAllDNSIssuesRequest) Reset() {
	*x = FindAllDNSIssuesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_dns_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindAllDNSIssuesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindAllDNSIssuesRequest) ProtoMessage() {}

func (x *FindAllDNSIssuesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_dns_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindAllDNSIssuesRequest.ProtoReflect.Descriptor instead.
func (*FindAllDNSIssuesRequest) Descriptor() ([]byte, []int) {
	return file_service_dns_proto_rawDescGZIP(), []int{0}
}

func (x *FindAllDNSIssuesRequest) GetNodeClusterId() int64 {
	if x != nil {
		return x.NodeClusterId
	}
	return 0
}

type FindAllDNSIssuesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Issues []*DNSIssue `protobuf:"bytes,1,rep,name=issues,proto3" json:"issues,omitempty"`
}

func (x *FindAllDNSIssuesResponse) Reset() {
	*x = FindAllDNSIssuesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_dns_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindAllDNSIssuesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindAllDNSIssuesResponse) ProtoMessage() {}

func (x *FindAllDNSIssuesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_dns_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindAllDNSIssuesResponse.ProtoReflect.Descriptor instead.
func (*FindAllDNSIssuesResponse) Descriptor() ([]byte, []int) {
	return file_service_dns_proto_rawDescGZIP(), []int{1}
}

func (x *FindAllDNSIssuesResponse) GetIssues() []*DNSIssue {
	if x != nil {
		return x.Issues
	}
	return nil
}

var File_service_dns_proto protoreflect.FileDescriptor

var file_service_dns_proto_rawDesc = []byte{
	0x0a, 0x11, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x1c, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x64, 0x6e, 0x73, 0x5f, 0x69, 0x73, 0x73, 0x75, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3f, 0x0a, 0x17, 0x46, 0x69, 0x6e, 0x64, 0x41, 0x6c, 0x6c,
	0x44, 0x4e, 0x53, 0x49, 0x73, 0x73, 0x75, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x24, 0x0a, 0x0d, 0x6e, 0x6f, 0x64, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x6e, 0x6f, 0x64, 0x65, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x49, 0x64, 0x22, 0x40, 0x0a, 0x18, 0x46, 0x69, 0x6e, 0x64, 0x41, 0x6c,
	0x6c, 0x44, 0x4e, 0x53, 0x49, 0x73, 0x73, 0x75, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x24, 0x0a, 0x06, 0x69, 0x73, 0x73, 0x75, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x62, 0x2e, 0x44, 0x4e, 0x53, 0x49, 0x73, 0x73, 0x75, 0x65,
	0x52, 0x06, 0x69, 0x73, 0x73, 0x75, 0x65, 0x73, 0x32, 0x5b, 0x0a, 0x0a, 0x44, 0x4e, 0x53, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4d, 0x0a, 0x10, 0x66, 0x69, 0x6e, 0x64, 0x41, 0x6c,
	0x6c, 0x44, 0x4e, 0x53, 0x49, 0x73, 0x73, 0x75, 0x65, 0x73, 0x12, 0x1b, 0x2e, 0x70, 0x62, 0x2e,
	0x46, 0x69, 0x6e, 0x64, 0x41, 0x6c, 0x6c, 0x44, 0x4e, 0x53, 0x49, 0x73, 0x73, 0x75, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6e,
	0x64, 0x41, 0x6c, 0x6c, 0x44, 0x4e, 0x53, 0x49, 0x73, 0x73, 0x75, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_service_dns_proto_rawDescOnce sync.Once
	file_service_dns_proto_rawDescData = file_service_dns_proto_rawDesc
)

func file_service_dns_proto_rawDescGZIP() []byte {
	file_service_dns_proto_rawDescOnce.Do(func() {
		file_service_dns_proto_rawDescData = protoimpl.X.CompressGZIP(file_service_dns_proto_rawDescData)
	})
	return file_service_dns_proto_rawDescData
}

var file_service_dns_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_service_dns_proto_goTypes = []interface{}{
	(*FindAllDNSIssuesRequest)(nil),  // 0: pb.FindAllDNSIssuesRequest
	(*FindAllDNSIssuesResponse)(nil), // 1: pb.FindAllDNSIssuesResponse
	(*DNSIssue)(nil),                 // 2: pb.DNSIssue
}
var file_service_dns_proto_depIdxs = []int32{
	2, // 0: pb.FindAllDNSIssuesResponse.issues:type_name -> pb.DNSIssue
	0, // 1: pb.DNSService.findAllDNSIssues:input_type -> pb.FindAllDNSIssuesRequest
	1, // 2: pb.DNSService.findAllDNSIssues:output_type -> pb.FindAllDNSIssuesResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_service_dns_proto_init() }
func file_service_dns_proto_init() {
	if File_service_dns_proto != nil {
		return
	}
	file_models_model_dns_issue_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_service_dns_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindAllDNSIssuesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_service_dns_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindAllDNSIssuesResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_service_dns_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_service_dns_proto_goTypes,
		DependencyIndexes: file_service_dns_proto_depIdxs,
		MessageInfos:      file_service_dns_proto_msgTypes,
	}.Build()
	File_service_dns_proto = out.File
	file_service_dns_proto_rawDesc = nil
	file_service_dns_proto_goTypes = nil
	file_service_dns_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// DNSServiceClient is the client API for DNSService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DNSServiceClient interface {
	// 查找问题
	FindAllDNSIssues(ctx context.Context, in *FindAllDNSIssuesRequest, opts ...grpc.CallOption) (*FindAllDNSIssuesResponse, error)
}

type dNSServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDNSServiceClient(cc grpc.ClientConnInterface) DNSServiceClient {
	return &dNSServiceClient{cc}
}

func (c *dNSServiceClient) FindAllDNSIssues(ctx context.Context, in *FindAllDNSIssuesRequest, opts ...grpc.CallOption) (*FindAllDNSIssuesResponse, error) {
	out := new(FindAllDNSIssuesResponse)
	err := c.cc.Invoke(ctx, "/pb.DNSService/findAllDNSIssues", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DNSServiceServer is the server API for DNSService service.
type DNSServiceServer interface {
	// 查找问题
	FindAllDNSIssues(context.Context, *FindAllDNSIssuesRequest) (*FindAllDNSIssuesResponse, error)
}

// UnimplementedDNSServiceServer can be embedded to have forward compatible implementations.
type UnimplementedDNSServiceServer struct {
}

func (*UnimplementedDNSServiceServer) FindAllDNSIssues(context.Context, *FindAllDNSIssuesRequest) (*FindAllDNSIssuesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindAllDNSIssues not implemented")
}

func RegisterDNSServiceServer(s *grpc.Server, srv DNSServiceServer) {
	s.RegisterService(&_DNSService_serviceDesc, srv)
}

func _DNSService_FindAllDNSIssues_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindAllDNSIssuesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DNSServiceServer).FindAllDNSIssues(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.DNSService/FindAllDNSIssues",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DNSServiceServer).FindAllDNSIssues(ctx, req.(*FindAllDNSIssuesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DNSService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.DNSService",
	HandlerType: (*DNSServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "findAllDNSIssues",
			Handler:    _DNSService_FindAllDNSIssues_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service_dns.proto",
}