// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package tickets

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/tickets"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UserTicketLogService 工单日志服务
type UserTicketLogService struct {
	services.BaseService
}

// CreateUserTicketLog 创建日志
func (this *UserTicketLogService) CreateUserTicketLog(ctx context.Context, req *pb.CreateUserTicketLogRequest) (*pb.CreateUserTicketLogResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 如果没有指定状态，就使用上次的状态
	if len(req.Status) == 0 {
		status, err := tickets.SharedUserTicketDAO.FindTicketStatus(tx, req.UserTicketId)
		if err != nil {
			return nil, err
		}
		req.Status = status
	}

	logId, err := tickets.SharedUserTicketLogDAO.CreateLog(tx, adminId, userId, req.UserTicketId, req.Status, req.Comment, false)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserTicketLogResponse{
		UserTicketLogId: logId,
	}, nil
}

// DeleteUserTicketLog 删除日志
func (this *UserTicketLogService) DeleteUserTicketLog(ctx context.Context, req *pb.DeleteUserTicketLogRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = tickets.SharedUserTicketLogDAO.CheckUserLog(tx, userId, req.UserTicketLogId)
		if err != nil {
			return nil, err
		}
	}

	isReadonly, err := tickets.SharedUserTicketLogDAO.CheckLogReadonly(tx, req.UserTicketLogId)
	if err != nil {
		return nil, err
	}
	if isReadonly {
		return nil, errors.New("the log is readonly")
	}

	err = tickets.SharedUserTicketLogDAO.DisableLog(tx, req.UserTicketLogId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountUserTicketLogs 查询日志数量
func (this *UserTicketLogService) CountUserTicketLogs(ctx context.Context, req *pb.CountUserTicketLogsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = tickets.SharedUserTicketDAO.CheckUserTicket(tx, userId, req.UserTicketId)
		if err != nil {
			return nil, err
		}
	}

	count, err := tickets.SharedUserTicketLogDAO.CountTicketLogs(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListUserTicketLogs 列出单页日志
func (this *UserTicketLogService) ListUserTicketLogs(ctx context.Context, req *pb.ListUserTicketLogsRequest) (*pb.ListUserTicketLogsResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = tickets.SharedUserTicketDAO.CheckUserTicket(tx, userId, req.UserTicketId)
		if err != nil {
			return nil, err
		}
	}

	logs, err := tickets.SharedUserTicketLogDAO.ListTicketLogs(tx, req.UserTicketId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbLogs = []*pb.UserTicketLog{}
	for _, log := range logs {
		// 只有管理员才可以看到管理员信息
		var pbAdmin *pb.Admin
		var pbUser *pb.User
		if adminId > 0 {
			if log.AdminId > 0 {
				admin, err := models.SharedAdminDAO.FindBasicAdmin(tx, int64(log.AdminId))
				if err != nil {
					return nil, err
				}
				if admin != nil {
					pbAdmin = &pb.Admin{
						Id:       int64(admin.Id),
						Fullname: admin.Fullname,
						Username: admin.Username,
					}
				}
			}

			if log.UserId > 0 {
				user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(log.UserId))
				if err != nil {
					return nil, err
				}
				if user != nil {
					pbUser = &pb.User{
						Id:       int64(user.Id),
						Fullname: user.Fullname,
						Username: user.Username,
					}
				}
			}
		}

		pbLogs = append(pbLogs, &pb.UserTicketLog{
			Id:         int64(log.Id),
			AdminId:    int64(log.AdminId),
			UserId:     int64(log.UserId),
			TicketId:   int64(log.TicketId),
			Status:     log.Status,
			Comment:    log.Comment,
			CreatedAt:  int64(log.CreatedAt),
			IsReadonly: log.IsReadonly,
			Admin:      pbAdmin,
			User:       pbUser,
		})
	}

	return &pb.ListUserTicketLogsResponse{
		UserTicketLogs: pbLogs,
	}, nil
}
