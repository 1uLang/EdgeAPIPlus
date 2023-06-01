// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package trafficpackages

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// TrafficPackageService 流量包服务
type TrafficPackageService struct {
	services.BaseService
}

// CreateTrafficPackage 创建流量包
func (this *TrafficPackageService) CreateTrafficPackage(ctx context.Context, req *pb.CreateTrafficPackageRequest) (*pb.CreateTrafficPackageResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO limit package size (< 8000PB)

	var tx = this.NullTx()
	packageId, err := models.SharedTrafficPackageDAO.CreatePackage(tx, req.Size, req.Unit)
	if err != nil {
		return nil, err
	}
	return &pb.CreateTrafficPackageResponse{TrafficPackageId: packageId}, nil
}

// UpdateTrafficPackage 修改流量包
func (this *TrafficPackageService) UpdateTrafficPackage(ctx context.Context, req *pb.UpdateTrafficPackageRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO limit package size (< 8000PB)

	var tx = this.NullTx()
	err = models.SharedTrafficPackageDAO.UpdatePackage(tx, req.TrafficPackageId, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteTrafficPackage 删除流量包
func (this *TrafficPackageService) DeleteTrafficPackage(ctx context.Context, req *pb.DeleteTrafficPackageRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedTrafficPackageDAO.DisableTrafficPackage(tx, req.TrafficPackageId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindTrafficPackage 查找流量包
func (this *TrafficPackageService) FindTrafficPackage(ctx context.Context, req *pb.FindTrafficPackageRequest) (*pb.FindTrafficPackageResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	p, err := models.SharedTrafficPackageDAO.FindEnabledTrafficPackage(tx, req.TrafficPackageId)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return &pb.FindTrafficPackageResponse{
			TrafficPackage: nil,
		}, nil
	}

	return &pb.FindTrafficPackageResponse{
		TrafficPackage: &pb.TrafficPackage{
			Id:    int64(p.Id),
			Size:  int32(p.Size),
			Unit:  p.Unit,
			Bytes: int64(p.Bytes),
			IsOn:  p.IsOn,
		}}, nil
}

// FindAllTrafficPackages 查找所有流量包
func (this *TrafficPackageService) FindAllTrafficPackages(ctx context.Context, req *pb.FindAllTrafficPackagesRequest) (*pb.FindAllTrafficPackagesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	packages, err := models.SharedTrafficPackageDAO.FindAllPackages(tx)
	if err != nil {
		return nil, err
	}
	var pbPackages = []*pb.TrafficPackage{}
	for _, p := range packages {
		pbPackages = append(pbPackages, &pb.TrafficPackage{
			Id:    int64(p.Id),
			Size:  int32(p.Size),
			Unit:  p.Unit,
			Bytes: int64(p.Bytes),
			IsOn:  p.IsOn,
		})
	}
	return &pb.FindAllTrafficPackagesResponse{
		TrafficPackages: pbPackages,
	}, nil
}

// FindAllAvailableTrafficPackages 查找所有可用流量包
func (this *TrafficPackageService) FindAllAvailableTrafficPackages(ctx context.Context, req *pb.FindAllAvailableTrafficPackagesRequest) (*pb.FindAllAvailableTrafficPackagesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	packages, err := models.SharedTrafficPackageDAO.FindAllAvailablePackages(tx)
	if err != nil {
		return nil, err
	}
	var pbPackages = []*pb.TrafficPackage{}
	for _, p := range packages {
		pbPackages = append(pbPackages, &pb.TrafficPackage{
			Id:    int64(p.Id),
			Size:  int32(p.Size),
			Unit:  p.Unit,
			Bytes: int64(p.Bytes),
			IsOn:  p.IsOn,
		})
	}
	return &pb.FindAllAvailableTrafficPackagesResponse{
		TrafficPackages: pbPackages,
	}, nil
}
