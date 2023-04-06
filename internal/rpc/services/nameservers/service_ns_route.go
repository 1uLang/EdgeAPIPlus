// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus

package nameservers

import (
	"context"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/iwind/TeaGo/types"
)

// NSRouteService 线路相关服务
type NSRouteService struct {
	services.BaseService
}

// CreateNSRoute 创建自定义线路
func (this *NSRouteService) CreateNSRoute(ctx context.Context, req *pb.CreateNSRouteRequest) (*pb.CreateNSRouteResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId

		// 暂时不允许在集群和域名下创建线路
		req.NsClusterId = 0
		req.NsDomainId = 0
	}

	// TODO 检查线路数限制

	routeId, err := nameservers.SharedNSRouteDAO.CreateRoute(tx, req.NsClusterId, req.NsDomainId, req.UserId, req.Name, req.RangesJSON)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSRouteResponse{NsRouteId: routeId}, nil
}

// UpdateNSRoute 修改自定义线路
func (this *NSRouteService) UpdateNSRoute(ctx context.Context, req *pb.UpdateNSRouteRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSRouteDAO.CheckUserRoute(tx, userId, req.NsRouteId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSRouteDAO.UpdateRoute(tx, req.NsRouteId, req.Name, req.RangesJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSRoute 删除自定义线路
func (this *NSRouteService) DeleteNSRoute(ctx context.Context, req *pb.DeleteNSRouteRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSRouteDAO.CheckUserRoute(tx, userId, req.NsRouteId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSRouteDAO.DisableNSRoute(tx, req.NsRouteId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSRoute 获取单个自定义路线信息
func (this *NSRouteService) FindNSRoute(ctx context.Context, req *pb.FindNSRouteRequest) (*pb.FindNSRouteResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSRouteDAO.CheckUserRoute(tx, userId, req.NsRouteId)
		if err != nil {
			return nil, err
		}
	}

	route, err := nameservers.SharedNSRouteDAO.FindEnabledNSRoute(tx, req.NsRouteId)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return &pb.FindNSRouteResponse{NsRoute: nil}, nil
	}

	// 集群
	var pbCluster *pb.NSCluster
	if route.ClusterId > 0 {
		cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(route.ClusterId))
		if err != nil {
			return nil, err
		}
		if cluster != nil {
			pbCluster = &pb.NSCluster{
				Id:   int64(cluster.Id),
				IsOn: cluster.IsOn,
				Name: cluster.Name,
			}
		}
	}

	// 域名
	var pbDomain *pb.NSDomain
	if route.DomainId > 0 {
		domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, int64(route.DomainId))
		if err != nil {
			return nil, err
		}
		if domain != nil {
			pbDomain = &pb.NSDomain{
				Id:   int64(domain.Id),
				Name: domain.Name,
				IsOn: domain.IsOn,
			}
		}
	}

	return &pb.FindNSRouteResponse{NsRoute: &pb.NSRoute{
		Id:         int64(route.Id),
		IsOn:       route.IsOn,
		Name:       route.Name,
		RangesJSON: route.Ranges,
		NsCluster:  pbCluster,
		NsDomain:   pbDomain,
	}}, nil
}

// CountAllNSRoutes 查询自定义线路数量
func (this *NSRouteService) CountAllNSRoutes(ctx context.Context, req *pb.CountAllNSRoutesRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	countRoutes, err := nameservers.SharedNSRouteDAO.CountAllEnabledRoutes(tx, req.NsClusterId, req.NsClusterId, req.UserId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(countRoutes)
}

// FindAllNSRoutes 读取所有自定义线路
func (this *NSRouteService) FindAllNSRoutes(ctx context.Context, req *pb.FindAllNSRoutesRequest) (*pb.FindAllNSRoutesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId
	}

	routes, err := nameservers.SharedNSRouteDAO.FindAllEnabledRoutes(tx, req.NsClusterId, req.NsDomainId, req.UserId)
	if err != nil {
		return nil, err
	}
	var pbRoutes = []*pb.NSRoute{}
	for _, route := range routes {
		// 集群
		var pbCluster *pb.NSCluster
		if route.ClusterId > 0 {
			cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(route.ClusterId))
			if err != nil {
				return nil, err
			}
			if cluster != nil {
				pbCluster = &pb.NSCluster{
					Id:   int64(cluster.Id),
					IsOn: cluster.IsOn,
					Name: cluster.Name,
				}
			}
		}

		// 域名
		var pbDomain *pb.NSDomain
		if route.DomainId > 0 {
			domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, int64(route.DomainId))
			if err != nil {
				return nil, err
			}
			if domain != nil {
				pbDomain = &pb.NSDomain{
					Id:   int64(domain.Id),
					Name: domain.Name,
					IsOn: domain.IsOn,
				}
			}
		}

		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Id:         int64(route.Id),
			IsOn:       route.IsOn,
			Code:       "id:" + types.String(route.Id),
			Name:       route.Name,
			RangesJSON: route.Ranges,
			NsCluster:  pbCluster,
			NsDomain:   pbDomain,
		})
	}
	return &pb.FindAllNSRoutesResponse{NsRoutes: pbRoutes}, nil
}

// UpdateNSRouteOrders 设置自定义线路排序
func (this *NSRouteService) UpdateNSRouteOrders(ctx context.Context, req *pb.UpdateNSRouteOrdersRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		for _, routeId := range req.NsRouteIds {
			err = nameservers.SharedNSRouteDAO.CheckUserRoute(tx, userId, routeId)
			if err != nil {
				return nil, err
			}
		}
	}

	err = nameservers.SharedNSRouteDAO.UpdateRouteOrders(tx, req.NsRouteIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ListNSRoutesAfterVersion 根据版本列出一组自定义线路
func (this *NSRouteService) ListNSRoutesAfterVersion(ctx context.Context, req *pb.ListNSRoutesAfterVersionRequest) (*pb.ListNSRoutesAfterVersionResponse, error) {
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	// 集群ID
	var tx = this.NullTx()
	routes, err := nameservers.SharedNSRouteDAO.ListRoutesAfterVersion(tx, req.Version, 2000)
	if err != nil {
		return nil, err
	}

	var pbRoutes []*pb.NSRoute
	for _, route := range routes {
		// 集群
		var pbCluster *pb.NSCluster
		if route.ClusterId > 0 {
			pbCluster = &pb.NSCluster{Id: int64(route.ClusterId)}
		}

		// 域名
		var pbDomain *pb.NSDomain
		if route.DomainId > 0 {
			pbDomain = &pb.NSDomain{Id: int64(route.DomainId)}
		}

		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Id:         int64(route.Id),
			IsOn:       route.IsOn,
			Name:       "",
			RangesJSON: route.Ranges,
			IsDeleted:  route.State == nameservers.NSRouteStateDisabled,
			Order:      int64(route.Order),
			Version:    int64(route.Version),
			NsCluster:  pbCluster,
			NsDomain:   pbDomain,
		})
	}
	return &pb.ListNSRoutesAfterVersionResponse{NsRoutes: pbRoutes}, nil
}

// FindAllDefaultWorldRegionRoutes 查找默认的世界区域线路
func (this *NSRouteService) FindAllDefaultWorldRegionRoutes(ctx context.Context, req *pb.FindAllDefaultWorldRegionRoutesRequest) (*pb.FindAllDefaultWorldRegionRoutesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var pbRoutes = []*pb.NSRoute{}
	for _, route := range dnsconfigs.AllDefaultWorldRegionRoutes {
		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Code: route.Code,
			Name: route.Name,
		})
	}
	return &pb.FindAllDefaultWorldRegionRoutesResponse{
		NsRoutes: pbRoutes,
	}, nil
}

// FindAllDefaultChinaProvinceRoutes 查找默认的中国省份线路
func (this *NSRouteService) FindAllDefaultChinaProvinceRoutes(ctx context.Context, req *pb.FindAllDefaultChinaProvinceRoutesRequest) (*pb.FindAllDefaultChinaProvinceRoutesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var pbRoutes = []*pb.NSRoute{}
	for _, route := range dnsconfigs.AllDefaultChinaProvinceRoutes {
		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Code: route.Code,
			Name: route.Name,
		})
	}
	return &pb.FindAllDefaultChinaProvinceRoutesResponse{
		NsRoutes: pbRoutes,
	}, nil
}

// FindAllDefaultISPRoutes 查找默认的ISP线路
func (this *NSRouteService) FindAllDefaultISPRoutes(ctx context.Context, req *pb.FindAllDefaultISPRoutesRequest) (*pb.FindAllDefaultISPRoutesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var pbRoutes = []*pb.NSRoute{}
	for _, route := range dnsconfigs.AllDefaultISPRoutes {
		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Code: route.Code,
			Name: route.Name,
		})
	}
	return &pb.FindAllDefaultISPRoutesResponse{
		NsRoutes: pbRoutes,
	}, nil
}
