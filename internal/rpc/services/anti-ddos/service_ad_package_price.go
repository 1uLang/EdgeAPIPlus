// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package antiddos

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type ADPackagePriceService struct {
	services.BaseService
}

// UpdateADPackagePrice 设置高防产品价格
func (this *ADPackagePriceService) UpdateADPackagePrice(ctx context.Context, req *pb.UpdateADPackagePriceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查各项参数的有效性

	err = models.SharedADPackagePriceDAO.UpdatePackagePrice(tx, req.AdPackageId, req.AdPackagePeriodId, req.Price)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindADPackagePrice 获取单个高防产品具体价格
func (this *ADPackagePriceService) FindADPackagePrice(ctx context.Context, req *pb.FindADPackagePriceRequest) (*pb.FindADPackagePriceResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	price, err := models.SharedADPackagePriceDAO.FindPackagePrice(tx, req.AdPackageId, req.AdPackagePeriodId)
	if err != nil {
		return nil, err
	}
	return &pb.FindADPackagePriceResponse{
		Price:  price,
		Amount: float64(req.Count) * price,
	}, nil
}

// CountADPackagePrices 计算高防产品价格项数量
func (this *ADPackagePriceService) CountADPackagePrices(ctx context.Context, req *pb.CountADPackagePricesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedADPackagePriceDAO.CountPackagePrices(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// FindADPackagePrices 查找高防产品价格
func (this *ADPackagePriceService) FindADPackagePrices(ctx context.Context, req *pb.FindADPackagePricesRequest) (*pb.FindADPackagePricesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	prices, err := models.SharedADPackagePriceDAO.FindPackagePrices(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}

	var pbPrices = []*pb.ADPackagePrice{}
	for _, price := range prices {
		pbPrices = append(pbPrices, &pb.ADPackagePrice{
			AdPackageId:       int64(price.PackageId),
			AdPackagePeriodId: int64(price.PeriodId),
			Price:             price.Price,
		})
	}
	return &pb.FindADPackagePricesResponse{
		AdPackagePrices: pbPrices,
	}, nil
}

// FindAllADPackagePrices 查找所有高防产品价格
func (this *ADPackagePriceService) FindAllADPackagePrices(ctx context.Context, req *pb.FindAllADPackagePricesRequest) (*pb.FindAllADPackagePricesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	prices, err := models.SharedADPackagePriceDAO.FindAllPackagePrices(tx)
	if err != nil {
		return nil, err
	}

	var pbPrices = []*pb.ADPackagePrice{}
	for _, price := range prices {
		pbPrices = append(pbPrices, &pb.ADPackagePrice{
			AdPackageId:       int64(price.PackageId),
			AdPackagePeriodId: int64(price.PeriodId),
			Price:             price.Price,
		})
	}
	return &pb.FindAllADPackagePricesResponse{
		AdPackagePrices: pbPrices,
	}, nil
}
