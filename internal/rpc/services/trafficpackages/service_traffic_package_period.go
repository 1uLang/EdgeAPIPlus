// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package trafficpackages

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// TrafficPackagePeriodService 流量包有效期服务
type TrafficPackagePeriodService struct {
	services.BaseService
}

// CreateTrafficPackagePeriod 创建有效期
func (this *TrafficPackagePeriodService) CreateTrafficPackagePeriod(ctx context.Context, req *pb.CreateTrafficPackagePeriodRequest) (*pb.CreateTrafficPackagePeriodResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	periodId, err := models.SharedTrafficPackagePeriodDAO.CreatePeriod(tx, req.Count, req.Unit)
	if err != nil {
		return nil, err
	}
	return &pb.CreateTrafficPackagePeriodResponse{
		TrafficPackagePeriodId: periodId,
	}, nil
}

// UpdateTrafficPackagePeriod 修改有效期
func (this *TrafficPackagePeriodService) UpdateTrafficPackagePeriod(ctx context.Context, req *pb.UpdateTrafficPackagePeriodRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedTrafficPackagePeriodDAO.UpdatePeriod(tx, req.TrafficPackagePeriodId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteTrafficPackagePeriod 删除有效期
func (this *TrafficPackagePeriodService) DeleteTrafficPackagePeriod(ctx context.Context, req *pb.DeleteTrafficPackagePeriodRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedTrafficPackagePeriodDAO.DisableTrafficPackagePeriod(tx, req.TrafficPackagePeriodId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindTrafficPackagePeriod 查找有效期
func (this *TrafficPackagePeriodService) FindTrafficPackagePeriod(ctx context.Context, req *pb.FindTrafficPackagePeriodRequest) (*pb.FindTrafficPackagePeriodResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	period, err := models.SharedTrafficPackagePeriodDAO.FindEnabledTrafficPackagePeriod(tx, req.TrafficPackagePeriodId)
	if err != nil {
		return nil, err
	}

	if period == nil {
		return &pb.FindTrafficPackagePeriodResponse{TrafficPackagePeriod: nil}, nil
	}

	return &pb.FindTrafficPackagePeriodResponse{TrafficPackagePeriod: &pb.TrafficPackagePeriod{
		Id:     int64(period.Id),
		IsOn:   period.IsOn,
		Count:  int32(period.Count),
		Unit:   period.Unit,
		Months: int32(period.Months),
	}}, nil
}

// FindAllTrafficPackagePeriods 列出所有有效期
func (this *TrafficPackagePeriodService) FindAllTrafficPackagePeriods(ctx context.Context, req *pb.FindAllTrafficPackagePeriodsRequest) (*pb.FindAllTrafficPackagePeriodsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	periods, err := models.SharedTrafficPackagePeriodDAO.FindAllPeriods(tx)
	if err != nil {
		return nil, err
	}

	var pbPeriods = []*pb.TrafficPackagePeriod{}
	for _, period := range periods {
		pbPeriods = append(pbPeriods, &pb.TrafficPackagePeriod{
			Id:     int64(period.Id),
			IsOn:   period.IsOn,
			Count:  int32(period.Count),
			Unit:   period.Unit,
			Months: int32(period.Months),
		})
	}

	return &pb.FindAllTrafficPackagePeriodsResponse{
		TrafficPackagePeriods: pbPeriods,
	}, nil
}

// FindAllAvailableTrafficPackagePeriods 列出所有可用有效期
func (this *TrafficPackagePeriodService) FindAllAvailableTrafficPackagePeriods(ctx context.Context, req *pb.FindAllAvailableTrafficPackagePeriodsRequest) (*pb.FindAllAvailableTrafficPackagePeriodsResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	periods, err := models.SharedTrafficPackagePeriodDAO.FindAllAvailablePeriods(tx)
	if err != nil {
		return nil, err
	}

	var pbPeriods = []*pb.TrafficPackagePeriod{}
	for _, period := range periods {
		pbPeriods = append(pbPeriods, &pb.TrafficPackagePeriod{
			Id:     int64(period.Id),
			IsOn:   period.IsOn,
			Count:  int32(period.Count),
			Unit:   period.Unit,
			Months: int32(period.Months),
		})
	}

	return &pb.FindAllAvailableTrafficPackagePeriodsResponse{
		TrafficPackagePeriods: pbPeriods,
	}, nil
}
