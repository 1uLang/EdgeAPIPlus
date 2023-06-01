// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package antiddos

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// ADPackageInstanceService 高防实例服务
type ADPackageInstanceService struct {
	services.BaseService
}

// CreateADPackageInstance 创建实例
func (this *ADPackageInstanceService) CreateADPackageInstance(ctx context.Context, req *pb.CreateADPackageInstanceRequest) (*pb.CreateADPackageInstanceResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// validate
	if req.AdPackageId <= 0 {
		return nil, errors.New("invalid 'adPackageId'")
	}
	if req.NodeClusterId <= 0 {
		return nil, errors.New("invalid 'nodeClusterId'")
	}

	instanceId, err := models.SharedADPackageInstanceDAO.CreateInstance(tx, req.AdPackageId, req.NodeClusterId, req.NodeIds, req.IpAddresses)
	if err != nil {
		return nil, err
	}
	return &pb.CreateADPackageInstanceResponse{AdPackageInstanceId: instanceId}, nil
}

// UpdateADPackageInstance 修改实例
func (this *ADPackageInstanceService) UpdateADPackageInstance(ctx context.Context, req *pb.UpdateADPackageInstanceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// validate
	if req.AdPackageInstanceId <= 0 {
		return nil, errors.New("invalid 'adPackageInstanceId'")
	}
	if req.NodeClusterId <= 0 {
		return nil, errors.New("invalid 'nodeClusterId'")
	}

	err = models.SharedADPackageInstanceDAO.UpdateInstance(tx, req.AdPackageInstanceId, req.NodeClusterId, req.NodeIds, req.IpAddresses, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindADPackageInstance 查找单个实例
func (this *ADPackageInstanceService) FindADPackageInstance(ctx context.Context, req *pb.FindADPackageInstanceRequest) (*pb.FindADPackageInstanceResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	instance, err := models.SharedADPackageInstanceDAO.FindEnabledADPackageInstance(tx, req.AdPackageInstanceId)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return &pb.FindADPackageInstanceResponse{
			AdPackageInstance: nil,
		}, nil
	}

	// package
	adPackage, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, int64(instance.PackageId))
	if err != nil {
		return nil, err
	}
	if adPackage == nil {
		return &pb.FindADPackageInstanceResponse{
			AdPackageInstance: nil,
		}, nil
	}

	var pbPackage *pb.ADPackage

	// network
	var pbNetwork *pb.ADNetwork
	network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(adPackage.NetworkId))
	if err != nil {
		return nil, err
	}
	if network == nil {
		return &pb.FindADPackageInstanceResponse{
			AdPackageInstance: nil,
		}, nil
	}
	pbNetwork = &pb.ADNetwork{
		Id:          int64(network.Id),
		IsOn:        network.IsOn,
		Name:        network.Name,
		Description: network.Description,
	}

	pbPackage = &pb.ADPackage{
		Id:                      int64(adPackage.Id),
		ProtectionBandwidthSize: types.Int32(adPackage.ProtectionBandwidthSize),
		ProtectionBandwidthUnit: adPackage.ProtectionBandwidthUnit,
		ServerBandwidthSize:     types.Int32(adPackage.ServerBandwidthSize),
		ServerBandwidthUnit:     adPackage.ServerBandwidthUnit,
		AdNetwork:               pbNetwork,
		IsOn:                    adPackage.IsOn,
		Summary:                 adPackage.Summary(network),
	}

	return &pb.FindADPackageInstanceResponse{AdPackageInstance: &pb.ADPackageInstance{
		Id:            int64(instance.Id),
		IsOn:          instance.IsOn,
		AdPackageId:   int64(instance.PackageId),
		NodeClusterId: int64(instance.ClusterId),
		NodeIds:       instance.DecodeNodeIds(),
		IpAddresses:   instance.DecodeIPAddresses(),
		AdPackage:     pbPackage,
	}}, nil
}

// FindAllADPackageInstances 列出单个高防产品所有实例
func (this *ADPackageInstanceService) FindAllADPackageInstances(ctx context.Context, req *pb.FindAllADPackageInstancesRequest) (*pb.FindAllADPackageInstancesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var pbInstances = []*pb.ADPackageInstance{}
	instances, err := models.SharedADPackageInstanceDAO.FindAllPackageInstances(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}

	// package
	adPackage, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}
	if adPackage == nil {
		return &pb.FindAllADPackageInstancesResponse{
			AdPackageInstances: nil,
		}, nil
	}

	var pbPackage *pb.ADPackage

	// network
	var pbNetwork *pb.ADNetwork
	network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(adPackage.NetworkId))
	if err != nil {
		return nil, err
	}
	if network != nil {
		pbNetwork = &pb.ADNetwork{
			Id:          int64(network.Id),
			IsOn:        network.IsOn,
			Name:        network.Name,
			Description: network.Description,
		}
	}

	pbPackage = &pb.ADPackage{
		Id:                      int64(adPackage.Id),
		ProtectionBandwidthSize: types.Int32(adPackage.ProtectionBandwidthSize),
		ProtectionBandwidthUnit: adPackage.ProtectionBandwidthUnit,
		ServerBandwidthSize:     types.Int32(adPackage.ServerBandwidthSize),
		ServerBandwidthUnit:     adPackage.ServerBandwidthUnit,
		AdNetwork:               pbNetwork,
		IsOn:                    adPackage.IsOn,
		Summary:                 adPackage.Summary(network),
	}

	for _, instance := range instances {
		// 集群
		var pbCluster *pb.NodeCluster
		cluster, err := models.SharedNodeClusterDAO.FindClusterBasicInfo(tx, int64(instance.ClusterId), nil)
		if err != nil {
			return nil, err
		}
		if cluster != nil {
			pbCluster = &pb.NodeCluster{
				Id:   int64(cluster.Id),
				Name: cluster.Name,
				IsOn: cluster.IsOn,
			}
		}

		pbInstances = append(pbInstances, &pb.ADPackageInstance{
			Id:            int64(instance.Id),
			IsOn:          instance.IsOn,
			AdPackageId:   int64(instance.PackageId),
			NodeClusterId: int64(instance.ClusterId),
			NodeIds:       instance.DecodeNodeIds(),
			IpAddresses:   instance.DecodeIPAddresses(),
			NodeCluster:   pbCluster,
			AdPackage:     pbPackage,
		})
	}

	return &pb.FindAllADPackageInstancesResponse{AdPackageInstances: pbInstances}, nil
}

// DeleteADPackageInstance 删除实例
func (this *ADPackageInstanceService) DeleteADPackageInstance(ctx context.Context, req *pb.DeleteADPackageInstanceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedADPackageInstanceDAO.DisableADPackageInstance(tx, req.AdPackageInstanceId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountIdleADPackageInstances 计算可购的实例数量
func (this *ADPackageInstanceService) CountIdleADPackageInstances(ctx context.Context, req *pb.CountIdleADPackageInstancesRequest) (*pb.RPCCountResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedADPackageInstanceDAO.CountIdleInstances(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// CountADPackageInstances 计算实例数量
func (this *ADPackageInstanceService) CountADPackageInstances(ctx context.Context, req *pb.CountADPackageInstancesRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := models.SharedADPackageInstanceDAO.CountInstances(tx, req.UserId, req.AdNetworkId, req.AdPackageId, req.Ip)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListADPackageInstances 列出单页实例
func (this *ADPackageInstanceService) ListADPackageInstances(ctx context.Context, req *pb.ListADPackageInstancesRequest) (*pb.ListADPackageInstancesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()

	var pbInstances = []*pb.ADPackageInstance{}
	instances, err := models.SharedADPackageInstanceDAO.ListInstances(tx, req.UserId, req.AdNetworkId, req.AdPackageId, req.Ip, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		// 集群
		var pbCluster *pb.NodeCluster
		if instance.ClusterId > 0 {
			cluster, err := models.SharedNodeClusterDAO.FindClusterBasicInfo(tx, int64(instance.ClusterId), nil)
			if err != nil {
				return nil, err
			}
			if cluster != nil {
				pbCluster = &pb.NodeCluster{
					Id:   int64(cluster.Id),
					Name: cluster.Name,
					IsOn: cluster.IsOn,
				}
			}
		}

		// package
		adPackage, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, int64(instance.PackageId))
		if err != nil {
			return nil, err
		}
		if adPackage == nil {
			continue
		}

		var pbPackage *pb.ADPackage

		// network
		var pbNetwork *pb.ADNetwork
		network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(adPackage.NetworkId))
		if err != nil {
			return nil, err
		}
		if network == nil {
			continue
		}
		pbNetwork = &pb.ADNetwork{
			Id:          int64(network.Id),
			IsOn:        network.IsOn,
			Name:        network.Name,
			Description: network.Description,
		}

		pbPackage = &pb.ADPackage{
			Id:                      int64(adPackage.Id),
			ProtectionBandwidthSize: types.Int32(adPackage.ProtectionBandwidthSize),
			ProtectionBandwidthUnit: adPackage.ProtectionBandwidthUnit,
			ServerBandwidthSize:     types.Int32(adPackage.ServerBandwidthSize),
			ServerBandwidthUnit:     adPackage.ServerBandwidthUnit,
			AdNetwork:               pbNetwork,
			IsOn:                    adPackage.IsOn,
			Summary:                 adPackage.Summary(network),
		}

		// user
		var pbUser *pb.User
		user, err := instance.CurrentUser()
		if err != nil {
			return nil, err
		}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Fullname: user.Fullname,
				Username: user.Username,
			}
		}

		pbInstances = append(pbInstances, &pb.ADPackageInstance{
			Id:            int64(instance.Id),
			IsOn:          instance.IsOn,
			AdPackageId:   int64(instance.PackageId),
			NodeClusterId: int64(instance.ClusterId),
			NodeIds:       instance.DecodeNodeIds(),
			IpAddresses:   instance.DecodeIPAddresses(),
			UserId:        int64(instance.UserId),
			UserDayTo:     instance.UserDayTo,
			NodeCluster:   pbCluster,
			AdPackage:     pbPackage,
			User:          pbUser,
		})
	}

	return &pb.ListADPackageInstancesResponse{AdPackageInstances: pbInstances}, nil
}
