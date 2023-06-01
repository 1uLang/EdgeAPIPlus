// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package antiddos

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// ADPackageService 高防产品服务
type ADPackageService struct {
	services.BaseService
}

// CreateADPackage 创建高防产品
func (this *ADPackageService) CreateADPackage(ctx context.Context, req *pb.CreateADPackageRequest) (*pb.CreateADPackageResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查线路
	if req.AdNetworkId <= 0 {
		return nil, errors.New("invalid adNetworkId")
	}
	network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, req.AdNetworkId)
	if err != nil {
		return nil, err
	}
	if network == nil {
		return nil, errors.New("invalid network")
	}

	packageId, err := models.SharedADPackageDAO.CreatePackage(tx, req.AdNetworkId, req.ProtectionBandwidthSize, req.ProtectionBandwidthUnit, req.ServerBandwidthSize, req.ServerBandwidthUnit)
	if err != nil {
		return nil, err
	}
	return &pb.CreateADPackageResponse{AdPackageId: packageId}, nil
}

// UpdateADPackage 修改高防产品
func (this *ADPackageService) UpdateADPackage(ctx context.Context, req *pb.UpdateADPackageRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查线路
	if req.AdNetworkId <= 0 {
		return nil, errors.New("invalid adNetworkId")
	}
	network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, req.AdNetworkId)
	if err != nil {
		return nil, err
	}
	if network == nil {
		return nil, errors.New("invalid network")
	}

	err = models.SharedADPackageDAO.UpdatePackage(tx, req.AdPackageId, req.IsOn, req.AdNetworkId, req.ProtectionBandwidthSize, req.ProtectionBandwidthUnit, req.ServerBandwidthSize, req.ServerBandwidthUnit)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindADPackage 查找单个高防产品
func (this *ADPackageService) FindADPackage(ctx context.Context, req *pb.FindADPackageRequest) (*pb.FindADPackageResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	adPackage, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}
	if adPackage == nil {
		return &pb.FindADPackageResponse{AdPackage: nil}, nil
	}

	// 线路
	var pbNetwork *pb.ADNetwork
	var network *models.ADNetwork
	if adPackage.NetworkId > 0 {
		network, err = models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(adPackage.NetworkId))
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
	}

	return &pb.FindADPackageResponse{AdPackage: &pb.ADPackage{
		Id:                      int64(adPackage.Id),
		IsOn:                    adPackage.IsOn,
		AdNetworkId:             int64(adPackage.NetworkId),
		ProtectionBandwidthSize: types.Int32(adPackage.ProtectionBandwidthSize),
		ProtectionBandwidthUnit: adPackage.ProtectionBandwidthUnit,
		ServerBandwidthSize:     types.Int32(adPackage.ServerBandwidthSize),
		ServerBandwidthUnit:     adPackage.ServerBandwidthUnit,
		Summary:                 adPackage.Summary(network),
		AdNetwork:               pbNetwork,
	}}, nil
}

// CountADPackages 查询高防产品数量
func (this *ADPackageService) CountADPackages(ctx context.Context, req *pb.CountADPackagesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedADPackageDAO.CountAllPackages(tx, req.AdNetworkId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// CountAllIdleADPackages 查询可用的产品数量
func (this *ADPackageService) CountAllIdleADPackages(ctx context.Context, req *pb.CountAllIdleADPackages) (*pb.RPCCountResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedADPackageDAO.CountAllIdlePackages(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListADPackages 列出单页高防产品
func (this *ADPackageService) ListADPackages(ctx context.Context, req *pb.ListADPackagesRequest) (*pb.ListADPackagesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	packages, err := models.SharedADPackageDAO.ListPackages(tx, req.AdNetworkId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbPackages = []*pb.ADPackage{}

	for _, p := range packages {
		// 线路
		var pbNetwork *pb.ADNetwork
		var network *models.ADNetwork
		if p.NetworkId > 0 {
			network, err = models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(p.NetworkId))
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
		}

		pbPackages = append(pbPackages, &pb.ADPackage{
			Id:                      int64(p.Id),
			IsOn:                    p.IsOn,
			AdNetworkId:             int64(p.NetworkId),
			ProtectionBandwidthSize: int32(p.ProtectionBandwidthSize),
			ProtectionBandwidthUnit: p.ProtectionBandwidthUnit,
			ServerBandwidthSize:     int32(p.ServerBandwidthSize),
			ServerBandwidthUnit:     p.ServerBandwidthUnit,
			Summary:                 p.Summary(network),
			AdNetwork:               pbNetwork,
		})
	}

	return &pb.ListADPackagesResponse{AdPackages: pbPackages}, nil
}

// FindAllIdleADPackages 列出所有可用的高防产品
func (this *ADPackageService) FindAllIdleADPackages(ctx context.Context, req *pb.FindAllIdleADPackagesRequest) (*pb.FindAllIdleADPackagesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	packages, err := models.SharedADPackageDAO.FindAllIdlePackages(tx)
	if err != nil {
		return nil, err
	}
	var pbPackages = []*pb.ADPackage{}
	for _, p := range packages {
		// 线路
		var pbNetwork *pb.ADNetwork
		var network *models.ADNetwork
		if p.NetworkId > 0 {
			network, err = models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(p.NetworkId))
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
		}

		// 可用实例
		countIdleInstances, err := models.SharedADPackageInstanceDAO.CountIdleInstances(tx, int64(p.Id))
		if err != nil {
			return nil, err
		}

		pbPackages = append(pbPackages, &pb.ADPackage{
			Id:                          int64(p.Id),
			IsOn:                        p.IsOn,
			AdNetworkId:                 int64(p.NetworkId),
			ProtectionBandwidthSize:     int32(p.ProtectionBandwidthSize),
			ProtectionBandwidthUnit:     p.ProtectionBandwidthUnit,
			ServerBandwidthSize:         int32(p.ServerBandwidthSize),
			ServerBandwidthUnit:         p.ServerBandwidthUnit,
			Summary:                     p.Summary(network),
			AdNetwork:                   pbNetwork,
			CountIdleADPackageInstances: countIdleInstances,
		})
	}

	return &pb.FindAllIdleADPackagesResponse{AdPackages: pbPackages}, nil
}

// DeleteADPackage 删除高防产品
func (this *ADPackageService) DeleteADPackage(ctx context.Context, req *pb.DeleteADPackageRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedADPackageDAO.DisableADPackage(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
