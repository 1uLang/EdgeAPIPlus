// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.19.4
// source: service_message_media.proto

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

// 获取所有支持的媒介
type FindAllMessageMediasRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FindAllMessageMediasRequest) Reset() {
	*x = FindAllMessageMediasRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_message_media_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindAllMessageMediasRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindAllMessageMediasRequest) ProtoMessage() {}

func (x *FindAllMessageMediasRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_message_media_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindAllMessageMediasRequest.ProtoReflect.Descriptor instead.
func (*FindAllMessageMediasRequest) Descriptor() ([]byte, []int) {
	return file_service_message_media_proto_rawDescGZIP(), []int{0}
}

type FindAllMessageMediasResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MessageMedias []*MessageMedia `protobuf:"bytes,1,rep,name=messageMedias,proto3" json:"messageMedias,omitempty"`
}

func (x *FindAllMessageMediasResponse) Reset() {
	*x = FindAllMessageMediasResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_message_media_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindAllMessageMediasResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindAllMessageMediasResponse) ProtoMessage() {}

func (x *FindAllMessageMediasResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_message_media_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindAllMessageMediasResponse.ProtoReflect.Descriptor instead.
func (*FindAllMessageMediasResponse) Descriptor() ([]byte, []int) {
	return file_service_message_media_proto_rawDescGZIP(), []int{1}
}

func (x *FindAllMessageMediasResponse) GetMessageMedias() []*MessageMedia {
	if x != nil {
		return x.MessageMedias
	}
	return nil
}

// 设置所有支持的媒介
type UpdateMessageMediasRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MessageMedias []*MessageMedia `protobuf:"bytes,2,rep,name=messageMedias,proto3" json:"messageMedias,omitempty"`
}

func (x *UpdateMessageMediasRequest) Reset() {
	*x = UpdateMessageMediasRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_message_media_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateMessageMediasRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateMessageMediasRequest) ProtoMessage() {}

func (x *UpdateMessageMediasRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_message_media_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateMessageMediasRequest.ProtoReflect.Descriptor instead.
func (*UpdateMessageMediasRequest) Descriptor() ([]byte, []int) {
	return file_service_message_media_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateMessageMediasRequest) GetMessageMedias() []*MessageMedia {
	if x != nil {
		return x.MessageMedias
	}
	return nil
}

// 发送媒介信息
type SendMediaMessageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MediaType   string `protobuf:"bytes,1,opt,name=mediaType,proto3" json:"mediaType,omitempty"`     // 媒介类型
	OptionsJSON []byte `protobuf:"bytes,2,opt,name=optionsJSON,proto3" json:"optionsJSON,omitempty"` // 媒介参数
	User        string `protobuf:"bytes,3,opt,name=user,proto3" json:"user,omitempty"`               // 接收用户
	Subject     string `protobuf:"bytes,4,opt,name=subject,proto3" json:"subject,omitempty"`         // 标题
	Body        string `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`               // 内容
}

func (x *SendMediaMessageRequest) Reset() {
	*x = SendMediaMessageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_message_media_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendMediaMessageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendMediaMessageRequest) ProtoMessage() {}

func (x *SendMediaMessageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_message_media_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendMediaMessageRequest.ProtoReflect.Descriptor instead.
func (*SendMediaMessageRequest) Descriptor() ([]byte, []int) {
	return file_service_message_media_proto_rawDescGZIP(), []int{3}
}

func (x *SendMediaMessageRequest) GetMediaType() string {
	if x != nil {
		return x.MediaType
	}
	return ""
}

func (x *SendMediaMessageRequest) GetOptionsJSON() []byte {
	if x != nil {
		return x.OptionsJSON
	}
	return nil
}

func (x *SendMediaMessageRequest) GetUser() string {
	if x != nil {
		return x.User
	}
	return ""
}

func (x *SendMediaMessageRequest) GetSubject() string {
	if x != nil {
		return x.Subject
	}
	return ""
}

func (x *SendMediaMessageRequest) GetBody() string {
	if x != nil {
		return x.Body
	}
	return ""
}

var File_service_message_media_proto protoreflect.FileDescriptor

var file_service_message_media_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70,
	0x62, 0x1a, 0x20, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x72, 0x70, 0x63, 0x5f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1d,
	0x0a, 0x1b, 0x46, 0x69, 0x6e, 0x64, 0x41, 0x6c, 0x6c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x4d, 0x65, 0x64, 0x69, 0x61, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x56, 0x0a,
	0x1c, 0x46, 0x69, 0x6e, 0x64, 0x41, 0x6c, 0x6c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d,
	0x65, 0x64, 0x69, 0x61, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x36, 0x0a,
	0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d,
	0x65, 0x64, 0x69, 0x61, 0x73, 0x22, 0x54, 0x0a, 0x1a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x36, 0x0a, 0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65,
	0x64, 0x69, 0x61, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x70, 0x62, 0x2e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x0d, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x73, 0x22, 0x9b, 0x01, 0x0a, 0x17,
	0x53, 0x65, 0x6e, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6d, 0x65, 0x64, 0x69, 0x61,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d, 0x65, 0x64, 0x69,
	0x61, 0x54, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x4a, 0x53, 0x4f, 0x4e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x4a, 0x53, 0x4f, 0x4e, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x73,
	0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x75,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x32, 0xf8, 0x01, 0x0a, 0x13, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x59, 0x0a, 0x14, 0x66, 0x69, 0x6e, 0x64, 0x41, 0x6c, 0x6c, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x73, 0x12, 0x1f, 0x2e, 0x70, 0x62, 0x2e, 0x46,
	0x69, 0x6e, 0x64, 0x41, 0x6c, 0x6c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64,
	0x69, 0x61, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x70, 0x62, 0x2e,
	0x46, 0x69, 0x6e, 0x64, 0x41, 0x6c, 0x6c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65,
	0x64, 0x69, 0x61, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x13,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64,
	0x69, 0x61, 0x73, 0x12, 0x1e, 0x2e, 0x70, 0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x50, 0x43, 0x53, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x12, 0x3f, 0x0a, 0x10, 0x73, 0x65, 0x6e, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x6e,
	0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x50, 0x43, 0x53, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_service_message_media_proto_rawDescOnce sync.Once
	file_service_message_media_proto_rawDescData = file_service_message_media_proto_rawDesc
)

func file_service_message_media_proto_rawDescGZIP() []byte {
	file_service_message_media_proto_rawDescOnce.Do(func() {
		file_service_message_media_proto_rawDescData = protoimpl.X.CompressGZIP(file_service_message_media_proto_rawDescData)
	})
	return file_service_message_media_proto_rawDescData
}

var file_service_message_media_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_service_message_media_proto_goTypes = []interface{}{
	(*FindAllMessageMediasRequest)(nil),  // 0: pb.FindAllMessageMediasRequest
	(*FindAllMessageMediasResponse)(nil), // 1: pb.FindAllMessageMediasResponse
	(*UpdateMessageMediasRequest)(nil),   // 2: pb.UpdateMessageMediasRequest
	(*SendMediaMessageRequest)(nil),      // 3: pb.SendMediaMessageRequest
	(*MessageMedia)(nil),                 // 4: pb.MessageMedia
	(*RPCSuccess)(nil),                   // 5: pb.RPCSuccess
}
var file_service_message_media_proto_depIdxs = []int32{
	4, // 0: pb.FindAllMessageMediasResponse.messageMedias:type_name -> pb.MessageMedia
	4, // 1: pb.UpdateMessageMediasRequest.messageMedias:type_name -> pb.MessageMedia
	0, // 2: pb.MessageMediaService.findAllMessageMedias:input_type -> pb.FindAllMessageMediasRequest
	2, // 3: pb.MessageMediaService.updateMessageMedias:input_type -> pb.UpdateMessageMediasRequest
	3, // 4: pb.MessageMediaService.sendMediaMessage:input_type -> pb.SendMediaMessageRequest
	1, // 5: pb.MessageMediaService.findAllMessageMedias:output_type -> pb.FindAllMessageMediasResponse
	5, // 6: pb.MessageMediaService.updateMessageMedias:output_type -> pb.RPCSuccess
	5, // 7: pb.MessageMediaService.sendMediaMessage:output_type -> pb.RPCSuccess
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_service_message_media_proto_init() }
func file_service_message_media_proto_init() {
	if File_service_message_media_proto != nil {
		return
	}
	file_models_model_message_media_proto_init()
	file_models_rpc_messages_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_service_message_media_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindAllMessageMediasRequest); i {
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
		file_service_message_media_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindAllMessageMediasResponse); i {
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
		file_service_message_media_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateMessageMediasRequest); i {
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
		file_service_message_media_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendMediaMessageRequest); i {
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
			RawDescriptor: file_service_message_media_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_service_message_media_proto_goTypes,
		DependencyIndexes: file_service_message_media_proto_depIdxs,
		MessageInfos:      file_service_message_media_proto_msgTypes,
	}.Build()
	File_service_message_media_proto = out.File
	file_service_message_media_proto_rawDesc = nil
	file_service_message_media_proto_goTypes = nil
	file_service_message_media_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// MessageMediaServiceClient is the client API for MessageMediaService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MessageMediaServiceClient interface {
	// 获取所有支持的媒介
	FindAllMessageMedias(ctx context.Context, in *FindAllMessageMediasRequest, opts ...grpc.CallOption) (*FindAllMessageMediasResponse, error)
	// 设置所有支持的媒介
	UpdateMessageMedias(ctx context.Context, in *UpdateMessageMediasRequest, opts ...grpc.CallOption) (*RPCSuccess, error)
	// 发送媒介信息
	SendMediaMessage(ctx context.Context, in *SendMediaMessageRequest, opts ...grpc.CallOption) (*RPCSuccess, error)
}

type messageMediaServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMessageMediaServiceClient(cc grpc.ClientConnInterface) MessageMediaServiceClient {
	return &messageMediaServiceClient{cc}
}

func (c *messageMediaServiceClient) FindAllMessageMedias(ctx context.Context, in *FindAllMessageMediasRequest, opts ...grpc.CallOption) (*FindAllMessageMediasResponse, error) {
	out := new(FindAllMessageMediasResponse)
	err := c.cc.Invoke(ctx, "/pb.MessageMediaService/findAllMessageMedias", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageMediaServiceClient) UpdateMessageMedias(ctx context.Context, in *UpdateMessageMediasRequest, opts ...grpc.CallOption) (*RPCSuccess, error) {
	out := new(RPCSuccess)
	err := c.cc.Invoke(ctx, "/pb.MessageMediaService/updateMessageMedias", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageMediaServiceClient) SendMediaMessage(ctx context.Context, in *SendMediaMessageRequest, opts ...grpc.CallOption) (*RPCSuccess, error) {
	out := new(RPCSuccess)
	err := c.cc.Invoke(ctx, "/pb.MessageMediaService/sendMediaMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MessageMediaServiceServer is the server API for MessageMediaService service.
type MessageMediaServiceServer interface {
	// 获取所有支持的媒介
	FindAllMessageMedias(context.Context, *FindAllMessageMediasRequest) (*FindAllMessageMediasResponse, error)
	// 设置所有支持的媒介
	UpdateMessageMedias(context.Context, *UpdateMessageMediasRequest) (*RPCSuccess, error)
	// 发送媒介信息
	SendMediaMessage(context.Context, *SendMediaMessageRequest) (*RPCSuccess, error)
}

// UnimplementedMessageMediaServiceServer can be embedded to have forward compatible implementations.
type UnimplementedMessageMediaServiceServer struct {
}

func (*UnimplementedMessageMediaServiceServer) FindAllMessageMedias(context.Context, *FindAllMessageMediasRequest) (*FindAllMessageMediasResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindAllMessageMedias not implemented")
}
func (*UnimplementedMessageMediaServiceServer) UpdateMessageMedias(context.Context, *UpdateMessageMediasRequest) (*RPCSuccess, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMessageMedias not implemented")
}
func (*UnimplementedMessageMediaServiceServer) SendMediaMessage(context.Context, *SendMediaMessageRequest) (*RPCSuccess, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMediaMessage not implemented")
}

func RegisterMessageMediaServiceServer(s *grpc.Server, srv MessageMediaServiceServer) {
	s.RegisterService(&_MessageMediaService_serviceDesc, srv)
}

func _MessageMediaService_FindAllMessageMedias_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindAllMessageMediasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageMediaServiceServer).FindAllMessageMedias(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.MessageMediaService/FindAllMessageMedias",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageMediaServiceServer).FindAllMessageMedias(ctx, req.(*FindAllMessageMediasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageMediaService_UpdateMessageMedias_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMessageMediasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageMediaServiceServer).UpdateMessageMedias(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.MessageMediaService/UpdateMessageMedias",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageMediaServiceServer).UpdateMessageMedias(ctx, req.(*UpdateMessageMediasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageMediaService_SendMediaMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMediaMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageMediaServiceServer).SendMediaMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.MessageMediaService/SendMediaMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageMediaServiceServer).SendMediaMessage(ctx, req.(*SendMediaMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _MessageMediaService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.MessageMediaService",
	HandlerType: (*MessageMediaServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "findAllMessageMedias",
			Handler:    _MessageMediaService_FindAllMessageMedias_Handler,
		},
		{
			MethodName: "updateMessageMedias",
			Handler:    _MessageMediaService_UpdateMessageMedias_Handler,
		},
		{
			MethodName: "sendMediaMessage",
			Handler:    _MessageMediaService_SendMediaMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service_message_media.proto",
}