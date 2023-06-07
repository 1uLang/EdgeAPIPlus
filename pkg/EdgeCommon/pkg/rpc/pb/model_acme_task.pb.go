// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.19.4
// source: models/model_acme_task.proto

package pb

import (
	proto "github.com/golang/protobuf/proto"
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

type ACMETask struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                int64        `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	IsOn              bool         `protobuf:"varint,2,opt,name=isOn,proto3" json:"isOn,omitempty"`
	DnsDomain         string       `protobuf:"bytes,3,opt,name=dnsDomain,proto3" json:"dnsDomain,omitempty"`
	Domains           []string     `protobuf:"bytes,4,rep,name=domains,proto3" json:"domains,omitempty"`
	CreatedAt         int64        `protobuf:"varint,5,opt,name=createdAt,proto3" json:"createdAt,omitempty"`
	AutoRenew         bool         `protobuf:"varint,6,opt,name=autoRenew,proto3" json:"autoRenew,omitempty"`
	AuthType          string       `protobuf:"bytes,7,opt,name=authType,proto3" json:"authType,omitempty"`
	AuthURL           string       `protobuf:"bytes,8,opt,name=authURL,proto3" json:"authURL,omitempty"`
	AcmeUser          *ACMEUser    `protobuf:"bytes,30,opt,name=acmeUser,proto3" json:"acmeUser,omitempty"`
	DnsProvider       *DNSProvider `protobuf:"bytes,31,opt,name=dnsProvider,proto3" json:"dnsProvider,omitempty"`
	SslCert           *SSLCert     `protobuf:"bytes,32,opt,name=sslCert,proto3" json:"sslCert,omitempty"`
	LatestACMETaskLog *ACMETaskLog `protobuf:"bytes,33,opt,name=latestACMETaskLog,proto3" json:"latestACMETaskLog,omitempty"`
}

func (x *ACMETask) Reset() {
	*x = ACMETask{}
	if protoimpl.UnsafeEnabled {
		mi := &file_models_model_acme_task_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ACMETask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ACMETask) ProtoMessage() {}

func (x *ACMETask) ProtoReflect() protoreflect.Message {
	mi := &file_models_model_acme_task_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ACMETask.ProtoReflect.Descriptor instead.
func (*ACMETask) Descriptor() ([]byte, []int) {
	return file_models_model_acme_task_proto_rawDescGZIP(), []int{0}
}

func (x *ACMETask) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ACMETask) GetIsOn() bool {
	if x != nil {
		return x.IsOn
	}
	return false
}

func (x *ACMETask) GetDnsDomain() string {
	if x != nil {
		return x.DnsDomain
	}
	return ""
}

func (x *ACMETask) GetDomains() []string {
	if x != nil {
		return x.Domains
	}
	return nil
}

func (x *ACMETask) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *ACMETask) GetAutoRenew() bool {
	if x != nil {
		return x.AutoRenew
	}
	return false
}

func (x *ACMETask) GetAuthType() string {
	if x != nil {
		return x.AuthType
	}
	return ""
}

func (x *ACMETask) GetAuthURL() string {
	if x != nil {
		return x.AuthURL
	}
	return ""
}

func (x *ACMETask) GetAcmeUser() *ACMEUser {
	if x != nil {
		return x.AcmeUser
	}
	return nil
}

func (x *ACMETask) GetDnsProvider() *DNSProvider {
	if x != nil {
		return x.DnsProvider
	}
	return nil
}

func (x *ACMETask) GetSslCert() *SSLCert {
	if x != nil {
		return x.SslCert
	}
	return nil
}

func (x *ACMETask) GetLatestACMETaskLog() *ACMETaskLog {
	if x != nil {
		return x.LatestACMETaskLog
	}
	return nil
}

var File_models_model_acme_task_proto protoreflect.FileDescriptor

var file_models_model_acme_task_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x61,
	0x63, 0x6d, 0x65, 0x5f, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02,
	0x70, 0x62, 0x1a, 0x1c, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x5f, 0x61, 0x63, 0x6d, 0x65, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x64,
	0x6e, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1b, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f,
	0x73, 0x73, 0x6c, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x61, 0x63, 0x6d,
	0x65, 0x5f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x9b, 0x03, 0x0a, 0x08, 0x41, 0x43, 0x4d, 0x45, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x69, 0x73, 0x4f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x69, 0x73, 0x4f,
	0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x6e, 0x73, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64, 0x6e, 0x73, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x12,
	0x18, 0x0a, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x75, 0x74, 0x6f, 0x52,
	0x65, 0x6e, 0x65, 0x77, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x61, 0x75, 0x74, 0x6f,
	0x52, 0x65, 0x6e, 0x65, 0x77, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x75, 0x74, 0x68, 0x54, 0x79, 0x70,
	0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x75, 0x74, 0x68, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x75, 0x74, 0x68, 0x55, 0x52, 0x4c, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x61, 0x75, 0x74, 0x68, 0x55, 0x52, 0x4c, 0x12, 0x28, 0x0a, 0x08, 0x61,
	0x63, 0x6d, 0x65, 0x55, 0x73, 0x65, 0x72, 0x18, 0x1e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e,
	0x70, 0x62, 0x2e, 0x41, 0x43, 0x4d, 0x45, 0x55, 0x73, 0x65, 0x72, 0x52, 0x08, 0x61, 0x63, 0x6d,
	0x65, 0x55, 0x73, 0x65, 0x72, 0x12, 0x31, 0x0a, 0x0b, 0x64, 0x6e, 0x73, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x18, 0x1f, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x62, 0x2e,
	0x44, 0x4e, 0x53, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x52, 0x0b, 0x64, 0x6e, 0x73,
	0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x12, 0x25, 0x0a, 0x07, 0x73, 0x73, 0x6c, 0x43,
	0x65, 0x72, 0x74, 0x18, 0x20, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x62, 0x2e, 0x53,
	0x53, 0x4c, 0x43, 0x65, 0x72, 0x74, 0x52, 0x07, 0x73, 0x73, 0x6c, 0x43, 0x65, 0x72, 0x74, 0x12,
	0x3d, 0x0a, 0x11, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x41, 0x43, 0x4d, 0x45, 0x54, 0x61, 0x73,
	0x6b, 0x4c, 0x6f, 0x67, 0x18, 0x21, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x62, 0x2e,
	0x41, 0x43, 0x4d, 0x45, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x52, 0x11, 0x6c, 0x61, 0x74,
	0x65, 0x73, 0x74, 0x41, 0x43, 0x4d, 0x45, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x42, 0x06,
	0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_models_model_acme_task_proto_rawDescOnce sync.Once
	file_models_model_acme_task_proto_rawDescData = file_models_model_acme_task_proto_rawDesc
)

func file_models_model_acme_task_proto_rawDescGZIP() []byte {
	file_models_model_acme_task_proto_rawDescOnce.Do(func() {
		file_models_model_acme_task_proto_rawDescData = protoimpl.X.CompressGZIP(file_models_model_acme_task_proto_rawDescData)
	})
	return file_models_model_acme_task_proto_rawDescData
}

var file_models_model_acme_task_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_models_model_acme_task_proto_goTypes = []interface{}{
	(*ACMETask)(nil),    // 0: pb.ACMETask
	(*ACMEUser)(nil),    // 1: pb.ACMEUser
	(*DNSProvider)(nil), // 2: pb.DNSProvider
	(*SSLCert)(nil),     // 3: pb.SSLCert
	(*ACMETaskLog)(nil), // 4: pb.ACMETaskLog
}
var file_models_model_acme_task_proto_depIdxs = []int32{
	1, // 0: pb.ACMETask.acmeUser:type_name -> pb.ACMEUser
	2, // 1: pb.ACMETask.dnsProvider:type_name -> pb.DNSProvider
	3, // 2: pb.ACMETask.sslCert:type_name -> pb.SSLCert
	4, // 3: pb.ACMETask.latestACMETaskLog:type_name -> pb.ACMETaskLog
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_models_model_acme_task_proto_init() }
func file_models_model_acme_task_proto_init() {
	if File_models_model_acme_task_proto != nil {
		return
	}
	file_models_model_acme_user_proto_init()
	file_models_model_dns_provider_proto_init()
	file_models_model_ssl_cert_proto_init()
	file_models_model_acme_task_log_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_models_model_acme_task_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ACMETask); i {
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
			RawDescriptor: file_models_model_acme_task_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_models_model_acme_task_proto_goTypes,
		DependencyIndexes: file_models_model_acme_task_proto_depIdxs,
		MessageInfos:      file_models_model_acme_task_proto_msgTypes,
	}.Build()
	File_models_model_acme_task_proto = out.File
	file_models_model_acme_task_proto_rawDesc = nil
	file_models_model_acme_task_proto_goTypes = nil
	file_models_model_acme_task_proto_depIdxs = nil
}