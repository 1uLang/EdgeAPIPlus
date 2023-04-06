// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus
// +build plus

package accounts

import (
	"context"
	"errors"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/userconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
)

// UserOrderService 用户订单相关服务
type UserOrderService struct {
	services.BaseService
}

// CreateUserOrder 创建订单
func (this *UserOrderService) CreateUserOrder(ctx context.Context, req *pb.CreateUserOrderRequest) (*pb.CreateUserOrderResponse, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	if !userconfigs.IsValidOrderType(req.Type) {
		return nil, errors.New("invalid order type '" + req.Type + "'")
	}

	var tx = this.NullTx()
	method, err := accounts.SharedOrderMethodDAO.FindEnabledOrderMethodWithCode(tx, req.OrderMethodCode)
	if err != nil {
		return nil, err
	}
	if method == nil {
		return nil, errors.New("can not find order method with code '" + req.OrderMethodCode + "'")
	}
	if !method.IsOn {
		return nil, errors.New("method is not enabled")
	}
	var methodId = int64(method.Id)

	if req.Amount <= 0 {
		return nil, errors.New("'amount' should be greater than 0")
	}

	var orderCode = ""
	err = this.RunTx(func(tx *dbs.Tx) error {
		_, code, err := accounts.SharedUserOrderDAO.CreateOrder(tx, 0, userId, req.Type, methodId, req.Amount)
		if err != nil {
			return err
		}
		orderCode = code
		return nil
	})
	if err != nil {
		return nil, err
	}

	order, err := accounts.SharedUserOrderDAO.FindUserOrderWithCode(tx, orderCode)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("can not find order with generated code '" + orderCode + "'")
	}
	payURL, err := order.PayURL()
	if err != nil {
		return nil, errors.New("generate pay url failed: " + err.Error())
	}

	return &pb.CreateUserOrderResponse{
		Code:   orderCode,
		PayURL: payURL,
	}, nil
}

// FindEnabledUserOrder 查看订单
func (this *UserOrderService) FindEnabledUserOrder(ctx context.Context, req *pb.FindEnabledUserOrderRequest) (*pb.FindEnabledUserOrderResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	order, err := accounts.SharedUserOrderDAO.FindUserOrderWithCode(tx, req.Code)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return &pb.FindEnabledUserOrderResponse{UserOrder: nil}, nil
	}

	// 检查用户权限
	if userId > 0 {
		if int64(order.UserId) != userId {
			return &pb.FindEnabledUserOrderResponse{UserOrder: nil}, nil
		}
	}

	// 用户
	var cacheMap = utils.NewCacheMap()
	user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(order.UserId), cacheMap)
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

	// 支付方式
	method, err := accounts.SharedOrderMethodDAO.FindEnabledOrderMethod(tx, int64(order.MethodId))
	if err != nil {
		return nil, err
	}
	var pbMethod *pb.OrderMethod
	if method != nil {
		pbMethod = &pb.OrderMethod{
			Id:   int64(method.Id),
			Name: method.Name,
			Code: method.Code,
			IsOn: method.IsOn,
		}
	}

	// 支付URL
	payURL, _ := order.PayURL()

	return &pb.FindEnabledUserOrderResponse{UserOrder: &pb.UserOrder{
		UserId:        int64(order.UserId),
		Code:          order.Code,
		Type:          order.Type,
		OrderMethodId: int64(order.MethodId),
		Status:        order.Status,
		Amount:        float32(order.Amount),
		ParamsJSON:    order.Params,
		CreatedAt:     int64(order.CreatedAt),
		CancelledAt:   int64(order.CancelledAt),
		FinishedAt:    int64(order.FinishedAt),
		IsExpired:     order.IsExpired(),
		User:          pbUser,
		OrderMethod:   pbMethod,
		CanPay:        !order.IsExpired() && order.Status == userconfigs.OrderStatusNone,
		PayURL:        payURL,
	}}, nil
}

// CancelUserOrder 取消订单
func (this *UserOrderService) CancelUserOrder(ctx context.Context, req *pb.CancelUserOrderRequest) (*pb.RPCSuccess, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	order, err := accounts.SharedUserOrderDAO.FindUserOrderWithCode(tx, req.Code)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, errors.New("can not find order")
	}

	if userId > 0 {
		if int64(order.UserId) != userId {
			return nil, errors.New("can not find order")
		}
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		return accounts.SharedUserOrderDAO.CancelOrder(tx, adminId, userId, int64(order.Id))
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FinishUserOrder 完成订单
func (this *UserOrderService) FinishUserOrder(ctx context.Context, req *pb.FinishUserOrderRequest) (*pb.RPCSuccess, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	order, err := accounts.SharedUserOrderDAO.FindUserOrderWithCode(tx, req.Code)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, errors.New("can not find order")
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		return accounts.SharedUserOrderDAO.FinishOrder(tx, adminId, 0, int64(order.Id))
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountEnabledUserOrders 计算订单数量
func (this *UserOrderService) CountEnabledUserOrders(ctx context.Context, req *pb.CountEnabledUserOrdersRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := accounts.SharedUserOrderDAO.CountEnabledUserOrders(tx, req.UserId, req.Status, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledUserOrders 列出单页订单
func (this *UserOrderService) ListEnabledUserOrders(ctx context.Context, req *pb.ListEnabledUserOrdersRequest) (*pb.ListEnabledUserOrdersResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	orders, err := accounts.SharedUserOrderDAO.ListEnabledUserOrders(tx, req.UserId, req.Status, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbOrders = []*pb.UserOrder{}
	var cacheMap = utils.NewCacheMap()
	for _, order := range orders {
		// 用户
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(order.UserId), cacheMap)
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

		// 支付方式
		method, err := accounts.SharedOrderMethodDAO.FindEnabledOrderMethod(tx, int64(order.MethodId))
		if err != nil {
			return nil, err
		}
		var pbMethod *pb.OrderMethod
		if method != nil {
			pbMethod = &pb.OrderMethod{
				Id:   int64(method.Id),
				Name: method.Name,
				Code: method.Code,
				IsOn: method.IsOn,
			}
		}

		pbOrders = append(pbOrders, &pb.UserOrder{
			UserId:        int64(order.UserId),
			Code:          order.Code,
			Type:          order.Type,
			OrderMethodId: int64(order.MethodId),
			Status:        order.Status,
			Amount:        float32(order.Amount),
			ParamsJSON:    order.Params,
			CreatedAt:     int64(order.CreatedAt),
			CancelledAt:   int64(order.CancelledAt),
			FinishedAt:    int64(order.FinishedAt),
			IsExpired:     order.IsExpired(),
			User:          pbUser,
			OrderMethod:   pbMethod,
		})
	}
	return &pb.ListEnabledUserOrdersResponse{
		UserOrders: pbOrders,
	}, nil
}
