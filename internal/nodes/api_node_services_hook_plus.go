// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nodes

import (
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/reporters"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/tickets"
	"google.golang.org/grpc"
)

func APINodeServicesRegister(node *APINode, server *grpc.Server) {
	{
		var instance = node.serviceInstance(&nameservers.NSClusterService{}).(*nameservers.NSClusterService)
		pb.RegisterNSClusterServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSNodeService{}).(*nameservers.NSNodeService)
		pb.RegisterNSNodeServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSDomainService{}).(*nameservers.NSDomainService)
		pb.RegisterNSDomainServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSDomainGroupService{}).(*nameservers.NSDomainGroupService)
		pb.RegisterNSDomainGroupServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSRecordService{}).(*nameservers.NSRecordService)
		pb.RegisterNSRecordServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSRouteService{}).(*nameservers.NSRouteService)
		pb.RegisterNSRouteServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSKeyService{}).(*nameservers.NSKeyService)
		pb.RegisterNSKeyServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSAccessLogService{}).(*nameservers.NSAccessLogService)
		pb.RegisterNSAccessLogServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSRecordHourlyStatService{}).(*nameservers.NSRecordHourlyStatService)
		pb.RegisterNSRecordHourlyStatServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSQuestionOptionService{}).(*nameservers.NSQuestionOptionService)
		pb.RegisterNSQuestionOptionServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSPlanService{}).(*nameservers.NSPlanService)
		pb.RegisterNSPlanServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSUserPlanService{}).(*nameservers.NSUserPlanService)
		pb.RegisterNSUserPlanServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&nameservers.NSService{}).(*nameservers.NSService)
		pb.RegisterNSServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&services.AuthorityKeyService{}).(*services.AuthorityKeyService)
		pb.RegisterAuthorityKeyServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&reporters.ReportNodeService{}).(*reporters.ReportNodeService)
		pb.RegisterReportNodeServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&reporters.ReportNodeGroupService{}).(*reporters.ReportNodeGroupService)
		pb.RegisterReportNodeGroupServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&reporters.ReportResultService{}).(*reporters.ReportResultService)
		pb.RegisterReportResultServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&accounts.UserAccountService{}).(*accounts.UserAccountService)
		pb.RegisterUserAccountServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&accounts.UserAccountLogService{}).(*accounts.UserAccountLogService)
		pb.RegisterUserAccountLogServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&accounts.UserAccountDailyStatService{}).(*accounts.UserAccountDailyStatService)
		pb.RegisterUserAccountDailyStatServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&accounts.UserOrderService{}).(*accounts.UserOrderService)
		pb.RegisterUserOrderServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&accounts.OrderMethodService{}).(*accounts.OrderMethodService)
		pb.RegisterOrderMethodServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&services.ScriptService{}).(*services.ScriptService)
		pb.RegisterScriptServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&tickets.UserTicketService{}).(*tickets.UserTicketService)
		pb.RegisterUserTicketServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&tickets.UserTicketCategoryService{}).(*tickets.UserTicketCategoryService)
		pb.RegisterUserTicketCategoryServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&tickets.UserTicketLogService{}).(*tickets.UserTicketLogService)
		pb.RegisterUserTicketLogServiceServer(server, instance)
		node.rest(instance)
	}
	{
		var instance = node.serviceInstance(&services.HTTPAccessLogPolicyService{}).(*services.HTTPAccessLogPolicyService)
		pb.RegisterHTTPAccessLogPolicyServiceServer(server, instance)
		node.rest(instance)
	}
}
