// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package reporters

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ReportNodeGroupService 监控节点分组
type ReportNodeGroupService struct {
	services.BaseService
}

// CreateReportNodeGroup 创建分组
func (this *ReportNodeGroupService) CreateReportNodeGroup(ctx context.Context, req *pb.CreateReportNodeGroupRequest) (*pb.CreateReportNodeGroupResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	groupId, err := models.SharedReportNodeGroupDAO.CreateGroup(tx, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateReportNodeGroupResponse{ReportNodeGroupId: groupId}, nil
}

// UpdateReportNodeGroup 修改分组
func (this *ReportNodeGroupService) UpdateReportNodeGroup(ctx context.Context, req *pb.UpdateReportNodeGroupRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedReportNodeGroupDAO.UpdateGroup(tx, req.ReportNodeGroupId, req.Name)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteReportNodeGroup 删除分组
func (this *ReportNodeGroupService) DeleteReportNodeGroup(ctx context.Context, req *pb.DeleteReportNodeGroupRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedReportNodeGroupDAO.DisableReportNodeGroup(tx, req.ReportNodeGroupId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllEnabledReportNodeGroups 查找所有分组
func (this *ReportNodeGroupService) FindAllEnabledReportNodeGroups(ctx context.Context, req *pb.FindAllEnabledReportNodeGroupsRequest) (*pb.FindAllEnabledReportNodeGroupsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	groups, err := models.SharedReportNodeGroupDAO.FindAllEnabledGroups(tx)
	if err != nil {
		return nil, err
	}
	var pbGroups = []*pb.ReportNodeGroup{}
	for _, group := range groups {
		pbGroups = append(pbGroups, &pb.ReportNodeGroup{
			Id:   int64(group.Id),
			Name: group.Name,
			IsOn: group.IsOn == 1,
		})
	}
	return &pb.FindAllEnabledReportNodeGroupsResponse{ReportNodeGroups: pbGroups}, nil
}

// FindEnabledReportNodeGroup 查找单个分组
func (this *ReportNodeGroupService) FindEnabledReportNodeGroup(ctx context.Context, req *pb.FindEnabledReportNodeGroupRequest) (*pb.FindEnabledReportNodeGroupResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	group, err := models.SharedReportNodeGroupDAO.FindEnabledReportNodeGroup(tx, req.ReportNodeGroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledReportNodeGroupResponse{ReportNodeGroup: nil}, nil
	}

	return &pb.FindEnabledReportNodeGroupResponse{
		ReportNodeGroup: &pb.ReportNodeGroup{
			Id:   int64(group.Id),
			Name: group.Name,
			IsOn: group.IsOn == 1,
		},
	}, nil
}

// CountAllEnabledReportNodeGroups 计算所有分组数量
func (this *ReportNodeGroupService) CountAllEnabledReportNodeGroups(ctx context.Context, req *pb.CountAllEnabledReportNodeGroupsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedReportNodeGroupDAO.CountAllEnabledGroups(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}
