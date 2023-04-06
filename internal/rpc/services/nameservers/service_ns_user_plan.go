// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package nameservers

import (
	"context"
	"errors"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"regexp"
)

// NSUserPlanService 用户DNS套餐服务
type NSUserPlanService struct {
	services.BaseService
	pb.UnimplementedNSUserPlanServiceServer
}

// CreateNSUserPlan 创建用户套餐
func (this *NSUserPlanService) CreateNSUserPlan(ctx context.Context, req *pb.CreateNSUserPlanRequest) (*pb.CreateNSUserPlanResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var dayFrom = req.DayFrom
	var dayTo = req.DayTo

	var dayReg = regexp.MustCompile(`^\d{8}$`)
	if !dayReg.MatchString(dayFrom) {
		return nil, errors.New("invalid dayFrom: " + dayFrom)
	}
	if !dayReg.MatchString(dayTo) {
		return nil, errors.New("invalid dayTo: " + dayTo)
	}

	if !lists.ContainsString([]string{nameservers.NSUserPlanPeriodUnitMonthly, nameservers.NSUserPlanPeriodUnitYearly}, req.PeriodUnit) {
		return nil, errors.New("invalid periodUnit: " + req.PeriodUnit)
	}

	// 检查plan是否存在
	var tx = this.NullTx()
	existPlan, err := nameservers.SharedNSPlanDAO.ExistPlan(tx, req.NsPlanId)
	if err != nil {
		return nil, err
	}
	if !existPlan {
		return nil, errors.New("plan '" + types.String(req.NsPlanId) + "' not found")
	}

	// 用户Plan是否存在
	var resultUserPlanId int64
	err = this.RunTx(func(tx *dbs.Tx) error {
		userPlan, err := nameservers.SharedNSUserPlanDAO.FindUserPlan(tx, req.UserId)
		if err != nil {
			return err
		}
		if userPlan == nil {
			userPlanId, err := nameservers.SharedNSUserPlanDAO.CreateUserPlan(tx, req.UserId, req.NsPlanId, dayFrom, dayTo, req.PeriodUnit)
			if err != nil {
				return err
			}
			resultUserPlanId = userPlanId
		} else {
			err = nameservers.SharedNSUserPlanDAO.UpdateUserPlan(tx, int64(userPlan.Id), req.NsPlanId, dayFrom, dayTo, req.PeriodUnit)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateNSUserPlanResponse{NsUserPlanId: resultUserPlanId}, nil
}

// UpdateNSUserPlan 修改用户套餐
func (this *NSUserPlanService) UpdateNSUserPlan(ctx context.Context, req *pb.UpdateNSUserPlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var dayFrom = req.DayFrom
	var dayTo = req.DayTo

	var dayReg = regexp.MustCompile(`^\d{8}$`)
	if !dayReg.MatchString(dayFrom) {
		return nil, errors.New("invalid dayFrom: " + dayFrom)
	}
	if !dayReg.MatchString(dayTo) {
		return nil, errors.New("invalid dayTo: " + dayTo)
	}

	if !lists.ContainsString([]string{nameservers.NSUserPlanPeriodUnitMonthly, nameservers.NSUserPlanPeriodUnitYearly}, req.PeriodUnit) {
		return nil, errors.New("invalid periodUnit: " + req.PeriodUnit)
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSUserPlanDAO.UpdateUserPlan(tx, req.NsUserPlanId, req.NsPlanId, dayFrom, dayTo, req.PeriodUnit)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteNSUserPlan 删除用户套餐
func (this *NSUserPlanService) DeleteNSUserPlan(ctx context.Context, req *pb.DeleteNSUserPlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = nameservers.SharedNSUserPlanDAO.DisableNSUserPlan(tx, req.NsUserPlanId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNSUserPlan 读取用户套餐
func (this *NSUserPlanService) FindNSUserPlan(ctx context.Context, req *pb.FindNSUserPlanRequest) (*pb.FindNSUserPlanResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId
	}

	var userPlan *nameservers.NSUserPlan
	if req.NsUserPlanId > 0 {
		userPlan, err = nameservers.SharedNSUserPlanDAO.FindEnabledNSUserPlan(tx, req.NsUserPlanId)
	} else if req.UserId > 0 {
		userPlan, err = nameservers.SharedNSUserPlanDAO.FindUserPlan(tx, req.UserId)
	} else {
		return &pb.FindNSUserPlanResponse{NsUserPlan: nil}, nil
	}

	if err != nil {
		return nil, err
	}
	if userPlan == nil {
		return &pb.FindNSUserPlanResponse{NsUserPlan: nil}, nil
	}

	plan, err := nameservers.SharedNSPlanDAO.FindEnabledNSPlan(tx, int64(userPlan.PlanId))
	if err != nil {
		return nil, err
	}

	if plan == nil {
		return &pb.FindNSUserPlanResponse{NsUserPlan: nil}, nil
	}

	// user
	user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(userPlan.UserId))
	if err != nil {
		return nil, err
	}
	var pbUser *pb.User
	if user != nil {
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	return &pb.FindNSUserPlanResponse{
		NsUserPlan: &pb.NSUserPlan{
			Id:         int64(userPlan.Id),
			NsPlanId:   int64(userPlan.PlanId),
			DayFrom:    userPlan.DayFrom,
			DayTo:      userPlan.DayTo,
			PeriodUnit: userPlan.PeriodUnit,
			NsPlan: &pb.NSPlan{
				Id:           int64(plan.Id),
				Name:         plan.Name,
				IsOn:         plan.IsOn,
				MonthlyPrice: float32(plan.MonthlyPrice),
				YearlyPrice:  float32(plan.YearlyPrice),
				ConfigJSON:   plan.Config,
			},
			User: pbUser,
		},
	}, nil
}

// CountNSUserPlans 计算用户套餐数量
func (this *NSUserPlanService) CountNSUserPlans(ctx context.Context, req *pb.CountNSUserPlansRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := nameservers.SharedNSUserPlanDAO.CountUserPlans(tx, req.UserId, req.NsPlanId, req.PeriodUnit, req.IsExpired, req.ExpireDays)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListNSUserPlans 列出单页套餐
func (this *NSUserPlanService) ListNSUserPlans(ctx context.Context, req *pb.ListNSUserPlansRequest) (*pb.ListNSUserPlansResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	userPlans, err := nameservers.SharedNSUserPlanDAO.ListUserPlans(tx, req.UserId, req.NsPlanId, req.PeriodUnit, req.IsExpired, req.ExpireDays, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbUserPlans = []*pb.NSUserPlan{}
	for _, userPlan := range userPlans {
		// plan
		plan, err := nameservers.SharedNSPlanDAO.FindEnabledNSPlan(tx, int64(userPlan.PlanId))
		if err != nil {
			return nil, err
		}
		if plan == nil {
			continue
		}

		// user
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(userPlan.UserId))
		if err != nil {
			return nil, err
		}
		var pbUser *pb.User
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}

		pbUserPlans = append(pbUserPlans, &pb.NSUserPlan{
			Id:         int64(userPlan.Id),
			NsPlanId:   int64(userPlan.PlanId),
			UserId:     int64(userPlan.UserId),
			DayFrom:    userPlan.DayFrom,
			DayTo:      userPlan.DayTo,
			PeriodUnit: userPlan.PeriodUnit,
			NsPlan: &pb.NSPlan{
				Id:           int64(plan.Id),
				Name:         plan.Name,
				IsOn:         plan.IsOn,
				MonthlyPrice: float32(plan.MonthlyPrice),
				YearlyPrice:  float32(plan.YearlyPrice),
				ConfigJSON:   plan.Config,
			},
			User: pbUser,
		})
	}

	return &pb.ListNSUserPlansResponse{
		NsUserPlans: pbUserPlans,
	}, nil
}
