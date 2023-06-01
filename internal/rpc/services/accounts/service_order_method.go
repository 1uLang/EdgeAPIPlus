// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package accounts

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
)

// OrderMethodService 订单支付方式相关服务
type OrderMethodService struct {
	services.BaseService
}

// CreateOrderMethod 创建支付方式
func (this *OrderMethodService) CreateOrderMethod(ctx context.Context, req *pb.CreateOrderMethodRequest) (*pb.CreateOrderMethodResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查代号是否相同
	exists, err := accounts.SharedOrderMethodDAO.ExistOrderMethodWithCode(tx, req.Code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("pay method code already exists")
	}

	var params any
	if len(req.ParentCode) > 0 {
		params, err = userconfigs.DecodePayMethodParams(req.ParentCode, req.ParamsJSON)
		if err != nil {
			return nil, errors.New("invalid params: " + err.Error())
		}
	}

	methodId, err := accounts.SharedOrderMethodDAO.CreateMethod(tx, req.Name, req.Code, req.Url, req.Description, req.ParentCode, params, req.ClientType, req.QrcodeTitle)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderMethodResponse{
		OrderMethodId: methodId,
	}, nil
}

// UpdateOrderMethod 修改支付方式
func (this *OrderMethodService) UpdateOrderMethod(ctx context.Context, req *pb.UpdateOrderMethodRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查代号是否相同
	exists, err := accounts.SharedOrderMethodDAO.ExistOrderMethodWithCode(tx, req.Code, req.OrderMethodId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("pay method code already exists")
	}

	method, err := accounts.SharedOrderMethodDAO.FindEnabledBasicOrderMethod(tx, req.OrderMethodId)
	if err != nil {
		return nil, err
	}
	if method == nil {
		return nil, errors.New("could not find method")
	}

	var params any
	if len(method.ParentCode) > 0 {
		params, err = userconfigs.DecodePayMethodParams(method.ParentCode, req.ParamsJSON)
		if err != nil {
			return nil, errors.New("invalid params: " + err.Error())
		}
	}

	err = accounts.SharedOrderMethodDAO.UpdateMethod(tx, req.OrderMethodId, req.Name, req.Code, req.Url, req.Description, params, req.ClientType, req.QrcodeTitle, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteOrderMethod 删除支付方式
func (this *OrderMethodService) DeleteOrderMethod(ctx context.Context, req *pb.DeleteOrderMethodRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = accounts.SharedOrderMethodDAO.DisableOrderMethod(tx, req.OrderMethodId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledOrderMethod 查找单个支付方式
func (this *OrderMethodService) FindEnabledOrderMethod(ctx context.Context, req *pb.FindEnabledOrderMethodRequest) (*pb.FindEnabledOrderMethodResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	method, err := accounts.SharedOrderMethodDAO.FindEnabledOrderMethod(tx, req.OrderMethodId)
	if err != nil {
		return nil, err
	}
	if method == nil {
		return &pb.FindEnabledOrderMethodResponse{
			OrderMethod: nil,
		}, nil
	}

	return &pb.FindEnabledOrderMethodResponse{
		OrderMethod: &pb.OrderMethod{
			Id:          int64(method.Id),
			Name:        method.Name,
			Code:        method.Code,
			Description: method.Description,
			Url:         method.Url,
			Secret:      method.Secret,
			IsOn:        method.IsOn,
			ParentCode:  method.ParentCode,
			Params:      method.Params, // 注意参数不能通过接口泄露给平台用户
			ClientType:  method.ClientType,
			QrcodeTitle: method.QrcodeTitle,
		},
	}, nil
}

// FindEnabledOrderMethodWithCode 根据代号查找支付方式
func (this *OrderMethodService) FindEnabledOrderMethodWithCode(ctx context.Context, req *pb.FindEnabledOrderMethodWithCodeRequest) (*pb.FindEnabledOrderMethodWithCodeResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	method, err := accounts.SharedOrderMethodDAO.FindEnabledOrderMethodWithCode(tx, req.Code)
	if err != nil {
		return nil, err
	}
	if method == nil {
		return &pb.FindEnabledOrderMethodWithCodeResponse{
			OrderMethod: nil,
		}, nil
	}

	// 保护数据
	if userId > 0 {
		method.Secret = ""
	}

	return &pb.FindEnabledOrderMethodWithCodeResponse{
		OrderMethod: &pb.OrderMethod{
			Id:          int64(method.Id),
			Name:        method.Name,
			Code:        method.Code,
			ParentCode:  method.ParentCode,
			Description: method.Description,
			Url:         method.Url,
			Secret:      method.Secret,
			IsOn:        method.IsOn,
			ClientType:  method.ClientType,
			QrcodeTitle: method.QrcodeTitle,
		},
	}, nil
}

// FindAllEnabledOrderMethods 查找所有支付方式
func (this *OrderMethodService) FindAllEnabledOrderMethods(ctx context.Context, req *pb.FindAllEnabledOrderMethodsRequest) (*pb.FindAllEnabledOrderMethodsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	methods, err := accounts.SharedOrderMethodDAO.FindAllEnabledMethodOrders(tx)
	if err != nil {
		return nil, err
	}
	var pbMethods = []*pb.OrderMethod{}
	for _, method := range methods {
		// 防止secret泄露
		if userId > 0 {
			method.Secret = ""
		}

		pbMethods = append(pbMethods, &pb.OrderMethod{
			Id:          int64(method.Id),
			Name:        method.Name,
			Code:        method.Code,
			Description: method.Description,
			Url:         method.Url,
			Secret:      method.Secret,
			IsOn:        method.IsOn,
			ParentCode:  method.ParentCode, // 不要返回params，以防止泄露
			ClientType:  method.ClientType,
		})
	}
	return &pb.FindAllEnabledOrderMethodsResponse{
		OrderMethods: pbMethods,
	}, nil
}

// FindAllAvailableOrderMethods 查找所有已启用的支付方式
func (this *OrderMethodService) FindAllAvailableOrderMethods(ctx context.Context, req *pb.FindAllAvailableOrderMethodsRequest) (*pb.FindAllAvailableOrderMethodsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	methods, err := accounts.SharedOrderMethodDAO.FindAllEnabledAndOnMethodOrders(tx)
	if err != nil {
		return nil, err
	}
	var pbMethods = []*pb.OrderMethod{}
	for _, method := range methods {
		// 防止secret泄露
		if userId > 0 {
			method.Secret = ""
		}

		pbMethods = append(pbMethods, &pb.OrderMethod{
			Id:          int64(method.Id),
			Name:        method.Name,
			Code:        method.Code,
			Description: method.Description,
			Url:         method.Url,
			Secret:      method.Secret,
			IsOn:        method.IsOn,
			ParentCode:  method.ParentCode, // 不返回params，防止泄露
			ClientType:  method.ClientType,
		})
	}
	return &pb.FindAllAvailableOrderMethodsResponse{
		OrderMethods: pbMethods,
	}, nil
}
