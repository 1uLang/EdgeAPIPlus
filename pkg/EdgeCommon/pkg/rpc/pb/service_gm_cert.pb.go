// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: service_gm_cert.proto

package pb

import (
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

// 创建证书
type CreateGMCertRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsOn         bool     `protobuf:"varint,1,opt,name=isOn,proto3" json:"isOn,omitempty"`
	UserId       int64    `protobuf:"varint,2,opt,name=userId,proto3" json:"userId,omitempty"` // 所属用户，仅管理员才能指定
	Name         string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description  string   `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	ServerName   string   `protobuf:"bytes,5,opt,name=serverName,proto3" json:"serverName,omitempty"`
	IsCA         bool     `protobuf:"varint,6,opt,name=isCA,proto3" json:"isCA,omitempty"`
	SignCertData []byte   `protobuf:"bytes,7,opt,name=signCertData,proto3" json:"signCertData,omitempty"`
	SignKeyData  []byte   `protobuf:"bytes,8,opt,name=signKeyData,proto3" json:"signKeyData,omitempty"`
	EncCertData  []byte   `protobuf:"bytes,9,opt,name=encCertData,proto3" json:"encCertData,omitempty"`
	EncKeyData   []byte   `protobuf:"bytes,10,opt,name=encKeyData,proto3" json:"encKeyData,omitempty"`
	TimeBeginAt  int64    `protobuf:"varint,11,opt,name=timeBeginAt,proto3" json:"timeBeginAt,omitempty"`
	TimeEndAt    int64    `protobuf:"varint,121,opt,name=timeEndAt,proto3" json:"timeEndAt,omitempty"`
	DnsNames     []string `protobuf:"bytes,13,rep,name=dnsNames,proto3" json:"dnsNames,omitempty"`
	CommonNames  []string `protobuf:"bytes,14,rep,name=commonNames,proto3" json:"commonNames,omitempty"`
}

func (x *CreateGMCertRequest) Reset() {
	*x = CreateGMCertRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateGMCertRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateGMCertRequest) ProtoMessage() {}

func (x *CreateGMCertRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateGMCertRequest.ProtoReflect.Descriptor instead.
func (*CreateGMCertRequest) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{0}
}

func (x *CreateGMCertRequest) GetIsOn() bool {
	if x != nil {
		return x.IsOn
	}
	return false
}

func (x *CreateGMCertRequest) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *CreateGMCertRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CreateGMCertRequest) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *CreateGMCertRequest) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

func (x *CreateGMCertRequest) GetIsCA() bool {
	if x != nil {
		return x.IsCA
	}
	return false
}

func (x *CreateGMCertRequest) GetSignCertData() []byte {
	if x != nil {
		return x.SignCertData
	}
	return nil
}

func (x *CreateGMCertRequest) GetSignKeyData() []byte {
	if x != nil {
		return x.SignKeyData
	}
	return nil
}

func (x *CreateGMCertRequest) GetEncCertData() []byte {
	if x != nil {
		return x.EncCertData
	}
	return nil
}

func (x *CreateGMCertRequest) GetEncKeyData() []byte {
	if x != nil {
		return x.EncKeyData
	}
	return nil
}

func (x *CreateGMCertRequest) GetTimeBeginAt() int64 {
	if x != nil {
		return x.TimeBeginAt
	}
	return 0
}

func (x *CreateGMCertRequest) GetTimeEndAt() int64 {
	if x != nil {
		return x.TimeEndAt
	}
	return 0
}

func (x *CreateGMCertRequest) GetDnsNames() []string {
	if x != nil {
		return x.DnsNames
	}
	return nil
}

func (x *CreateGMCertRequest) GetCommonNames() []string {
	if x != nil {
		return x.CommonNames
	}
	return nil
}

type CreateGMCertResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GmCertId int64 `protobuf:"varint,1,opt,name=gmCertId,proto3" json:"gmCertId,omitempty"`
}

func (x *CreateGMCertResponse) Reset() {
	*x = CreateGMCertResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateGMCertResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateGMCertResponse) ProtoMessage() {}

func (x *CreateGMCertResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateGMCertResponse.ProtoReflect.Descriptor instead.
func (*CreateGMCertResponse) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{1}
}

func (x *CreateGMCertResponse) GetGmCertId() int64 {
	if x != nil {
		return x.GmCertId
	}
	return 0
}

// 查找证书配置
type FindEnabledGMCertConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GmCertId int64 `protobuf:"varint,1,opt,name=gmCertId,proto3" json:"gmCertId,omitempty"`
}

func (x *FindEnabledGMCertConfigRequest) Reset() {
	*x = FindEnabledGMCertConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindEnabledGMCertConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindEnabledGMCertConfigRequest) ProtoMessage() {}

func (x *FindEnabledGMCertConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindEnabledGMCertConfigRequest.ProtoReflect.Descriptor instead.
func (*FindEnabledGMCertConfigRequest) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{2}
}

func (x *FindEnabledGMCertConfigRequest) GetGmCertId() int64 {
	if x != nil {
		return x.GmCertId
	}
	return 0
}

type FindEnabledGMCertConfigResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GmCertJSON []byte `protobuf:"bytes,1,opt,name=gmCertJSON,proto3" json:"gmCertJSON,omitempty"`
}

func (x *FindEnabledGMCertConfigResponse) Reset() {
	*x = FindEnabledGMCertConfigResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindEnabledGMCertConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindEnabledGMCertConfigResponse) ProtoMessage() {}

func (x *FindEnabledGMCertConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindEnabledGMCertConfigResponse.ProtoReflect.Descriptor instead.
func (*FindEnabledGMCertConfigResponse) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{3}
}

func (x *FindEnabledGMCertConfigResponse) GetGmCertJSON() []byte {
	if x != nil {
		return x.GmCertJSON
	}
	return nil
}

// 计算匹配的证书数量
type CountGMCertRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsAvailable  bool     `protobuf:"varint,1,opt,name=isAvailable,proto3" json:"isAvailable,omitempty"`   // 是否可用（在有效期内）
	IsExpired    bool     `protobuf:"varint,2,opt,name=isExpired,proto3" json:"isExpired,omitempty"`       // 是否已过期
	ExpiringDays int32    `protobuf:"varint,3,opt,name=expiringDays,proto3" json:"expiringDays,omitempty"` // 离过期日的天数
	Keyword      string   `protobuf:"bytes,4,opt,name=keyword,proto3" json:"keyword,omitempty"`            // 关键词
	UserId       int64    `protobuf:"varint,5,opt,name=userId,proto3" json:"userId,omitempty"`             // 用户ID
	Domains      []string `protobuf:"bytes,6,rep,name=domains,proto3" json:"domains,omitempty"`            // 搜索使用的域名列表
}

func (x *CountGMCertRequest) Reset() {
	*x = CountGMCertRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CountGMCertRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CountGMCertRequest) ProtoMessage() {}

func (x *CountGMCertRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CountGMCertRequest.ProtoReflect.Descriptor instead.
func (*CountGMCertRequest) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{4}
}

func (x *CountGMCertRequest) GetIsAvailable() bool {
	if x != nil {
		return x.IsAvailable
	}
	return false
}

func (x *CountGMCertRequest) GetIsExpired() bool {
	if x != nil {
		return x.IsExpired
	}
	return false
}

func (x *CountGMCertRequest) GetExpiringDays() int32 {
	if x != nil {
		return x.ExpiringDays
	}
	return 0
}

func (x *CountGMCertRequest) GetKeyword() string {
	if x != nil {
		return x.Keyword
	}
	return ""
}

func (x *CountGMCertRequest) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *CountGMCertRequest) GetDomains() []string {
	if x != nil {
		return x.Domains
	}
	return nil
}

// 列出单页匹配的证书
type ListGMCertsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsAvailable  bool     `protobuf:"varint,1,opt,name=isAvailable,proto3" json:"isAvailable,omitempty"`   // 是否可用（在有效期内）
	IsExpired    bool     `protobuf:"varint,2,opt,name=isExpired,proto3" json:"isExpired,omitempty"`       // 是否已过期
	ExpiringDays int32    `protobuf:"varint,3,opt,name=expiringDays,proto3" json:"expiringDays,omitempty"` // 离过期日的天数
	Keyword      string   `protobuf:"bytes,4,opt,name=keyword,proto3" json:"keyword,omitempty"`            // 关键词
	UserId       int64    `protobuf:"varint,5,opt,name=userId,proto3" json:"userId,omitempty"`             // 用户ID
	Domains      []string `protobuf:"bytes,6,rep,name=domains,proto3" json:"domains,omitempty"`            // 搜索使用的域名列表
	Offset       int64    `protobuf:"varint,7,opt,name=offset,proto3" json:"offset,omitempty"`             // 读取位置
	Size         int64    `protobuf:"varint,8,opt,name=size,proto3" json:"size,omitempty"`                 // 读取长度
}

func (x *ListGMCertsRequest) Reset() {
	*x = ListGMCertsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListGMCertsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListGMCertsRequest) ProtoMessage() {}

func (x *ListGMCertsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListGMCertsRequest.ProtoReflect.Descriptor instead.
func (*ListGMCertsRequest) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{5}
}

func (x *ListGMCertsRequest) GetIsAvailable() bool {
	if x != nil {
		return x.IsAvailable
	}
	return false
}

func (x *ListGMCertsRequest) GetIsExpired() bool {
	if x != nil {
		return x.IsExpired
	}
	return false
}

func (x *ListGMCertsRequest) GetExpiringDays() int32 {
	if x != nil {
		return x.ExpiringDays
	}
	return 0
}

func (x *ListGMCertsRequest) GetKeyword() string {
	if x != nil {
		return x.Keyword
	}
	return ""
}

func (x *ListGMCertsRequest) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *ListGMCertsRequest) GetDomains() []string {
	if x != nil {
		return x.Domains
	}
	return nil
}

func (x *ListGMCertsRequest) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *ListGMCertsRequest) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

type ListGMCertsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GmCertsJSON []byte `protobuf:"bytes,1,opt,name=gmCertsJSON,proto3" json:"gmCertsJSON,omitempty"`
}

func (x *ListGMCertsResponse) Reset() {
	*x = ListGMCertsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListGMCertsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListGMCertsResponse) ProtoMessage() {}

func (x *ListGMCertsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListGMCertsResponse.ProtoReflect.Descriptor instead.
func (*ListGMCertsResponse) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{6}
}

func (x *ListGMCertsResponse) GetGmCertsJSON() []byte {
	if x != nil {
		return x.GmCertsJSON
	}
	return nil
}

// 修改证书
type UpdateGMCertRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GmCertId     int64    `protobuf:"varint,1,opt,name=gmCertId,proto3" json:"gmCertId,omitempty"`
	IsOn         bool     `protobuf:"varint,2,opt,name=isOn,proto3" json:"isOn,omitempty"`
	Name         string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description  string   `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	ServerName   string   `protobuf:"bytes,5,opt,name=serverName,proto3" json:"serverName,omitempty"`
	SignCertData []byte   `protobuf:"bytes,6,opt,name=signCertData,proto3" json:"signCertData,omitempty"`
	SignKeyData  []byte   `protobuf:"bytes,7,opt,name=signKeyData,proto3" json:"signKeyData,omitempty"`
	EncCertData  []byte   `protobuf:"bytes,8,opt,name=encCertData,proto3" json:"encCertData,omitempty"`
	EncKeyData   []byte   `protobuf:"bytes,9,opt,name=encKeyData,proto3" json:"encKeyData,omitempty"`
	TimeBeginAt  int64    `protobuf:"varint,10,opt,name=timeBeginAt,proto3" json:"timeBeginAt,omitempty"`
	TimeEndAt    int64    `protobuf:"varint,11,opt,name=timeEndAt,proto3" json:"timeEndAt,omitempty"`
	DnsNames     []string `protobuf:"bytes,12,rep,name=dnsNames,proto3" json:"dnsNames,omitempty"`
	CommonNames  []string `protobuf:"bytes,13,rep,name=commonNames,proto3" json:"commonNames,omitempty"`
}

func (x *UpdateGMCertRequest) Reset() {
	*x = UpdateGMCertRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateGMCertRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateGMCertRequest) ProtoMessage() {}

func (x *UpdateGMCertRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateGMCertRequest.ProtoReflect.Descriptor instead.
func (*UpdateGMCertRequest) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateGMCertRequest) GetGmCertId() int64 {
	if x != nil {
		return x.GmCertId
	}
	return 0
}

func (x *UpdateGMCertRequest) GetIsOn() bool {
	if x != nil {
		return x.IsOn
	}
	return false
}

func (x *UpdateGMCertRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *UpdateGMCertRequest) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *UpdateGMCertRequest) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

func (x *UpdateGMCertRequest) GetSignCertData() []byte {
	if x != nil {
		return x.SignCertData
	}
	return nil
}

func (x *UpdateGMCertRequest) GetSignKeyData() []byte {
	if x != nil {
		return x.SignKeyData
	}
	return nil
}

func (x *UpdateGMCertRequest) GetEncCertData() []byte {
	if x != nil {
		return x.EncCertData
	}
	return nil
}

func (x *UpdateGMCertRequest) GetEncKeyData() []byte {
	if x != nil {
		return x.EncKeyData
	}
	return nil
}

func (x *UpdateGMCertRequest) GetTimeBeginAt() int64 {
	if x != nil {
		return x.TimeBeginAt
	}
	return 0
}

func (x *UpdateGMCertRequest) GetTimeEndAt() int64 {
	if x != nil {
		return x.TimeEndAt
	}
	return 0
}

func (x *UpdateGMCertRequest) GetDnsNames() []string {
	if x != nil {
		return x.DnsNames
	}
	return nil
}

func (x *UpdateGMCertRequest) GetCommonNames() []string {
	if x != nil {
		return x.CommonNames
	}
	return nil
}

type DeleteGMCertRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GmCertId int64 `protobuf:"varint,1,opt,name=gmCertId,proto3" json:"gmCertId,omitempty"`
}

func (x *DeleteGMCertRequest) Reset() {
	*x = DeleteGMCertRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_gm_cert_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteGMCertRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteGMCertRequest) ProtoMessage() {}

func (x *DeleteGMCertRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_gm_cert_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteGMCertRequest.ProtoReflect.Descriptor instead.
func (*DeleteGMCertRequest) Descriptor() ([]byte, []int) {
	return file_service_gm_cert_proto_rawDescGZIP(), []int{8}
}

func (x *DeleteGMCertRequest) GetGmCertId() int64 {
	if x != nil {
		return x.GmCertId
	}
	return 0
}

var File_service_gm_cert_proto protoreflect.FileDescriptor

var file_service_gm_cert_proto_rawDesc = []byte{
	0x0a, 0x15, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x67, 0x6d, 0x5f, 0x63, 0x65, 0x72,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x19, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2f, 0x72, 0x70, 0x63, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb1, 0x03, 0x0a, 0x13, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x69, 0x73, 0x4f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x69, 0x73,
	0x4f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x69, 0x73, 0x43, 0x41, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04,
	0x69, 0x73, 0x43, 0x41, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x69, 0x67, 0x6e, 0x43, 0x65, 0x72, 0x74,
	0x44, 0x61, 0x74, 0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x73, 0x69, 0x67, 0x6e,
	0x43, 0x65, 0x72, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x69, 0x67, 0x6e,
	0x4b, 0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x73,
	0x69, 0x67, 0x6e, 0x4b, 0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x20, 0x0a, 0x0b, 0x65, 0x6e,
	0x63, 0x43, 0x65, 0x72, 0x74, 0x44, 0x61, 0x74, 0x61, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x0b, 0x65, 0x6e, 0x63, 0x43, 0x65, 0x72, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1e, 0x0a, 0x0a,
	0x65, 0x6e, 0x63, 0x4b, 0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x0a, 0x65, 0x6e, 0x63, 0x4b, 0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x20, 0x0a, 0x0b,
	0x74, 0x69, 0x6d, 0x65, 0x42, 0x65, 0x67, 0x69, 0x6e, 0x41, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0b, 0x74, 0x69, 0x6d, 0x65, 0x42, 0x65, 0x67, 0x69, 0x6e, 0x41, 0x74, 0x12, 0x1c,
	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x45, 0x6e, 0x64, 0x41, 0x74, 0x18, 0x79, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x45, 0x6e, 0x64, 0x41, 0x74, 0x12, 0x1a, 0x0a, 0x08,
	0x64, 0x6e, 0x73, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08,
	0x64, 0x6e, 0x73, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x32, 0x0a, 0x14, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x22, 0x3c,
	0x0a, 0x1e, 0x46, 0x69, 0x6e, 0x64, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x47, 0x4d, 0x43,
	0x65, 0x72, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x08, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x22, 0x41, 0x0a, 0x1f,
	0x46, 0x69, 0x6e, 0x64, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x47, 0x4d, 0x43, 0x65, 0x72,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x1e, 0x0a, 0x0a, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x4a, 0x53, 0x4f, 0x4e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0a, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x4a, 0x53, 0x4f, 0x4e, 0x22,
	0xc4, 0x01, 0x0a, 0x12, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x69, 0x73, 0x41, 0x76, 0x61, 0x69,
	0x6c, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x69, 0x73, 0x41,
	0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x73, 0x45, 0x78,
	0x70, 0x69, 0x72, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x69, 0x73, 0x45,
	0x78, 0x70, 0x69, 0x72, 0x65, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x65, 0x78, 0x70, 0x69, 0x72, 0x69,
	0x6e, 0x67, 0x44, 0x61, 0x79, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x65, 0x78,
	0x70, 0x69, 0x72, 0x69, 0x6e, 0x67, 0x44, 0x61, 0x79, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6b, 0x65,
	0x79, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6b, 0x65, 0x79,
	0x77, 0x6f, 0x72, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07,
	0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x64,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x22, 0xf0, 0x01, 0x0a, 0x12, 0x4c, 0x69, 0x73, 0x74, 0x47,
	0x4d, 0x43, 0x65, 0x72, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a,
	0x0b, 0x69, 0x73, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0b, 0x69, 0x73, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x69, 0x73, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x09, 0x69, 0x73, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x64, 0x12, 0x22, 0x0a,
	0x0c, 0x65, 0x78, 0x70, 0x69, 0x72, 0x69, 0x6e, 0x67, 0x44, 0x61, 0x79, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0c, 0x65, 0x78, 0x70, 0x69, 0x72, 0x69, 0x6e, 0x67, 0x44, 0x61, 0x79,
	0x73, 0x12, 0x18, 0x0a, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x75,
	0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65,
	0x72, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x18, 0x06,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f,
	0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x37, 0x0a, 0x13, 0x4c, 0x69, 0x73,
	0x74, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x20, 0x0a, 0x0b, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x73, 0x4a, 0x53, 0x4f, 0x4e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x73, 0x4a, 0x53,
	0x4f, 0x4e, 0x22, 0xa1, 0x03, 0x0a, 0x13, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x4d, 0x43,
	0x65, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x67, 0x6d,
	0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x67, 0x6d,
	0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x69, 0x73, 0x4f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x69, 0x73, 0x4f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x22, 0x0a, 0x0c, 0x73, 0x69, 0x67, 0x6e, 0x43, 0x65, 0x72, 0x74, 0x44, 0x61, 0x74, 0x61,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x73, 0x69, 0x67, 0x6e, 0x43, 0x65, 0x72, 0x74,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x4b, 0x65, 0x79, 0x44,
	0x61, 0x74, 0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x4b,
	0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x20, 0x0a, 0x0b, 0x65, 0x6e, 0x63, 0x43, 0x65, 0x72,
	0x74, 0x44, 0x61, 0x74, 0x61, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x65, 0x6e, 0x63,
	0x43, 0x65, 0x72, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x6e, 0x63, 0x4b,
	0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x65, 0x6e,
	0x63, 0x4b, 0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x20, 0x0a, 0x0b, 0x74, 0x69, 0x6d, 0x65,
	0x42, 0x65, 0x67, 0x69, 0x6e, 0x41, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x74,
	0x69, 0x6d, 0x65, 0x42, 0x65, 0x67, 0x69, 0x6e, 0x41, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x45, 0x6e, 0x64, 0x41, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74,
	0x69, 0x6d, 0x65, 0x45, 0x6e, 0x64, 0x41, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x6e, 0x73, 0x4e,
	0x61, 0x6d, 0x65, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x64, 0x6e, 0x73, 0x4e,
	0x61, 0x6d, 0x65, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x31, 0x0a, 0x13, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x08, 0x67, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x49, 0x64, 0x32, 0xa6, 0x03, 0x0a, 0x0d, 0x47, 0x4d,
	0x43, 0x65, 0x72, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41, 0x0a, 0x0c, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x12, 0x17, 0x2e, 0x70, 0x62,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x62,
	0x0a, 0x17, 0x66, 0x69, 0x6e, 0x64, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x47, 0x4d, 0x43,
	0x65, 0x72, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x22, 0x2e, 0x70, 0x62, 0x2e, 0x46,
	0x69, 0x6e, 0x64, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e,
	0x70, 0x62, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x47, 0x4d,
	0x43, 0x65, 0x72, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x37, 0x0a, 0x0c, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x4d, 0x43, 0x65,
	0x72, 0x74, 0x12, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x4d,
	0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x62,
	0x2e, 0x52, 0x50, 0x43, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x3c, 0x0a, 0x0c, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x73, 0x12, 0x16, 0x2e, 0x70, 0x62,
	0x2e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x50, 0x43, 0x43, 0x6f, 0x75, 0x6e,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x37, 0x0a, 0x0c, 0x64, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x12, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x50, 0x43, 0x53, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x12, 0x3e, 0x0a, 0x0b, 0x6c, 0x69, 0x73, 0x74, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74,
	0x73, 0x12, 0x16, 0x2e, 0x70, 0x62, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x47, 0x4d, 0x43, 0x65, 0x72,
	0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x47, 0x4d, 0x43, 0x65, 0x72, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_service_gm_cert_proto_rawDescOnce sync.Once
	file_service_gm_cert_proto_rawDescData = file_service_gm_cert_proto_rawDesc
)

func file_service_gm_cert_proto_rawDescGZIP() []byte {
	file_service_gm_cert_proto_rawDescOnce.Do(func() {
		file_service_gm_cert_proto_rawDescData = protoimpl.X.CompressGZIP(file_service_gm_cert_proto_rawDescData)
	})
	return file_service_gm_cert_proto_rawDescData
}

var file_service_gm_cert_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_service_gm_cert_proto_goTypes = []interface{}{
	(*CreateGMCertRequest)(nil),             // 0: pb.CreateGMCertRequest
	(*CreateGMCertResponse)(nil),            // 1: pb.CreateGMCertResponse
	(*FindEnabledGMCertConfigRequest)(nil),  // 2: pb.FindEnabledGMCertConfigRequest
	(*FindEnabledGMCertConfigResponse)(nil), // 3: pb.FindEnabledGMCertConfigResponse
	(*CountGMCertRequest)(nil),              // 4: pb.CountGMCertRequest
	(*ListGMCertsRequest)(nil),              // 5: pb.ListGMCertsRequest
	(*ListGMCertsResponse)(nil),             // 6: pb.ListGMCertsResponse
	(*UpdateGMCertRequest)(nil),             // 7: pb.UpdateGMCertRequest
	(*DeleteGMCertRequest)(nil),             // 8: pb.DeleteGMCertRequest
	(*RPCSuccess)(nil),                      // 9: pb.RPCSuccess
	(*RPCCountResponse)(nil),                // 10: pb.RPCCountResponse
}
var file_service_gm_cert_proto_depIdxs = []int32{
	0,  // 0: pb.GMCertService.createGMCert:input_type -> pb.CreateGMCertRequest
	2,  // 1: pb.GMCertService.findEnabledGMCertConfig:input_type -> pb.FindEnabledGMCertConfigRequest
	7,  // 2: pb.GMCertService.updateGMCert:input_type -> pb.UpdateGMCertRequest
	4,  // 3: pb.GMCertService.countGMCerts:input_type -> pb.CountGMCertRequest
	8,  // 4: pb.GMCertService.deleteGMCert:input_type -> pb.DeleteGMCertRequest
	5,  // 5: pb.GMCertService.listGMCerts:input_type -> pb.ListGMCertsRequest
	1,  // 6: pb.GMCertService.createGMCert:output_type -> pb.CreateGMCertResponse
	3,  // 7: pb.GMCertService.findEnabledGMCertConfig:output_type -> pb.FindEnabledGMCertConfigResponse
	9,  // 8: pb.GMCertService.updateGMCert:output_type -> pb.RPCSuccess
	10, // 9: pb.GMCertService.countGMCerts:output_type -> pb.RPCCountResponse
	9,  // 10: pb.GMCertService.deleteGMCert:output_type -> pb.RPCSuccess
	6,  // 11: pb.GMCertService.listGMCerts:output_type -> pb.ListGMCertsResponse
	6,  // [6:12] is the sub-list for method output_type
	0,  // [0:6] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_service_gm_cert_proto_init() }
func file_service_gm_cert_proto_init() {
	if File_service_gm_cert_proto != nil {
		return
	}
	file_models_rpc_messages_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_service_gm_cert_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateGMCertRequest); i {
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
		file_service_gm_cert_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateGMCertResponse); i {
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
		file_service_gm_cert_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindEnabledGMCertConfigRequest); i {
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
		file_service_gm_cert_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindEnabledGMCertConfigResponse); i {
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
		file_service_gm_cert_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CountGMCertRequest); i {
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
		file_service_gm_cert_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListGMCertsRequest); i {
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
		file_service_gm_cert_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListGMCertsResponse); i {
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
		file_service_gm_cert_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateGMCertRequest); i {
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
		file_service_gm_cert_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteGMCertRequest); i {
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
			RawDescriptor: file_service_gm_cert_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_service_gm_cert_proto_goTypes,
		DependencyIndexes: file_service_gm_cert_proto_depIdxs,
		MessageInfos:      file_service_gm_cert_proto_msgTypes,
	}.Build()
	File_service_gm_cert_proto = out.File
	file_service_gm_cert_proto_rawDesc = nil
	file_service_gm_cert_proto_goTypes = nil
	file_service_gm_cert_proto_depIdxs = nil
}