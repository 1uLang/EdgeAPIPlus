// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NSRouteCategoryService 线路分类服务
type NSRouteCategoryService struct {
	services.BaseService
}

// CreateNSRouteCategory 创建线路分类
func (this *NSRouteCategoryService) CreateNSRouteCategory(ctx context.Context, req *pb.CreateNSRouteCategoryRequest) (*pb.CreateNSRouteCategoryResponse, error) {
	// TODO 需要防止用户恶意创建非常多的分类
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categoryId, err := nameservers.SharedNSRouteCategoryDAO.CreateCategory(tx, adminId, userId, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNSRouteCategoryResponse{NsRouteCategoryId: categoryId}, nil
}

// UpdateNSRouteCategory 修改线路分类
func (this *NSRouteCategoryService) UpdateNSRouteCategory(ctx context.Context, req *pb.UpdateNSRouteCategoryRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSRouteCategoryDAO.CheckUserCategory(tx, userId, req.NsRouteCategoryId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSRouteCategoryDAO.UpdateCategory(tx, req.NsRouteCategoryId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSRouteCategory 删除线路分类
func (this *NSRouteCategoryService) DeleteNSRouteCategory(ctx context.Context, req *pb.DeleteNSRouteCategoryRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSRouteCategoryDAO.CheckUserCategory(tx, userId, req.NsRouteCategoryId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSRouteCategoryDAO.DisableNSRouteCategory(tx, req.NsRouteCategoryId)
	if err != nil {
		return nil, err
	}

	// 重置线路
	err = nameservers.SharedNSRouteDAO.ResetRoutesCategory(tx, req.NsRouteCategoryId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllNSRouteCategories 列出所有线路分类
func (this *NSRouteCategoryService) FindAllNSRouteCategories(ctx context.Context, req *pb.FindAllNSRouteCategoriesRequest) (*pb.FindAllNSRouteCategoriesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categories, err := nameservers.SharedNSRouteCategoryDAO.FindAllCategories(tx, userId)
	if err != nil {
		return nil, err
	}

	var pbCategories = []*pb.NSRouteCategory{}
	for _, category := range categories {
		pbCategories = append(pbCategories, &pb.NSRouteCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn,
		})
	}

	return &pb.FindAllNSRouteCategoriesResponse{
		NsRouteCategories: pbCategories,
	}, nil
}

// UpdateNSRouteCategoryOrders 对线路分类进行排序
func (this *NSRouteCategoryService) UpdateNSRouteCategoryOrders(ctx context.Context, req *pb.UpdateNSRouteCategoryOrders) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		for _, categoryId := range req.NsRouteCategoryIds {
			err = nameservers.SharedNSRouteCategoryDAO.CheckUserCategory(tx, userId, categoryId)
			if err != nil {
				return nil, err
			}
		}
	}

	err = nameservers.SharedNSRouteCategoryDAO.UpdateCategoryOrders(tx, userId, req.NsRouteCategoryIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNSRouteCategory 查找单个线路分类
func (this *NSRouteCategoryService) FindNSRouteCategory(ctx context.Context, req *pb.FindNSRouteCategoryRequest) (*pb.FindNSRouteCategoryResponse, error) {

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSRouteCategoryDAO.CheckUserCategory(tx, userId, req.NsRouteCategoryId)
		if err != nil {
			return nil, err
		}
	}

	category, err := nameservers.SharedNSRouteCategoryDAO.FindCategory(tx, req.NsRouteCategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return &pb.FindNSRouteCategoryResponse{
			NsRouteCategory: nil,
		}, nil
	}

	return &pb.FindNSRouteCategoryResponse{
		NsRouteCategory: &pb.NSRouteCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn,
		},
	}, nil
}
