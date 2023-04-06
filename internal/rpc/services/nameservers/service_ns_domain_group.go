// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package nameservers

import (
	"context"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
)

// NSDomainGroupService 域名分组服务
type NSDomainGroupService struct {
	services.BaseService
}

// CreateNSDomainGroup 创建分组
func (this *NSDomainGroupService) CreateNSDomainGroup(ctx context.Context, req *pb.CreateNSDomainGroupRequest) (*pb.CreateNSDomainGroupResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	groupId, err := nameservers.SharedNSDomainGroupDAO.CreateGroup(tx, userId, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNSDomainGroupResponse{
		NsDomainGroupId: groupId,
	}, nil
}

// UpdateNSDomainGroup 修改分组
func (this *NSDomainGroupService) UpdateNSDomainGroup(ctx context.Context, req *pb.UpdateNSDomainGroupRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSDomainGroupDAO.CheckUserGroup(tx, userId, req.NsDomainGroupId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSDomainGroupDAO.UpdateGroup(tx, req.NsDomainGroupId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSDomainGroup 删除分组
func (this *NSDomainGroupService) DeleteNSDomainGroup(ctx context.Context, req *pb.DeleteNSDomainGroupRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSDomainGroupDAO.CheckUserGroup(tx, userId, req.NsDomainGroupId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSDomainGroupDAO.DisableNSDomainGroup(tx, req.NsDomainGroupId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllNSDomainGroups 查询所有分组
func (this *NSDomainGroupService) FindAllNSDomainGroups(ctx context.Context, req *pb.FindAllNSDomainGroupsRequest) (*pb.FindAllNSDomainGroupsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId
	}

	groups, err := nameservers.SharedNSDomainGroupDAO.FindAllGroups(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	var pbGroups = []*pb.NSDomainGroup{}
	for _, group := range groups {
		pbGroups = append(pbGroups, &pb.NSDomainGroup{
			Id:     int64(group.Id),
			Name:   group.Name,
			IsOn:   group.IsOn,
			UserId: int64(group.UserId),
		})
	}
	return &pb.FindAllNSDomainGroupsResponse{
		NsDomainGroups: pbGroups,
	}, nil
}

// CountAllAvailableNSDomainGroups 查询可用分组数量
func (this *NSDomainGroupService) CountAllAvailableNSDomainGroups(ctx context.Context, req *pb.CountAllAvailableNSDomainGroupsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := nameservers.SharedNSDomainGroupDAO.CountAllAvailableGroups(tx, req.UserId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// FindAllAvailableNSDomainGroups 查询所有分组
func (this *NSDomainGroupService) FindAllAvailableNSDomainGroups(ctx context.Context, req *pb.FindAllAvailableNSDomainGroupsRequest) (*pb.FindAllAvailableNSDomainGroupsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	groups, err := nameservers.SharedNSDomainGroupDAO.FindAllAvailableGroups(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	var pbGroups = []*pb.NSDomainGroup{}
	for _, group := range groups {
		pbGroups = append(pbGroups, &pb.NSDomainGroup{
			Id:     int64(group.Id),
			Name:   group.Name,
			IsOn:   group.IsOn,
			UserId: int64(group.UserId),
		})
	}
	return &pb.FindAllAvailableNSDomainGroupsResponse{
		NsDomainGroups: pbGroups,
	}, nil
}

// FindNSDomainGroup 查找单个分组
func (this *NSDomainGroupService) FindNSDomainGroup(ctx context.Context, req *pb.FindNSDomainGroupRequest) (*pb.FindNSDomainGroupResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	group, err := nameservers.SharedNSDomainGroupDAO.FindEnabledNSDomainGroup(tx, req.NsDomainGroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindNSDomainGroupResponse{
			NsDomainGroup: nil,
		}, nil
	}

	if int64(group.UserId) != userId {
		return &pb.FindNSDomainGroupResponse{
			NsDomainGroup: nil,
		}, nil
	}

	return &pb.FindNSDomainGroupResponse{
		NsDomainGroup: &pb.NSDomainGroup{
			Id:     int64(group.Id),
			Name:   group.Name,
			IsOn:   group.IsOn,
			UserId: int64(group.UserId),
		},
	}, nil
}
