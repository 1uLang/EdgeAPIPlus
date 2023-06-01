// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package trafficpackages

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strings"
)

// UserTrafficPackageService 用户流量包服务
type UserTrafficPackageService struct {
	services.BaseService
}

// CreateUserTrafficPackage 创建用户流量包
func (this *UserTrafficPackageService) CreateUserTrafficPackage(ctx context.Context, req *pb.CreateUserTrafficPackageRequest) (*pb.CreateUserTrafficPackageResponse, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查各项参数有效性
	if req.TrafficPackageId <= 0 {
		return nil, errors.New("invalid 'trafficPackageId'")
	}
	if req.NodeRegionId <= 0 {
		return nil, errors.New("invalid 'nodeRegionId'")
	}
	if req.TrafficPackagePeriodId <= 0 {
		return nil, errors.New("invalid 'trafficPackagePeriodId'")
	}
	if req.Count <= 0 {
		return nil, errors.New("invalid 'count'")
	}

	var userPackageIds = []int64{}

	for i := 1; i <= int(req.Count); i++ {
		userPackageId, err := models.SharedUserTrafficPackageDAO.CreateUserPackage(tx, req.UserId, adminId, req.TrafficPackageId, req.NodeRegionId, req.TrafficPackagePeriodId)
		if err != nil {
			return nil, err
		}
		userPackageIds = append(userPackageIds, userPackageId)
	}

	return &pb.CreateUserTrafficPackageResponse{
		UserTrafficPackageIds: userPackageIds,
	}, nil
}

// BuyUserTrafficPackage 购买用户流量包
func (this *UserTrafficPackageService) BuyUserTrafficPackage(ctx context.Context, req *pb.BuyUserTrafficPackageRequest) (*pb.BuyUserTrafficPackageResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}
	userId = req.UserId

	// TODO 检查各项参数有效性
	if req.TrafficPackageId <= 0 {
		return nil, errors.New("invalid 'trafficPackageId'")
	}
	if req.NodeRegionId <= 0 {
		return nil, errors.New("invalid 'nodeRegionId'")
	}
	if req.TrafficPackagePeriodId <= 0 {
		return nil, errors.New("invalid 'trafficPackagePeriodId'")
	}
	if req.Count <= 0 {
		return nil, errors.New("invalid 'count'")
	}
	// check count
	// TODO 改成设置的默认值
	if req.Count >= 20 {
		req.Count = 20
	}

	var userPackageIds = []int64{}

	err = this.RunTx(func(tx *dbs.Tx) error {
		// check package
		p, err := models.SharedTrafficPackageDAO.FindEnabledTrafficPackage(tx, req.TrafficPackageId)
		if err != nil {
			return err
		}
		if p == nil || !p.IsOn {
			return errors.New("invalid 'trafficPackageId'")
		}

		// check region
		region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(tx, req.NodeRegionId)
		if err != nil {
			return err
		}
		if region == nil || !region.IsOn {
			return errors.New("invalid 'nodeRegionId'")
		}

		// check period
		period, err := models.SharedTrafficPackagePeriodDAO.FindEnabledTrafficPackagePeriod(tx, req.TrafficPackagePeriodId)
		if err != nil {
			return err
		}
		if period == nil || !period.IsOn {
			return errors.New("invalid 'trafficPackagePeriodId'")
		}

		// 获取价格
		price, err := models.SharedTrafficPackagePriceDAO.FindPackagePrice(tx, req.TrafficPackageId, req.NodeRegionId, req.TrafficPackagePeriodId)
		if err != nil {
			return err
		}
		if price == 0 {
			return errors.New("invalid package price")
		}
		var amount = price * float64(req.Count)

		// 先减少余额
		account, err := accounts.SharedUserAccountDAO.FindUserAccountWithUserId(tx, userId)
		if err != nil {
			return err
		}
		if account == nil || account.Total < amount {
			return errors.New("no enough balance to buy the traffic package")
		}

		err = accounts.SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), -amount, userconfigs.AccountEventTypeBuyTrafficPackage, "购买流量包\""+types.String(p.Size)+strings.ToUpper(p.Unit)+" / "+region.Name+" / "+types.String(period.Count)+userconfigs.TrafficPackagePeriodUnitName(period.Unit)+"\" x "+types.String(req.Count), maps.Map{
			"trafficPackageId":          req.TrafficPackageId,
			"nodeRegionId":              req.NodeRegionId,
			"trafficPackagePeriodId":    req.TrafficPackagePeriodId,
			"trafficPackagePeriodCount": period.Count,
			"trafficPackagePeriodUnit":  period.Unit,
			"count":                     req.Count,
		})
		if err != nil {
			return err
		}

		for i := 1; i <= int(req.Count); i++ {
			userPackageId, err := models.SharedUserTrafficPackageDAO.CreateUserPackage(tx, req.UserId, adminId, req.TrafficPackageId, req.NodeRegionId, req.TrafficPackagePeriodId)
			if err != nil {
				return err
			}
			userPackageIds = append(userPackageIds, userPackageId)
		}

		return nil
	})
	return &pb.BuyUserTrafficPackageResponse{
		UserTrafficPackageIds: userPackageIds,
	}, nil
}

// CountUserTrafficPackages 查询当前流量包数量
func (this *UserTrafficPackageService) CountUserTrafficPackages(ctx context.Context, req *pb.CountUserTrafficPackagesRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := models.SharedUserTrafficPackageDAO.CountUserPackages(tx, req.TrafficPackageId, req.UserId, req.NodeRegionId, req.TrafficPackagePeriodId, req.ExpiresDay, req.AvailableOnly)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserTrafficPackages 列出单页流量包
func (this *UserTrafficPackageService) ListUserTrafficPackages(ctx context.Context, req *pb.ListUserTrafficPackagesRequest) (*pb.ListUserTrafficPackagesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	userPackages, err := models.SharedUserTrafficPackageDAO.ListUserPackages(tx, req.TrafficPackageId, req.UserId, req.NodeRegionId, req.TrafficPackagePeriodId, req.ExpiresDay, req.AvailableOnly, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbUserPackages = []*pb.UserTrafficPackage{}
	for _, userPackage := range userPackages {
		// package
		p, err := models.SharedTrafficPackageDAO.FindEnabledTrafficPackage(tx, int64(userPackage.PackageId))
		if err != nil {
			return nil, err
		}
		var pbPackage *pb.TrafficPackage
		if p != nil {
			pbPackage = &pb.TrafficPackage{
				Id:    int64(p.Id),
				Size:  int32(p.Size),
				Unit:  p.Unit,
				Bytes: int64(p.Bytes),
				IsOn:  p.IsOn,
			}
		}

		// node region
		region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(tx, int64(userPackage.RegionId))
		if err != nil {
			return nil, err
		}
		var pbRegion *pb.NodeRegion
		if region != nil {
			pbRegion = &pb.NodeRegion{
				Id:          int64(region.Id),
				IsOn:        region.IsOn,
				Name:        region.Name,
				Description: region.Description,
			}
		}

		// user
		var pbUser *pb.User
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(userPackage.UserId))
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

		pbUserPackages = append(pbUserPackages, &pb.UserTrafficPackage{
			Id:                        int64(userPackage.Id),
			UserId:                    int64(userPackage.UserId),
			TrafficPackageId:          int64(userPackage.PackageId),
			TotalBytes:                int64(userPackage.TotalBytes),
			UsedBytes:                 int64(userPackage.UsedBytes),
			NodeRegionId:              int64(userPackage.RegionId),
			TrafficPackagePeriodId:    int64(userPackage.PeriodId),
			TrafficPackagePeriodCount: int32(userPackage.PeriodCount),
			TrafficPackagePeriodUnit:  userPackage.PeriodUnit,
			DayFrom:                   userPackage.DayFrom,
			DayTo:                     userPackage.DayTo,
			CreatedAt:                 int64(userPackage.CreatedAt),
			TrafficPackage:            pbPackage,
			NodeRegion:                pbRegion,
			User:                      pbUser,
			CanDelete:                 userPackage.AdminId > 0,
		})
	}
	return &pb.ListUserTrafficPackagesResponse{
		UserTrafficPackages: pbUserPackages,
	}, nil
}

// DeleteUserTrafficPackage 删除流量包
func (this *UserTrafficPackageService) DeleteUserTrafficPackage(ctx context.Context, req *pb.DeleteUserTrafficPackageRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserTrafficPackageDAO.DisableUserTrafficPackage(tx, req.UserTrafficPackageId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
