// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package trafficpackages

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// TrafficPackagePriceService 流量包价格服务
type TrafficPackagePriceService struct {
	services.BaseService
}

// UpdateTrafficPackagePrice 设置流量包价格
func (this *TrafficPackagePriceService) UpdateTrafficPackagePrice(ctx context.Context, req *pb.UpdateTrafficPackagePriceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查各项参数的有效性

	err = models.SharedTrafficPackagePriceDAO.UpdatePackagePrice(tx, req.TrafficPackageId, req.NodeRegionId, req.TrafficPackagePeriodId, req.Price)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindTrafficPackagePrice 获取单个流量包具体价格
func (this *TrafficPackagePriceService) FindTrafficPackagePrice(ctx context.Context, req *pb.FindTrafficPackagePriceRequest) (*pb.FindTrafficPackagePriceResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	price, err := models.SharedTrafficPackagePriceDAO.FindPackagePrice(tx, req.TrafficPackageId, req.NodeRegionId, req.TrafficPackagePeriodId)
	if err != nil {
		return nil, err
	}
	return &pb.FindTrafficPackagePriceResponse{
		Price:  price,
		Amount: float64(req.Count) * price,
	}, nil
}

// CountTrafficPackagePrices 计算流量包价格项数量
func (this *TrafficPackagePriceService) CountTrafficPackagePrices(ctx context.Context, req *pb.CountTrafficPackagePricesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedTrafficPackagePriceDAO.CountPackagePrices(tx, req.TrafficPackageId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// FindTrafficPackagePrices 查找流量包价格
func (this *TrafficPackagePriceService) FindTrafficPackagePrices(ctx context.Context, req *pb.FindTrafficPackagePricesRequest) (*pb.FindTrafficPackagePricesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	prices, err := models.SharedTrafficPackagePriceDAO.FindPackagePrices(tx, req.TrafficPackageId)
	if err != nil {
		return nil, err
	}

	var pbPrices = []*pb.TrafficPackagePrice{}
	for _, price := range prices {
		pbPrices = append(pbPrices, &pb.TrafficPackagePrice{
			TrafficPackageId:       int64(price.PackageId),
			NodeRegionId:           int64(price.RegionId),
			TrafficPackagePeriodId: int64(price.PeriodId),
			Price:                  price.Price,
		})
	}
	return &pb.FindTrafficPackagePricesResponse{
		TrafficPackagePrices: pbPrices,
	}, nil
}

// FindAllTrafficPackagePrices 查找所有流量包价格
func (this *TrafficPackagePriceService) FindAllTrafficPackagePrices(ctx context.Context, req *pb.FindAllTrafficPackagePricesRequest) (*pb.FindAllTrafficPackagePricesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	prices, err := models.SharedTrafficPackagePriceDAO.FindAllPackagePrices(tx)
	if err != nil {
		return nil, err
	}

	var pbPrices = []*pb.TrafficPackagePrice{}
	for _, price := range prices {
		pbPrices = append(pbPrices, &pb.TrafficPackagePrice{
			TrafficPackageId:       int64(price.PackageId),
			NodeRegionId:           int64(price.RegionId),
			TrafficPackagePeriodId: int64(price.PeriodId),
			Price:                  price.Price,
		})
	}
	return &pb.FindAllTrafficPackagePricesResponse{
		TrafficPackagePrices: pbPrices,
	}, nil
}
