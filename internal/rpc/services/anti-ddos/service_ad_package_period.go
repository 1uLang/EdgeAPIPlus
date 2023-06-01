// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package antiddos

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ADPackagePeriodService 高防实例有效期服务
type ADPackagePeriodService struct {
	services.BaseService
}

// CreateADPackagePeriod 创建有效期
func (this *ADPackagePeriodService) CreateADPackagePeriod(ctx context.Context, req *pb.CreateADPackagePeriodRequest) (*pb.CreateADPackagePeriodResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	periodId, err := models.SharedADPackagePeriodDAO.CreatePeriod(tx, req.Count, req.Unit)
	if err != nil {
		return nil, err
	}
	return &pb.CreateADPackagePeriodResponse{
		AdPackagePeriodId: periodId,
	}, nil
}

// UpdateADPackagePeriod 修改有效期
func (this *ADPackagePeriodService) UpdateADPackagePeriod(ctx context.Context, req *pb.UpdateADPackagePeriodRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedADPackagePeriodDAO.UpdatePeriod(tx, req.AdPackagePeriodId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteADPackagePeriod 删除有效期
func (this *ADPackagePeriodService) DeleteADPackagePeriod(ctx context.Context, req *pb.DeleteADPackagePeriodRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedADPackagePeriodDAO.DisableADPackagePeriod(tx, req.AdPackagePeriodId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindADPackagePeriod 查找有效期
func (this *ADPackagePeriodService) FindADPackagePeriod(ctx context.Context, req *pb.FindADPackagePeriodRequest) (*pb.FindADPackagePeriodResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	period, err := models.SharedADPackagePeriodDAO.FindEnabledADPackagePeriod(tx, req.AdPackagePeriodId)
	if err != nil {
		return nil, err
	}

	if period == nil {
		return &pb.FindADPackagePeriodResponse{AdPackagePeriod: nil}, nil
	}

	return &pb.FindADPackagePeriodResponse{AdPackagePeriod: &pb.ADPackagePeriod{
		Id:     int64(period.Id),
		IsOn:   period.IsOn,
		Count:  int32(period.Count),
		Unit:   period.Unit,
		Months: int32(period.Months),
	}}, nil
}

// FindAllADPackagePeriods 列出所有有效期
func (this *ADPackagePeriodService) FindAllADPackagePeriods(ctx context.Context, req *pb.FindAllADPackagePeriodsRequest) (*pb.FindAllADPackagePeriodsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	periods, err := models.SharedADPackagePeriodDAO.FindAllPeriods(tx)
	if err != nil {
		return nil, err
	}

	var pbPeriods = []*pb.ADPackagePeriod{}
	for _, period := range periods {
		pbPeriods = append(pbPeriods, &pb.ADPackagePeriod{
			Id:     int64(period.Id),
			IsOn:   period.IsOn,
			Count:  int32(period.Count),
			Unit:   period.Unit,
			Months: int32(period.Months),
		})
	}

	return &pb.FindAllADPackagePeriodsResponse{
		AdPackagePeriods: pbPeriods,
	}, nil
}

// FindAllAvailableADPackagePeriods 列出所有可用有效期
func (this *ADPackagePeriodService) FindAllAvailableADPackagePeriods(ctx context.Context, req *pb.FindAllAvailableADPackagePeriodsRequest) (*pb.FindAllAvailableADPackagePeriodsResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	periods, err := models.SharedADPackagePeriodDAO.FindAllAvailablePeriods(tx)
	if err != nil {
		return nil, err
	}

	var pbPeriods = []*pb.ADPackagePeriod{}
	for _, period := range periods {
		pbPeriods = append(pbPeriods, &pb.ADPackagePeriod{
			Id:     int64(period.Id),
			IsOn:   period.IsOn,
			Count:  int32(period.Count),
			Unit:   period.Unit,
			Months: int32(period.Months),
		})
	}

	return &pb.FindAllAvailableADPackagePeriodsResponse{
		AdPackagePeriods: pbPeriods,
	}, nil
}
