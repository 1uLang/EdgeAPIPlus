// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package nodes

import (
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/reporters"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc"
)

func APINodeServicesRegister(node *APINode, server *grpc.Server) {
	{
		instance := node.serviceInstance(&services.AuthorityKeyService{}).(*services.AuthorityKeyService)
		pb.RegisterAuthorityKeyServiceServer(server, instance)
		node.rest(instance)
	}
	{
		instance := node.serviceInstance(&reporters.ReportNodeService{}).(*reporters.ReportNodeService)
		pb.RegisterReportNodeServiceServer(server, instance)
		node.rest(instance)
	}
	{
		instance := node.serviceInstance(&reporters.ReportNodeGroupService{}).(*reporters.ReportNodeGroupService)
		pb.RegisterReportNodeGroupServiceServer(server, instance)
		node.rest(instance)
	}
	{
		instance := node.serviceInstance(&reporters.ReportResultService{}).(*reporters.ReportResultService)
		pb.RegisterReportResultServiceServer(server, instance)
		node.rest(instance)
	}
	{
		instance := node.serviceInstance(&accounts.UserAccountService{}).(*accounts.UserAccountService)
		pb.RegisterUserAccountServiceServer(server, instance)
		node.rest(instance)
	}
	{
		instance := node.serviceInstance(&accounts.UserAccountLogService{}).(*accounts.UserAccountLogService)
		pb.RegisterUserAccountLogServiceServer(server, instance)
		node.rest(instance)
	}
	{
		instance := node.serviceInstance(&accounts.UserAccountDailyStatService{}).(*accounts.UserAccountDailyStatService)
		pb.RegisterUserAccountDailyStatServiceServer(server, instance)
		node.rest(instance)
	}
}
