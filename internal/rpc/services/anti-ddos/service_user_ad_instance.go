// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package antiddos

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strings"
)

// UserADInstanceService 用户高防实例服务
type UserADInstanceService struct {
	services.BaseService
}

// CreateUserADInstance 创建用户高防实例
func (this *UserADInstanceService) CreateUserADInstance(ctx context.Context, req *pb.CreateUserADInstanceRequest) (*pb.CreateUserADInstanceResponse, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 高防产品
	if req.AdPackageId <= 0 {
		return nil, errors.New("invalid 'adPackageId'")
	}
	adPackage, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}
	if adPackage == nil {
		return nil, errors.New("invalid 'adPackage'")
	}

	// 有效期选项
	if req.AdPackagePeriodId <= 0 {
		return nil, errors.New("invalid 'adPackagePeriodId'")
	}

	period, err := models.SharedADPackagePeriodDAO.FindEnabledADPackagePeriod(tx, req.AdPackagePeriodId)
	if err != nil {
		return nil, err
	}
	if period == nil || !period.IsOn {
		return nil, errors.New("could not find instance period with id '" + types.String(req.AdPackagePeriodId) + "'")
	}
	_, dayTo := period.DayPeriod()

	// 数量
	if req.Count <= 0 {
		return nil, errors.New("invalid 'count'")
	}

	instances, err := models.SharedADPackageInstanceDAO.FindIdlePackageInstances(tx, req.AdPackageId, req.Count)
	if err != nil {
		return nil, err
	}

	var countInstances = int32(len(instances))
	if countInstances < req.Count {
		return nil, errors.New("no enough instances")
	}
	if countInstances > req.Count {
		instances = instances[:req.Count]
	}

	var userInstanceIds = []int64{}
	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, instance := range instances {
			var instanceId = int64(instance.Id)
			userInstanceId, err := models.SharedUserADInstanceDAO.CreateUserInstance(tx, req.UserId, adminId, instanceId, req.AdPackagePeriodId)
			if err != nil {
				return err
			}

			err = models.SharedADPackageInstanceDAO.UpdateInstanceUser(tx, instanceId, req.UserId, dayTo, userInstanceId)
			if err != nil {
				return err
			}

			userInstanceIds = append(userInstanceIds, userInstanceId)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserADInstanceResponse{
		UserADInstanceIds: userInstanceIds,
	}, nil
}

// BuyUserADInstance 购买用户高防实例
func (this *UserADInstanceService) BuyUserADInstance(ctx context.Context, req *pb.BuyUserADInstanceRequest) (*pb.BuyUserADInstanceResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}
	userId = req.UserId

	var tx = this.NullTx()

	// 高防产品
	if req.AdPackageId <= 0 {
		return nil, errors.New("invalid 'adPackageId'")
	}
	adPackage, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, req.AdPackageId)
	if err != nil {
		return nil, err
	}
	if adPackage == nil {
		return nil, errors.New("invalid 'adPackage'")
	}

	// 线路
	network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(adPackage.NetworkId))
	if err != nil {
		return nil, err
	}
	if network == nil {
		return nil, errors.New("invalid 'network'")
	}

	// 有效期选项
	if req.AdPackagePeriodId <= 0 {
		return nil, errors.New("invalid 'adPackagePeriodId'")
	}

	period, err := models.SharedADPackagePeriodDAO.FindEnabledADPackagePeriod(tx, req.AdPackagePeriodId)
	if err != nil {
		return nil, err
	}
	if period == nil || !period.IsOn {
		return nil, errors.New("could not find instance period with id '" + types.String(req.AdPackagePeriodId) + "'")
	}
	_, dayTo := period.DayPeriod()

	// 数量
	if req.Count <= 0 {
		return nil, errors.New("invalid 'count'")
	}

	var userInstanceIds = []int64{}

	err = this.RunTx(func(tx *dbs.Tx) error {
		var packageId = int64(adPackage.Id)
		instances, err := models.SharedADPackageInstanceDAO.FindIdlePackageInstances(tx, packageId, req.Count)
		if err != nil {
			return err
		}

		var countInstances = int32(len(instances))
		if countInstances < req.Count {
			return errors.New("no enough instances")
		}
		if countInstances > req.Count {
			instances = instances[:req.Count]
		}

		// 获取价格
		price, err := models.SharedADPackagePriceDAO.FindPackagePrice(tx, packageId, req.AdPackagePeriodId)
		if err != nil {
			return err
		}
		if price <= 0 {
			return errors.New("invalid package price, id:" + types.String(packageId))
		}
		var amount = price * float64(req.Count)

		// 先减少余额
		account, err := accounts.SharedUserAccountDAO.FindUserAccountWithUserId(tx, userId)
		if err != nil {
			return err
		}
		if account == nil || account.Total < amount {
			return errors.New("no enough balance to buy the package")
		}

		err = accounts.SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), -amount, userconfigs.AccountEventTypeBuyAntiDDoSPackage, "购买DDoS高防实例，线路："+network.Name+" / 防护带宽："+types.String(adPackage.ProtectionBandwidthSize)+userconfigs.ADPackageSizeFullUnit(adPackage.ProtectionBandwidthUnit)+" / 业务带宽："+types.String(adPackage.ServerBandwidthSize)+userconfigs.ADPackageSizeFullUnit(adPackage.ServerBandwidthUnit)+" / "+types.String(period.Count)+userconfigs.ADPackagePeriodUnitName(period.Unit)+"\" x "+types.String(req.Count), maps.Map{
			"adNetworkId":             network.Id,
			"adPackageId":             packageId,
			"protectionBandwidthSize": adPackage.ProtectionBandwidthSize,
			"protectionBandwidthUnit": adPackage.ProtectionBandwidthUnit,
			"serverBandwidthSize":     adPackage.ServerBandwidthSize,
			"serverBandwidthUnit":     adPackage.ServerBandwidthUnit,
			"adPackagePeriodId":       req.AdPackagePeriodId,
			"adPackagePeriodCount":    period.Count,
			"adPackagePeriodUnit":     period.Unit,
			"count":                   req.Count,
		})
		if err != nil {
			return err
		}

		for _, instance := range instances {
			var instanceId = int64(instance.Id)
			userInstanceId, err := models.SharedUserADInstanceDAO.CreateUserInstance(tx, req.UserId, adminId, instanceId, req.AdPackagePeriodId)
			if err != nil {
				return err
			}

			err = models.SharedADPackageInstanceDAO.UpdateInstanceUser(tx, instanceId, req.UserId, dayTo, userInstanceId)
			if err != nil {
				return err
			}

			userInstanceIds = append(userInstanceIds, userInstanceId)
		}

		return nil
	})
	return &pb.BuyUserADInstanceResponse{
		UserADInstanceIds: userInstanceIds,
	}, nil
}

// CountUserADInstances 查询当前高防实例数量
func (this *UserADInstanceService) CountUserADInstances(ctx context.Context, req *pb.CountUserADInstancesRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := models.SharedUserADInstanceDAO.CountUserInstances(tx, req.AdNetworkId, req.UserId, req.AdPackagePeriodId, req.ExpiresDay, req.AvailableOnly)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserADInstances 列出单页高防实例
func (this *UserADInstanceService) ListUserADInstances(ctx context.Context, req *pb.ListUserADInstancesRequest) (*pb.ListUserADInstancesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var fromUser = false
	if userId > 0 {
		fromUser = true
		req.UserId = userId
	}

	var tx = this.NullTx()
	userInstances, err := models.SharedUserADInstanceDAO.ListUserInstances(tx, req.AdNetworkId, req.UserId, req.AdPackagePeriodId, req.ExpiresDay, req.AvailableOnly, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbUserInstances = []*pb.UserADInstance{}
	for _, userInstance := range userInstances {
		instanceIsAvailable, err := userInstance.CheckAvailable(tx)
		if err != nil {
			return nil, err
		}

		// instance
		instance, err := models.SharedADPackageInstanceDAO.FindEnabledADPackageInstance(tx, int64(userInstance.InstanceId))
		if err != nil {
			return nil, err
		}
		if instance == nil {
			continue
		}

		// package
		p, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, int64(instance.PackageId))
		if err != nil {
			return nil, err
		}

		var pbPackage *pb.ADPackage
		if p != nil {
			// network
			var pbNetwork *pb.ADNetwork
			network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(p.NetworkId))
			if err != nil {
				return nil, err
			}
			if network != nil {
				pbNetwork = &pb.ADNetwork{
					Id:          int64(network.Id),
					IsOn:        network.IsOn,
					Name:        network.Name,
					Description: network.Description,
				}
			}

			pbPackage = &pb.ADPackage{
				Id:                      int64(p.Id),
				ProtectionBandwidthSize: types.Int32(p.ProtectionBandwidthSize),
				ProtectionBandwidthUnit: p.ProtectionBandwidthUnit,
				ServerBandwidthSize:     types.Int32(p.ServerBandwidthSize),
				ServerBandwidthUnit:     p.ServerBandwidthUnit,
				Summary:                 p.Summary(network),
				AdNetwork:               pbNetwork,
				IsOn:                    p.IsOn,
			}
		}

		// 集群
		var pbCluster *pb.NodeCluster
		if !fromUser {
			cluster, err := models.SharedNodeClusterDAO.FindClusterBasicInfo(tx, int64(instance.ClusterId), nil)
			if err != nil {
				return nil, err
			}
			if cluster != nil {
				pbCluster = &pb.NodeCluster{
					Id:   int64(cluster.Id),
					Name: cluster.Name,
					IsOn: cluster.IsOn,
				}
			}
		}

		var pbInstance = &pb.ADPackageInstance{
			Id:             int64(instance.Id),
			NodeClusterId:  int64(instance.ClusterId),
			NodeIds:        instance.DecodeNodeIds(),
			IpAddresses:    instance.DecodeIPAddresses(),
			NodeCluster:    pbCluster,
			AdPackage:      pbPackage,
			UserInstanceId: int64(instance.UserInstanceId),
		}

		// user
		var pbUser *pb.User
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(userInstance.UserId))
		if err != nil {
			return nil, err
		}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
				IsOn:     user.IsOn,
			}
		}

		pbUserInstances = append(pbUserInstances, &pb.UserADInstance{
			Id:                   int64(userInstance.Id),
			UserId:               int64(userInstance.UserId),
			AdPackageInstanceId:  int64(userInstance.InstanceId),
			AdPackagePeriodId:    int64(userInstance.PeriodId),
			AdPackagePeriodCount: int32(userInstance.PeriodCount),
			AdPackagePeriodUnit:  userInstance.PeriodUnit,
			DayFrom:              userInstance.DayFrom,
			DayTo:                userInstance.DayTo,
			CreatedAt:            int64(userInstance.CreatedAt),
			MaxObjects:           types.Int32(userInstance.MaxObjects),
			ObjectCodes:          userInstance.DecodeObjectCodes(),
			AdPackageInstance:    pbInstance,
			User:                 pbUser,
			CanDelete:            userInstance.AdminId > 0,
			IsAvailable:          instanceIsAvailable,
			CountObjects:         int32(len(userInstance.DecodeObjectCodes())),
		})
	}
	return &pb.ListUserADInstancesResponse{
		UserADInstances: pbUserInstances,
	}, nil
}

// FindUserADInstance 查找单个用户高防实例
func (this *UserADInstanceService) FindUserADInstance(ctx context.Context, req *pb.FindUserADInstanceRequest) (*pb.FindUserADInstanceResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userInstance, err := models.SharedUserADInstanceDAO.FindEnabledUserADInstance(tx, req.UserADInstanceId)
	if err != nil {
		return nil, err
	}

	if userInstance == nil {
		return &pb.FindUserADInstanceResponse{
			UserADInstance: nil,
		}, nil
	}

	// 检查用户
	if userId > 0 && int64(userInstance.UserId) != userId {
		return nil, this.PermissionError()
	}

	// 是否有效
	instanceIsAvailable, err := userInstance.CheckAvailable(tx)
	if err != nil {
		return nil, err
	}

	// 防护对象
	objects, err := userInstance.DecodeObjects()
	if err != nil {
		return nil, err
	}
	if objects == nil {
		objects = []maps.Map{}
	}
	objectsJSON, err := json.Marshal(objects)
	if err != nil {
		return nil, err
	}

	// instance
	instance, err := models.SharedADPackageInstanceDAO.FindEnabledADPackageInstance(tx, int64(userInstance.InstanceId))
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return &pb.FindUserADInstanceResponse{
			UserADInstance: nil,
		}, nil
	}

	// package
	p, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, int64(instance.PackageId))
	if err != nil {
		return nil, err
	}

	var pbPackage *pb.ADPackage
	if p != nil {
		// network
		var pbNetwork *pb.ADNetwork
		network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(p.NetworkId))
		if err != nil {
			return nil, err
		}
		if network != nil {
			pbNetwork = &pb.ADNetwork{
				Id:          int64(network.Id),
				IsOn:        network.IsOn,
				Name:        network.Name,
				Description: network.Description,
			}
		}

		pbPackage = &pb.ADPackage{
			Id:                      int64(p.Id),
			ProtectionBandwidthSize: types.Int32(p.ProtectionBandwidthSize),
			ProtectionBandwidthUnit: p.ProtectionBandwidthUnit,
			ServerBandwidthSize:     types.Int32(p.ServerBandwidthSize),
			ServerBandwidthUnit:     p.ServerBandwidthUnit,
			Summary:                 p.Summary(network),
			AdNetwork:               pbNetwork,
			IsOn:                    p.IsOn,
		}
	}

	var pbInstance = &pb.ADPackageInstance{
		Id:             int64(instance.Id),
		AdPackageId:    int64(instance.PackageId),
		NodeClusterId:  int64(instance.ClusterId),
		NodeIds:        instance.DecodeNodeIds(),
		IpAddresses:    instance.DecodeIPAddresses(),
		NodeCluster:    nil,
		AdPackage:      pbPackage,
		UserInstanceId: int64(instance.UserInstanceId),
	}

	// user
	var pbUser *pb.User
	user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(userInstance.UserId))
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

	return &pb.FindUserADInstanceResponse{
		UserADInstance: &pb.UserADInstance{
			Id:                   int64(userInstance.Id),
			UserId:               int64(userInstance.UserId),
			AdPackageInstanceId:  int64(userInstance.InstanceId),
			AdPackagePeriodId:    int64(userInstance.PeriodId),
			AdPackagePeriodCount: int32(userInstance.PeriodCount),
			IsAvailable:          instanceIsAvailable,
			AdPackagePeriodUnit:  userInstance.PeriodUnit,
			DayFrom:              userInstance.DayFrom,
			DayTo:                userInstance.DayTo,
			CreatedAt:            int64(userInstance.CreatedAt),
			MaxObjects:           types.Int32(userInstance.MaxObjects),
			ObjectCodes:          userInstance.DecodeObjectCodes(),
			ObjectsJSON:          objectsJSON,
			AdPackageInstance:    pbInstance,
			User:                 pbUser,
			CanDelete:            userInstance.AdminId > 0,
		}}, nil
}

// DeleteUserADInstance 删除高防实例
func (this *UserADInstanceService) DeleteUserADInstance(ctx context.Context, req *pb.DeleteUserADInstanceRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userInstance, err := models.SharedUserADInstanceDAO.FindEnabledUserADInstance(tx, req.UserADInstanceId)
	if err != nil {
		return nil, err
	}
	if userInstance == nil {
		// 不存在，则直接成功
		return this.Success()
	}

	// 检查用户
	if userId > 0 {
		if userId != int64(userInstance.UserId) {
			return nil, this.PermissionError()
		}
	}

	var instanceId = int64(userInstance.InstanceId)

	err = this.RunTx(func(tx *dbs.Tx) error {
		err = models.SharedUserADInstanceDAO.DisableUserADInstance(tx, req.UserADInstanceId)
		if err != nil {
			return err
		}

		return models.SharedADPackageInstanceDAO.ResetInstanceUser(tx, instanceId)
	})

	return this.Success()
}

// RenewUserADInstance 续期
func (this *UserADInstanceService) RenewUserADInstance(ctx context.Context, req *pb.RenewUserADInstanceRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userInstance, err := models.SharedUserADInstanceDAO.FindEnabledUserADInstance(tx, req.UserADInstanceId)
	if err != nil {
		return nil, err
	}
	if userInstance == nil {
		return nil, errors.New("could not find user instance to renew")
	}

	// 检查用户
	if userId > 0 {
		if userId != int64(userInstance.UserId) {
			return nil, errors.New("could not find user instance to renew")
		}
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		// 查找实例信息
		var instanceId = int64(userInstance.InstanceId)
		instance, err := models.SharedADPackageInstanceDAO.FindEnabledADPackageInstance(tx, instanceId)
		if err != nil {
			return err
		}
		if instance == nil {
			return errors.New("the instance has been invalid")
		}

		// 确保操作的是同一个实例
		if instance.UserInstanceId > 0 && int64(instance.UserInstanceId) != req.UserADInstanceId {
			return errors.New("the instance has been token by other user")
		}

		var packageId = int64(instance.PackageId)
		adPackage, err := models.SharedADPackageDAO.FindEnabledADPackage(tx, packageId)
		if err != nil {
			return err
		}
		if adPackage == nil || !adPackage.IsOn {
			return errors.New("the package has been invalid")
		}

		// 线路
		network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, int64(adPackage.NetworkId))
		if err != nil {
			return err
		}
		if network == nil {
			return errors.New("the network has been invalid")
		}

		// 检查有效期
		if req.AdPackagePeriodId <= 0 {
			return errors.New("invalid 'adPackagePeriodId'")
		}
		period, err := models.SharedADPackagePeriodDAO.FindEnabledADPackagePeriod(tx, req.AdPackagePeriodId)
		if err != nil {
			return err
		}
		if period == nil {
			return errors.New("could not find period '" + types.String(req.AdPackagePeriodId) + "'")
		}

		price, err := models.SharedADPackagePriceDAO.FindPackagePrice(tx, packageId, req.AdPackagePeriodId)
		if err != nil {
			return err
		}
		if price <= 0 {
			return errors.New("can not find price for the instance")
		}

		// 如果是用户需要支付费用
		if userId > 0 {
			var amount = price

			// 先减少余额
			account, err := accounts.SharedUserAccountDAO.FindUserAccountWithUserId(tx, userId)
			if err != nil {
				return err
			}
			if account == nil || account.Total < amount {
				return errors.New("no enough balance to buy the package")
			}

			err = accounts.SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), -amount, userconfigs.AccountEventTypeRenewAntiDDoSPackage, "续费高DDoS防实例，线路："+network.Name+" / 防护带宽："+types.String(adPackage.ProtectionBandwidthSize)+userconfigs.ADPackageSizeFullUnit(adPackage.ProtectionBandwidthUnit)+" / 业务带宽："+types.String(adPackage.ServerBandwidthSize)+userconfigs.ADPackageSizeFullUnit(adPackage.ServerBandwidthUnit)+" / 高防IP："+strings.Join(instance.DecodeIPAddresses(), "，")+" / "+types.String(period.Count)+userconfigs.ADPackagePeriodUnitName(period.Unit), maps.Map{
				"adNetworkId":             network.Id,
				"adPackageId":             packageId,
				"protectionBandwidthSize": adPackage.ProtectionBandwidthSize,
				"protectionBandwidthUnit": adPackage.ProtectionBandwidthUnit,
				"serverBandwidthSize":     adPackage.ServerBandwidthSize,
				"serverBandwidthUnit":     adPackage.ServerBandwidthUnit,
				"adPackagePeriodId":       req.AdPackagePeriodId,
				"adPackagePeriodCount":    period.Count,
				"adPackagePeriodUnit":     period.Unit,
				"count":                   1,
			})
			if err != nil {
				return err
			}
		}

		_, err = models.SharedUserADInstanceDAO.RenewUserInstance(tx, userInstance, period)
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

// UpdateUserADInstanceObjects 修改实例防护对象
func (this *UserADInstanceService) UpdateUserADInstanceObjects(ctx context.Context, req *pb.UpdateUserADInstanceObjectsRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userInstance, err := models.SharedUserADInstanceDAO.FindEnabledUserADInstance(tx, req.UserADInstanceId)
	if err != nil {
		return nil, err
	}
	if userInstance == nil {
		return nil, errors.New("could not find user instance with id '" + types.String(req.UserADInstanceId) + "'")
	}
	var instanceId = int64(userInstance.InstanceId)

	// 检查用户
	if userId > 0 {
		if int64(userInstance.UserId) != userId {
			return nil, this.PermissionError()
		}
	}

	// 检查当前实例是否有效
	isAvailable, err := userInstance.CheckAvailable(tx)
	if err != nil {
		return nil, err
	}
	if !isAvailable {
		return nil, errors.New("the user instance is not available")
	}

	// TODO 检查有没有超出最大防护对象数量

	err = this.RunTx(func(tx *dbs.Tx) error {
		err = models.SharedUserADInstanceDAO.UpdateUserInstanceObjects(tx, req.UserADInstanceId, req.ObjectCodes)
		if err != nil {
			return err
		}

		return models.SharedADPackageInstanceDAO.UpdateInstanceObjects(tx, instanceId, req.ObjectCodes)
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}
