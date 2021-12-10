// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

//go:build plus
// +build plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strings"
	"time"
)

// UserPlanService 用户购买的套餐
type UserPlanService struct {
	BaseService
}

// BuyUserPlan 添加已购套餐
func (this *UserPlanService) BuyUserPlan(ctx context.Context, req *pb.BuyUserPlanRequest) (*pb.BuyUserPlanResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if req.CountPeriod <= 0 {
		req.CountPeriod = 1
	}

	var userPlanId int64
	err = this.RunTx(func(tx *dbs.Tx) error {
		// 套餐
		plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, req.PlanId)
		if err != nil {
			return err
		}
		if plan == nil {
			return errors.New("can not find plan with id '" + types.String(req.PlanId) + "'")
		}

		// 周期
		var dayTo = req.DayTo
		if plan.PriceType == serverconfigs.PlanPriceTypePeriod {
			var cost float32
			var periodDescription = ""
			switch req.Period {
			case "monthly":
				dayTo = timeutil.Format("Y-m-d", time.Now().AddDate(0, int(req.CountPeriod), 0))
				cost = float32(plan.MonthlyPrice) * float32(req.CountPeriod)
				periodDescription = types.String(req.CountPeriod) + "个月"
			case "seasonally":
				dayTo = timeutil.Format("Y-m-d", time.Now().AddDate(0, int(req.CountPeriod*3), 0))
				cost = float32(plan.SeasonallyPrice) * float32(req.CountPeriod)
				periodDescription = types.String(req.CountPeriod) + "个季度"
			case "yearly":
				dayTo = timeutil.Format("Y-m-d", time.Now().AddDate(int(req.CountPeriod), 0, 0))
				cost = float32(plan.YearlyPrice) * float32(req.CountPeriod)
				periodDescription = types.String(req.CountPeriod) + "年"
			default:
				return errors.New("invalid period '" + req.Period + "'")
			}

			// 用户账户
			account, err := accounts.SharedUserAccountDAO.FindUserAccountWithUserId(tx, req.UserId)
			if err != nil {
				return err
			}
			if account == nil {
				return errors.New("can not find account for user '" + types.String(req.UserId) + "'")
			}

			if float32(account.Total) < cost {
				return errors.New("not enough quota to buy")
			}

			// 扣费
			err = accounts.SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), -cost, userconfigs.AccountEventTypeBuyPlan, "购买套餐："+types.String(plan.Name)+"/"+periodDescription+"/到期时间："+dayTo, maps.Map{"planId": plan.Id})
			if err != nil {
				return err
			}
		} else if plan.PriceType == serverconfigs.PlanPriceTypeTraffic {
			// DO NOTHING
		} else {
			return errors.New("price type '" + plan.PriceType + "' is not supported yet")
		}

		userPlanId, err = models.SharedUserPlanDAO.CreateUserPlan(tx, req.UserId, req.PlanId, dayTo)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.BuyUserPlanResponse{UserPlanId: userPlanId}, nil
}

// RenewUserPlan 续费套餐
func (this *UserPlanService) RenewUserPlan(ctx context.Context, req *pb.RenewUserPlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if req.CountPeriod <= 0 {
		req.CountPeriod = 1
	}

	var tx = this.NullTx()
	userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, req.UserPlanId, nil)
	if err != nil {
		return nil, err
	}
	if userPlan == nil || userPlan.State != models.UserPlanStateEnabled {
		return nil, errors.New("can not find user plan to renew")
	}

	var planId = int64(userPlan.PlanId)
	var userId = int64(userPlan.UserId)

	// 套餐
	plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, planId)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("can not find plan with id '" + types.String(planId) + "'")
	}

	if len(userPlan.DayTo) == 0 {
		userPlan.DayTo = timeutil.Format("Y-m-d")
	}
	var pieces = strings.Split(userPlan.DayTo, "-")
	if len(pieces) != 3 {
		return nil, errors.New("invalid 'dayTo': " + userPlan.DayTo)
	}
	var year = types.Int(pieces[0])
	var month = types.Int(pieces[1])
	var day = types.Int(pieces[2])
	var startTime = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)

	err = this.RunTx(func(tx *dbs.Tx) error {
		// 周期
		var dayTo = req.DayTo
		if plan.PriceType == serverconfigs.PlanPriceTypePeriod {
			var cost float32
			var periodDescription = ""
			switch req.Period {
			case "monthly":
				dayTo = timeutil.Format("Y-m-d", startTime.AddDate(0, int(req.CountPeriod), 0))
				cost = float32(plan.MonthlyPrice) * float32(req.CountPeriod)
				periodDescription = types.String(req.CountPeriod) + "个月"
			case "seasonally":
				dayTo = timeutil.Format("Y-m-d", startTime.AddDate(0, int(req.CountPeriod*3), 0))
				cost = float32(plan.SeasonallyPrice) * float32(req.CountPeriod)
				periodDescription = types.String(req.CountPeriod) + "个季度"
			case "yearly":
				dayTo = timeutil.Format("Y-m-d", startTime.AddDate(int(req.CountPeriod), 0, 0))
				cost = float32(plan.YearlyPrice) * float32(req.CountPeriod)
				periodDescription = types.String(req.CountPeriod) + "年"
			default:
				return errors.New("invalid period '" + req.Period + "'")
			}

			// 用户账户
			account, err := accounts.SharedUserAccountDAO.FindUserAccountWithUserId(tx, userId)
			if err != nil {
				return err
			}
			if account == nil {
				return errors.New("can not find account for user '" + types.String(userId) + "'")
			}

			if float32(account.Total) < cost {
				return errors.New("not enough quota to buy")
			}

			// 扣费
			err = accounts.SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), -cost, userconfigs.AccountEventTypeBuyPlan, "续费套餐："+types.String(plan.Name)+"/"+periodDescription+"/到期时间："+dayTo, maps.Map{"planId": plan.Id})
			if err != nil {
				return err
			}
		} else if plan.PriceType == serverconfigs.PlanPriceTypeTraffic {
			// DO NOTHING
		} else {
			return errors.New("price type '" + plan.PriceType + "' is not supported yet")
		}

		err = models.SharedUserPlanDAO.UpdateUserPlanDayTo(tx, req.UserPlanId, dayTo)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledUserPlan 查找单个已购套餐信息
func (this *UserPlanService) FindEnabledUserPlan(ctx context.Context, req *pb.FindEnabledUserPlanRequest) (*pb.FindEnabledUserPlanResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, req.UserPlanId, nil)
	if err != nil {
		return nil, err
	}

	if userPlan == nil {
		return &pb.FindEnabledUserPlanResponse{UserPlan: nil}, nil
	}

	// user
	var pbUser = &pb.User{Id: int64(userPlan.UserId)}
	user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(userPlan.UserId), nil)
	if err != nil {
		return nil, err
	}
	if user != nil {
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	// plan
	var pbPlan = &pb.Plan{Id: int64(userPlan.PlanId)}
	plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
	if err != nil {
		return nil, err
	}
	if plan != nil {
		pbPlan = &pb.Plan{
			Id:   int64(plan.Id),
			Name: plan.Name,
		}
	}

	return &pb.FindEnabledUserPlanResponse{UserPlan: &pb.UserPlan{
		Id:     int64(userPlan.Id),
		UserId: int64(userPlan.UserId),
		PlanId: int64(userPlan.PlanId),
		IsOn:   userPlan.IsOn == 1,
		DayTo:  userPlan.DayTo,
		User:   pbUser,
		Plan:   pbPlan,
	}}, nil
}

// UpdateUserPlan 修改已购套餐
func (this *UserPlanService) UpdateUserPlan(ctx context.Context, req *pb.UpdateUserPlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserPlanDAO.UpdateUserPlan(tx, req.UserPlanId, req.PlanId, req.DayTo, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteUserPlan 删除已购套餐
func (this *UserPlanService) DeleteUserPlan(ctx context.Context, req *pb.DeleteUserPlanRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserPlanDAO.DisableUserPlan(tx, req.UserPlanId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledUserPlans 计算已购套餐数
func (this *UserPlanService) CountAllEnabledUserPlans(ctx context.Context, req *pb.CountAllEnabledUserPlansRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedUserPlanDAO.CountAllEnabledUserPlans(tx, req.IsAvailable, req.IsExpired, req.ExpiringDays)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledUserPlans 列出单页已购套餐
func (this *UserPlanService) ListEnabledUserPlans(ctx context.Context, req *pb.ListEnabledUserPlansRequest) (*pb.ListEnabledUserPlansResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userPlans, err := models.SharedUserPlanDAO.ListEnabledUserPlans(tx, req.UserId, req.IsAvailable, req.IsExpired, req.ExpiringDays, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbUserPlans = []*pb.UserPlan{}
	var cacheMap = utils.NewCacheMap()
	for _, userPlan := range userPlans {
		// user
		var pbUser = &pb.User{Id: int64(userPlan.UserId)}
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(userPlan.UserId), cacheMap)
		if err != nil {
			return nil, err
		}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}

		// plan
		var pbPlan = &pb.Plan{Id: int64(userPlan.PlanId)}
		plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
		if err != nil {
			return nil, err
		}
		if plan != nil {
			pbPlan = &pb.Plan{
				Id:   int64(plan.Id),
				Name: plan.Name,
			}
		}

		pbUserPlans = append(pbUserPlans, &pb.UserPlan{
			Id:     int64(userPlan.Id),
			UserId: int64(userPlan.UserId),
			PlanId: int64(userPlan.PlanId),
			IsOn:   userPlan.IsOn == 1,
			DayTo:  userPlan.DayTo,
			User:   pbUser,
			Plan:   pbPlan,
		})
	}

	return &pb.ListEnabledUserPlansResponse{UserPlans: pbUserPlans}, nil
}

// FindAllEnabledUserPlansForServer 查找所有服务可用的套餐
func (this *UserPlanService) FindAllEnabledUserPlansForServer(ctx context.Context, req *pb.FindAllEnabledUserPlansForServerRequest) (*pb.FindAllEnabledUserPlansForServerResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userPlans, err := models.SharedUserPlanDAO.FindAllEnabledPlansForServer(tx, req.UserId, req.ServerId)
	if err != nil {
		return nil, err
	}

	var pbUserPlans = []*pb.UserPlan{}
	var cacheMap = utils.NewCacheMap()
	for _, userPlan := range userPlans {
		// user
		var pbUser = &pb.User{Id: int64(userPlan.UserId)}
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(userPlan.UserId), cacheMap)
		if err != nil {
			return nil, err
		}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}

		// plan
		var pbPlan = &pb.Plan{Id: int64(userPlan.PlanId)}
		plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
		if err != nil {
			return nil, err
		}
		if plan != nil {
			pbPlan = &pb.Plan{
				Id:   int64(plan.Id),
				Name: plan.Name,
			}
		}

		pbUserPlans = append(pbUserPlans, &pb.UserPlan{
			Id:     int64(userPlan.Id),
			UserId: int64(userPlan.UserId),
			PlanId: int64(userPlan.PlanId),
			IsOn:   userPlan.IsOn == 1,
			DayTo:  userPlan.DayTo,
			User:   pbUser,
			Plan:   pbPlan,
		})
	}
	return &pb.FindAllEnabledUserPlansForServerResponse{UserPlans: pbUserPlans}, nil
}
