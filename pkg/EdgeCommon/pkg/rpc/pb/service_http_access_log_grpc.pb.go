// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: service_http_access_log.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// HTTPAccessLogServiceClient is the client API for HTTPAccessLogService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HTTPAccessLogServiceClient interface {
	// 创建访问日志
	CreateHTTPAccessLogs(ctx context.Context, in *CreateHTTPAccessLogsRequest, opts ...grpc.CallOption) (*CreateHTTPAccessLogsResponse, error)
	// 列出单页访问日志
	ListHTTPAccessLogs(ctx context.Context, in *ListHTTPAccessLogsRequest, opts ...grpc.CallOption) (*ListHTTPAccessLogsResponse, error)
	// 查找单个日志
	FindHTTPAccessLog(ctx context.Context, in *FindHTTPAccessLogRequest, opts ...grpc.CallOption) (*FindHTTPAccessLogResponse, error)
	// 查找日志分区
	FindHTTPAccessLogPartitions(ctx context.Context, in *FindHTTPAccessLogPartitionsRequest, opts ...grpc.CallOption) (*FindHTTPAccessLogPartitionsResponse, error)
	// dengbao-waf 定制rpc接口 列出单页访问日志 新增 时间条件
	SearchHTTPAccessLogs(ctx context.Context, in *SearchHTTPAccessLogsRequest, opts ...grpc.CallOption) (*SearchHTTPAccessLogsResponse, error)
	// 统计用户相关域名的攻击排行（ip/区域）
	StatisticsHTTPAccessTop(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessTopResponse, error)
	// 统计指定日期用户相关域名对应的攻击次数
	StatisticsHTTPAccess(ctx context.Context, in *StatisticsHTTPAccessRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessResponse, error)
	// 统计指定日期用户相关域名对应的攻击类型次数
	StatisticsHTTPAccessType(ctx context.Context, in *StatisticsHTTPAccessTypeRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessTypeResponse, error)
	// 统计统计指定用户日期下 各访问的 访问条数 访问总次数  防护总次数 访问IP总数 拦截IP总数
	StatisticsHTTPAccessLogs(ctx context.Context, in *StatisticsHTTPAccessTypeRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessLogResponse, error)
	// 统计最受攻击的域名排行
	StatisticsAttackURLTop(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsHTTPAttackURLTopResponse, error)
	// 统计客户端访问IP排行
	StatisticsAccessIPTop(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessIPTopResponse, error)
	// 统计客户端访问IP排行
	StatusCodeStatistics(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsStatusCodeTopResponse, error)
}

type hTTPAccessLogServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHTTPAccessLogServiceClient(cc grpc.ClientConnInterface) HTTPAccessLogServiceClient {
	return &hTTPAccessLogServiceClient{cc}
}

func (c *hTTPAccessLogServiceClient) CreateHTTPAccessLogs(ctx context.Context, in *CreateHTTPAccessLogsRequest, opts ...grpc.CallOption) (*CreateHTTPAccessLogsResponse, error) {
	out := new(CreateHTTPAccessLogsResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/createHTTPAccessLogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) ListHTTPAccessLogs(ctx context.Context, in *ListHTTPAccessLogsRequest, opts ...grpc.CallOption) (*ListHTTPAccessLogsResponse, error) {
	out := new(ListHTTPAccessLogsResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/listHTTPAccessLogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) FindHTTPAccessLog(ctx context.Context, in *FindHTTPAccessLogRequest, opts ...grpc.CallOption) (*FindHTTPAccessLogResponse, error) {
	out := new(FindHTTPAccessLogResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/findHTTPAccessLog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) FindHTTPAccessLogPartitions(ctx context.Context, in *FindHTTPAccessLogPartitionsRequest, opts ...grpc.CallOption) (*FindHTTPAccessLogPartitionsResponse, error) {
	out := new(FindHTTPAccessLogPartitionsResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/findHTTPAccessLogPartitions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) SearchHTTPAccessLogs(ctx context.Context, in *SearchHTTPAccessLogsRequest, opts ...grpc.CallOption) (*SearchHTTPAccessLogsResponse, error) {
	out := new(SearchHTTPAccessLogsResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/searchHTTPAccessLogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) StatisticsHTTPAccessTop(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessTopResponse, error) {
	out := new(StatisticsHTTPAccessTopResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/statisticsHTTPAccessTop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) StatisticsHTTPAccess(ctx context.Context, in *StatisticsHTTPAccessRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessResponse, error) {
	out := new(StatisticsHTTPAccessResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/statisticsHTTPAccess", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) StatisticsHTTPAccessType(ctx context.Context, in *StatisticsHTTPAccessTypeRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessTypeResponse, error) {
	out := new(StatisticsHTTPAccessTypeResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/statisticsHTTPAccessType", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) StatisticsHTTPAccessLogs(ctx context.Context, in *StatisticsHTTPAccessTypeRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessLogResponse, error) {
	out := new(StatisticsHTTPAccessLogResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/StatisticsHTTPAccessLogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) StatisticsAttackURLTop(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsHTTPAttackURLTopResponse, error) {
	out := new(StatisticsHTTPAttackURLTopResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/StatisticsAttackURLTop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) StatisticsAccessIPTop(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsHTTPAccessIPTopResponse, error) {
	out := new(StatisticsHTTPAccessIPTopResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/StatisticsAccessIPTop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hTTPAccessLogServiceClient) StatusCodeStatistics(ctx context.Context, in *StatisticsHTTPAccessTopRequest, opts ...grpc.CallOption) (*StatisticsStatusCodeTopResponse, error) {
	out := new(StatisticsStatusCodeTopResponse)
	err := c.cc.Invoke(ctx, "/pb.HTTPAccessLogService/StatusCodeStatistics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HTTPAccessLogServiceServer is the server API for HTTPAccessLogService service.
// All implementations must embed UnimplementedHTTPAccessLogServiceServer
// for forward compatibility
type HTTPAccessLogServiceServer interface {
	// 创建访问日志
	CreateHTTPAccessLogs(context.Context, *CreateHTTPAccessLogsRequest) (*CreateHTTPAccessLogsResponse, error)
	// 列出单页访问日志
	ListHTTPAccessLogs(context.Context, *ListHTTPAccessLogsRequest) (*ListHTTPAccessLogsResponse, error)
	// 查找单个日志
	FindHTTPAccessLog(context.Context, *FindHTTPAccessLogRequest) (*FindHTTPAccessLogResponse, error)
	// 查找日志分区
	FindHTTPAccessLogPartitions(context.Context, *FindHTTPAccessLogPartitionsRequest) (*FindHTTPAccessLogPartitionsResponse, error)
	// dengbao-waf 定制rpc接口 列出单页访问日志 新增 时间条件
	SearchHTTPAccessLogs(context.Context, *SearchHTTPAccessLogsRequest) (*SearchHTTPAccessLogsResponse, error)
	// 统计用户相关域名的攻击排行（ip/区域）
	StatisticsHTTPAccessTop(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsHTTPAccessTopResponse, error)
	// 统计指定日期用户相关域名对应的攻击次数
	StatisticsHTTPAccess(context.Context, *StatisticsHTTPAccessRequest) (*StatisticsHTTPAccessResponse, error)
	// 统计指定日期用户相关域名对应的攻击类型次数
	StatisticsHTTPAccessType(context.Context, *StatisticsHTTPAccessTypeRequest) (*StatisticsHTTPAccessTypeResponse, error)
	// 统计统计指定用户日期下 各访问的 访问条数 访问总次数  防护总次数 访问IP总数 拦截IP总数
	StatisticsHTTPAccessLogs(context.Context, *StatisticsHTTPAccessTypeRequest) (*StatisticsHTTPAccessLogResponse, error)
	// 统计最受攻击的域名排行
	StatisticsAttackURLTop(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsHTTPAttackURLTopResponse, error)
	// 统计客户端访问IP排行
	StatisticsAccessIPTop(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsHTTPAccessIPTopResponse, error)
	// 统计客户端访问IP排行
	StatusCodeStatistics(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsStatusCodeTopResponse, error)
	mustEmbedUnimplementedHTTPAccessLogServiceServer()
}

// UnimplementedHTTPAccessLogServiceServer must be embedded to have forward compatible implementations.
type UnimplementedHTTPAccessLogServiceServer struct {
}

func (UnimplementedHTTPAccessLogServiceServer) CreateHTTPAccessLogs(context.Context, *CreateHTTPAccessLogsRequest) (*CreateHTTPAccessLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateHTTPAccessLogs not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) ListHTTPAccessLogs(context.Context, *ListHTTPAccessLogsRequest) (*ListHTTPAccessLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListHTTPAccessLogs not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) FindHTTPAccessLog(context.Context, *FindHTTPAccessLogRequest) (*FindHTTPAccessLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindHTTPAccessLog not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) FindHTTPAccessLogPartitions(context.Context, *FindHTTPAccessLogPartitionsRequest) (*FindHTTPAccessLogPartitionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindHTTPAccessLogPartitions not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) SearchHTTPAccessLogs(context.Context, *SearchHTTPAccessLogsRequest) (*SearchHTTPAccessLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchHTTPAccessLogs not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) StatisticsHTTPAccessTop(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsHTTPAccessTopResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StatisticsHTTPAccessTop not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) StatisticsHTTPAccess(context.Context, *StatisticsHTTPAccessRequest) (*StatisticsHTTPAccessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StatisticsHTTPAccess not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) StatisticsHTTPAccessType(context.Context, *StatisticsHTTPAccessTypeRequest) (*StatisticsHTTPAccessTypeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StatisticsHTTPAccessType not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) StatisticsHTTPAccessLogs(context.Context, *StatisticsHTTPAccessTypeRequest) (*StatisticsHTTPAccessLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StatisticsHTTPAccessLogs not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) StatisticsAttackURLTop(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsHTTPAttackURLTopResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StatisticsAttackURLTop not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) StatisticsAccessIPTop(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsHTTPAccessIPTopResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StatisticsAccessIPTop not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) StatusCodeStatistics(context.Context, *StatisticsHTTPAccessTopRequest) (*StatisticsStatusCodeTopResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StatusCodeStatistics not implemented")
}
func (UnimplementedHTTPAccessLogServiceServer) mustEmbedUnimplementedHTTPAccessLogServiceServer() {}

// UnsafeHTTPAccessLogServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HTTPAccessLogServiceServer will
// result in compilation errors.
type UnsafeHTTPAccessLogServiceServer interface {
	mustEmbedUnimplementedHTTPAccessLogServiceServer()
}

func RegisterHTTPAccessLogServiceServer(s grpc.ServiceRegistrar, srv HTTPAccessLogServiceServer) {
	s.RegisterService(&HTTPAccessLogService_ServiceDesc, srv)
}

func _HTTPAccessLogService_CreateHTTPAccessLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateHTTPAccessLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).CreateHTTPAccessLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/createHTTPAccessLogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).CreateHTTPAccessLogs(ctx, req.(*CreateHTTPAccessLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_ListHTTPAccessLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListHTTPAccessLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).ListHTTPAccessLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/listHTTPAccessLogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).ListHTTPAccessLogs(ctx, req.(*ListHTTPAccessLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_FindHTTPAccessLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindHTTPAccessLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).FindHTTPAccessLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/findHTTPAccessLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).FindHTTPAccessLog(ctx, req.(*FindHTTPAccessLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_FindHTTPAccessLogPartitions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindHTTPAccessLogPartitionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).FindHTTPAccessLogPartitions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/findHTTPAccessLogPartitions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).FindHTTPAccessLogPartitions(ctx, req.(*FindHTTPAccessLogPartitionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_SearchHTTPAccessLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchHTTPAccessLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).SearchHTTPAccessLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/searchHTTPAccessLogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).SearchHTTPAccessLogs(ctx, req.(*SearchHTTPAccessLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_StatisticsHTTPAccessTop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsHTTPAccessTopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccessTop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/statisticsHTTPAccessTop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccessTop(ctx, req.(*StatisticsHTTPAccessTopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_StatisticsHTTPAccess_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsHTTPAccessRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccess(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/statisticsHTTPAccess",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccess(ctx, req.(*StatisticsHTTPAccessRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_StatisticsHTTPAccessType_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsHTTPAccessTypeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccessType(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/statisticsHTTPAccessType",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccessType(ctx, req.(*StatisticsHTTPAccessTypeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_StatisticsHTTPAccessLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsHTTPAccessTypeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccessLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/StatisticsHTTPAccessLogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).StatisticsHTTPAccessLogs(ctx, req.(*StatisticsHTTPAccessTypeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_StatisticsAttackURLTop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsHTTPAccessTopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).StatisticsAttackURLTop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/StatisticsAttackURLTop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).StatisticsAttackURLTop(ctx, req.(*StatisticsHTTPAccessTopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_StatisticsAccessIPTop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsHTTPAccessTopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).StatisticsAccessIPTop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/StatisticsAccessIPTop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).StatisticsAccessIPTop(ctx, req.(*StatisticsHTTPAccessTopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HTTPAccessLogService_StatusCodeStatistics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsHTTPAccessTopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPAccessLogServiceServer).StatusCodeStatistics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.HTTPAccessLogService/StatusCodeStatistics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPAccessLogServiceServer).StatusCodeStatistics(ctx, req.(*StatisticsHTTPAccessTopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// HTTPAccessLogService_ServiceDesc is the grpc.ServiceDesc for HTTPAccessLogService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HTTPAccessLogService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.HTTPAccessLogService",
	HandlerType: (*HTTPAccessLogServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "createHTTPAccessLogs",
			Handler:    _HTTPAccessLogService_CreateHTTPAccessLogs_Handler,
		},
		{
			MethodName: "listHTTPAccessLogs",
			Handler:    _HTTPAccessLogService_ListHTTPAccessLogs_Handler,
		},
		{
			MethodName: "findHTTPAccessLog",
			Handler:    _HTTPAccessLogService_FindHTTPAccessLog_Handler,
		},
		{
			MethodName: "findHTTPAccessLogPartitions",
			Handler:    _HTTPAccessLogService_FindHTTPAccessLogPartitions_Handler,
		},
		{
			MethodName: "searchHTTPAccessLogs",
			Handler:    _HTTPAccessLogService_SearchHTTPAccessLogs_Handler,
		},
		{
			MethodName: "statisticsHTTPAccessTop",
			Handler:    _HTTPAccessLogService_StatisticsHTTPAccessTop_Handler,
		},
		{
			MethodName: "statisticsHTTPAccess",
			Handler:    _HTTPAccessLogService_StatisticsHTTPAccess_Handler,
		},
		{
			MethodName: "statisticsHTTPAccessType",
			Handler:    _HTTPAccessLogService_StatisticsHTTPAccessType_Handler,
		},
		{
			MethodName: "StatisticsHTTPAccessLogs",
			Handler:    _HTTPAccessLogService_StatisticsHTTPAccessLogs_Handler,
		},
		{
			MethodName: "StatisticsAttackURLTop",
			Handler:    _HTTPAccessLogService_StatisticsAttackURLTop_Handler,
		},
		{
			MethodName: "StatisticsAccessIPTop",
			Handler:    _HTTPAccessLogService_StatisticsAccessIPTop_Handler,
		},
		{
			MethodName: "StatusCodeStatistics",
			Handler:    _HTTPAccessLogService_StatusCodeStatistics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service_http_access_log.proto",
}