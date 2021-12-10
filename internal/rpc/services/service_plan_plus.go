// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// PlanService 套餐相关服务
type PlanService struct {
	BaseService
}

// CreatePlan 创建套餐
func (this *PlanService) CreatePlan(ctx context.Context, req *pb.CreatePlanRequest) (*pb.CreatePlanResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	planId, err := models.SharedPlanDAO.CreatePlan(tx, req.Name, req.ClusterId, req.TrafficLimitJSON, req.FeaturesJSON, req.PriceType, req.TrafficPriceJSON, req.MonthlyPrice, req.SeasonallyPrice, req.YearlyPrice)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePlanResponse{PlanId: planId}, nil
}

// UpdatePlan 修改套餐
func (this *PlanService) UpdatePlan(ctx context.Context, req *pb.UpdatePlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedPlanDAO.UpdatePlan(tx, req.PlanId, req.Name, req.IsOn, req.ClusterId, req.TrafficLimitJSON, req.FeaturesJSON, req.PriceType, req.TrafficPriceJSON, req.MonthlyPrice, req.SeasonallyPrice, req.YearlyPrice)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeletePlan 删除套餐
func (this *PlanService) DeletePlan(ctx context.Context, req *pb.DeletePlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedPlanDAO.DisablePlan(tx, req.PlanId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledPlan 查找单个套餐
func (this *PlanService) FindEnabledPlan(ctx context.Context, req *pb.FindEnabledPlanRequest) (*pb.FindEnabledPlanResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, req.PlanId)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledPlanResponse{
		Plan: &pb.Plan{
			Id:               int64(plan.Id),
			IsOn:             plan.IsOn == 1,
			Name:             plan.Name,
			ClusterId:        int64(plan.ClusterId),
			TrafficLimitJSON: []byte(plan.TrafficLimit),
			FeaturesJSON:     []byte(plan.Features),
			PriceType:        plan.PriceType,
			TrafficPriceJSON: []byte(plan.TrafficPrice),
			MonthlyPrice:     float32(plan.MonthlyPrice),
			SeasonallyPrice:  float32(plan.SeasonallyPrice),
			YearlyPrice:      float32(plan.YearlyPrice),
		},
	}, nil
}

// CountAllEnabledPlans 计算套餐数量
func (this *PlanService) CountAllEnabledPlans(ctx context.Context, req *pb.CountAllEnabledPlansRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedPlanDAO.CountAllEnabledPlans(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledPlans 列出单页套餐
func (this *PlanService) ListEnabledPlans(ctx context.Context, req *pb.ListEnabledPlansRequest) (*pb.ListEnabledPlansResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	plans, err := models.SharedPlanDAO.ListEnabledPlans(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbPlans = []*pb.Plan{}
	for _, plan := range plans {
		pbPlans = append(pbPlans, &pb.Plan{
			Id:               int64(plan.Id),
			IsOn:             plan.IsOn == 1,
			Name:             plan.Name,
			ClusterId:        int64(plan.ClusterId),
			TrafficLimitJSON: []byte(plan.TrafficLimit),
			FeaturesJSON:     []byte(plan.Features),
			PriceType:        plan.PriceType,
			TrafficPriceJSON: []byte(plan.TrafficPrice),
			MonthlyPrice:     float32(plan.MonthlyPrice),
			SeasonallyPrice:  float32(plan.SeasonallyPrice),
			YearlyPrice:      float32(plan.YearlyPrice),
		})
	}

	return &pb.ListEnabledPlansResponse{Plans: pbPlans}, nil
}

// SortPlans 对套餐进行排序
func (this *PlanService) SortPlans(ctx context.Context, req *pb.SortPlansRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedPlanDAO.SortPlans(tx, req.PlanIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
