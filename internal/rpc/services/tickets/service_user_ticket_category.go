// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package tickets

import (
	"context"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/tickets"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
)

// UserTicketCategoryService 工单分类服务
type UserTicketCategoryService struct {
	services.BaseService
}

// CreateUserTicketCategory 创建分类
func (this *UserTicketCategoryService) CreateUserTicketCategory(ctx context.Context, req *pb.CreateUserTicketCategoryRequest) (*pb.CreateUserTicketCategoryResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categoryId, err := tickets.SharedUserTicketCategoryDAO.CreateCategory(tx, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserTicketCategoryResponse{
		UserTicketCategoryId: categoryId,
	}, nil
}

// UpdateUserTicketCategory 修改分类
func (this *UserTicketCategoryService) UpdateUserTicketCategory(ctx context.Context, req *pb.UpdateUserTicketCategoryRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = tickets.SharedUserTicketCategoryDAO.UpdateCategory(tx, req.UserTicketCategoryId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteUserTicketCategory 删除分类
func (this *UserTicketCategoryService) DeleteUserTicketCategory(ctx context.Context, req *pb.DeleteUserTicketCategoryRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = tickets.SharedUserTicketCategoryDAO.DisableUserTicketCategory(tx, req.UserTicketCategoryId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllUserTicketCategories 查找所有分类
func (this *UserTicketCategoryService) FindAllUserTicketCategories(ctx context.Context, req *pb.FindAllUserTicketCategoriesRequest) (*pb.FindAllUserTicketCategoriesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categories, err := tickets.SharedUserTicketCategoryDAO.FindAllEnabledCategories(tx)
	if err != nil {
		return nil, err
	}
	var pbCategories = []*pb.UserTicketCategory{}
	for _, category := range categories {
		pbCategories = append(pbCategories, &pb.UserTicketCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn,
		})
	}
	return &pb.FindAllUserTicketCategoriesResponse{
		UserTicketCategories: pbCategories,
	}, nil
}

// FindAllAvailableUserTicketCategories 查找所有启用中的分类
func (this *UserTicketCategoryService) FindAllAvailableUserTicketCategories(ctx context.Context, req *pb.FindAllAvailableUserTicketCategoriesRequest) (*pb.FindAllAvailableUserTicketCategoriesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categories, err := tickets.SharedUserTicketCategoryDAO.FindAllEnabledAndOnCategories(tx)
	if err != nil {
		return nil, err
	}
	var pbCategories = []*pb.UserTicketCategory{}
	for _, category := range categories {
		pbCategories = append(pbCategories, &pb.UserTicketCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn,
		})
	}
	return &pb.FindAllAvailableUserTicketCategoriesResponse{
		UserTicketCategories: pbCategories,
	}, nil
}

// FindUserTicketCategory 查询单个分类
func (this *UserTicketCategoryService) FindUserTicketCategory(ctx context.Context, req *pb.FindUserTicketCategoryRequest) (*pb.FindUserTicketCategoryResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	category, err := tickets.SharedUserTicketCategoryDAO.FindEnabledUserTicketCategory(tx, req.UserTicketCategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return &pb.FindUserTicketCategoryResponse{
			UserTicketCategory: nil,
		}, nil
	}

	return &pb.FindUserTicketCategoryResponse{
		UserTicketCategory: &pb.UserTicketCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn,
		},
	}, nil
}
