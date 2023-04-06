// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package nameservers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
)

// NSPlanService DNS套餐服务
type NSPlanService struct {
	services.BaseService
}

// CreateNSPlan 创建DNS套餐
func (this *NSPlanService) CreateNSPlan(ctx context.Context, req *pb.CreateNSPlanRequest) (*pb.CreateNSPlanResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if len(req.ConfigJSON) == 0 {
		return nil, errors.New("invalid 'configJSON'")
	}

	var config = dnsconfigs.DefaultNSPlanConfig()
	err = json.Unmarshal(req.ConfigJSON, config)
	if err != nil {
		return nil, errors.New("decode 'configJSON' failed: " + err.Error())
	}

	planId, err := nameservers.SharedNSPlanDAO.CreatePlan(tx, req.Name, req.MonthlyPrice, req.YearlyPrice, config)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNSPlanResponse{NsPlanId: planId}, nil
}

// UpdateNSPlan 修改DNS套餐
func (this *NSPlanService) UpdateNSPlan(ctx context.Context, req *pb.UpdateNSPlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if len(req.ConfigJSON) == 0 {
		return nil, errors.New("invalid 'configJSON'")
	}

	var config = dnsconfigs.DefaultNSPlanConfig()
	err = json.Unmarshal(req.ConfigJSON, config)
	if err != nil {
		return nil, errors.New("decode 'configJSON' failed: " + err.Error())
	}

	err = nameservers.SharedNSPlanDAO.UpdatePlan(tx, req.NsPlanId, req.Name, req.IsOn, req.MonthlyPrice, req.YearlyPrice, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// SortNSPlanOrders 修改DNS套餐顺序
func (this *NSPlanService) SortNSPlanOrders(ctx context.Context, req *pb.SortNSPlansRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSPlanDAO.UpdatePlanOrders(tx, req.NsPlanIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllNSPlans 查找所有DNS套餐
func (this *NSPlanService) FindAllNSPlans(ctx context.Context, req *pb.FindAllNSPlansRequest) (*pb.FindAllNSPlansResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var pbPlans = []*pb.NSPlan{}
	plans, err := nameservers.SharedNSPlanDAO.FindAllPlans(tx)
	if err != nil {
		return nil, err
	}
	for _, plan := range plans {
		pbPlans = append(pbPlans, &pb.NSPlan{
			Id:           int64(plan.Id),
			Name:         plan.Name,
			IsOn:         plan.IsOn,
			MonthlyPrice: float32(plan.MonthlyPrice),
			YearlyPrice:  float32(plan.YearlyPrice),
			ConfigJSON:   plan.Config,
		})
	}

	return &pb.FindAllNSPlansResponse{NsPlans: pbPlans}, nil
}

// FindAllEnabledNSPlans 查找所有可用DNS套餐
func (this *NSPlanService) FindAllEnabledNSPlans(ctx context.Context, req *pb.FindAllEnabledNSPlansRequest) (*pb.FindAllEnabledNSPlansResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var pbPlans = []*pb.NSPlan{}
	plans, err := nameservers.SharedNSPlanDAO.FindAllEnabledPlans(tx)
	if err != nil {
		return nil, err
	}
	for _, plan := range plans {
		pbPlans = append(pbPlans, &pb.NSPlan{
			Id:           int64(plan.Id),
			Name:         plan.Name,
			IsOn:         plan.IsOn,
			MonthlyPrice: float32(plan.MonthlyPrice),
			YearlyPrice:  float32(plan.YearlyPrice),
			ConfigJSON:   plan.Config,
		})
	}

	return &pb.FindAllEnabledNSPlansResponse{NsPlans: pbPlans}, nil
}

// FindNSPlan 查找DNS套餐
func (this *NSPlanService) FindNSPlan(ctx context.Context, req *pb.FindNSPlanRequest) (*pb.FindNSPlanResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	plan, err := nameservers.SharedNSPlanDAO.FindEnabledNSPlan(tx, req.NsPlanId)
	if err != nil {
		return nil, err
	}

	if plan == nil {
		return &pb.FindNSPlanResponse{}, nil
	}

	return &pb.FindNSPlanResponse{
		NsPlan: &pb.NSPlan{
			Id:           int64(plan.Id),
			Name:         plan.Name,
			IsOn:         plan.IsOn,
			MonthlyPrice: float32(plan.MonthlyPrice),
			YearlyPrice:  float32(plan.YearlyPrice),
			ConfigJSON:   plan.Config,
		},
	}, nil
}

// DeleteNSPlan 删除DNS套餐
func (this *NSPlanService) DeleteNSPlan(ctx context.Context, req *pb.DeleteNSPlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = nameservers.SharedNSPlanDAO.DisableNSPlan(tx, req.NsPlanId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
