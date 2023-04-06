// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package tickets

import (
	"context"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/tickets"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
)

// UserTicketService 工单服务
type UserTicketService struct {
	services.BaseService
}

// CreateUserTicket 创建工单
func (this *UserTicketService) CreateUserTicket(ctx context.Context, req *pb.CreateUserTicketRequest) (*pb.CreateUserTicketResponse, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	// 检查工单分类是否存在
	if req.UserTicketCategoryId > 0 {
		// TODO 检查工单分类是否存在
	}

	var tx = this.NullTx()
	ticketId, err := tickets.SharedUserTicketDAO.CreateTicket(tx, userId, req.UserTicketCategoryId, req.Subject, req.Body)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserTicketResponse{
		UserTicketId: ticketId,
	}, nil
}

// UpdateUserTicket 修改工单
func (this *UserTicketService) UpdateUserTicket(ctx context.Context, req *pb.UpdateUserTicketRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = tickets.SharedUserTicketDAO.CheckUserTicket(tx, userId, req.UserTicketId)
	if err != nil {
		return nil, err
	}

	err = tickets.SharedUserTicketDAO.UpdateTicket(tx, req.UserTicketId, req.UserTicketCategoryId, req.Subject, req.Body)
	if err != nil {
		return nil, err
	}

	// 创建日志
	ticketStatus, err := tickets.SharedUserTicketDAO.FindTicketStatus(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}

	_, err = tickets.SharedUserTicketLogDAO.CreateLog(tx, 0, userId, req.UserTicketId, ticketStatus, "修改工单内容", true /** 系统记录，不允许删除 **/)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteUserTicket 删除工单
func (this *UserTicketService) DeleteUserTicket(ctx context.Context, req *pb.DeleteUserTicketRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = tickets.SharedUserTicketDAO.CheckUserTicket(tx, userId, req.UserTicketId)
	if err != nil {
		return nil, err
	}

	ticketStatus, err := tickets.SharedUserTicketDAO.FindTicketStatus(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}

	err = tickets.SharedUserTicketDAO.DisableUserTicket(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}

	// 创建日志
	_, err = tickets.SharedUserTicketLogDAO.CreateLog(tx, 0, userId, req.UserTicketId, ticketStatus, "删除工单", true /** 系统记录，不允许删除 **/)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountUserTickets 计算工单数量
func (this *UserTicketService) CountUserTickets(ctx context.Context, req *pb.CountUserTicketsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	count, err := tickets.SharedUserTicketDAO.CountAllTickets(tx, req.UserId, req.UserTicketCategoryId, req.Status)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserTickets 列出单页工单
func (this *UserTicketService) ListUserTickets(ctx context.Context, req *pb.ListUserTicketsRequest) (*pb.ListUserTicketsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	ticketList, err := tickets.SharedUserTicketDAO.ListTickets(tx, req.UserId, req.UserTicketCategoryId, req.Status, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbTickets = []*pb.UserTicket{}
	for _, ticket := range ticketList {
		// 分类
		category, err := tickets.SharedUserTicketCategoryDAO.FindEnabledUserTicketCategory(tx, int64(ticket.CategoryId))
		if err != nil {
			return nil, err
		}
		var pbCategory *pb.UserTicketCategory
		if category != nil {
			pbCategory = &pb.UserTicketCategory{
				Id:   int64(category.Id),
				Name: category.Name,
				IsOn: category.IsOn,
			}
		}

		// 用户
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(ticket.UserId))
		if err != nil {
			return nil, err
		}
		var pbUser *pb.User
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Fullname: user.Fullname,
				Username: user.Username,
			}
		}

		// 最新一条日志
		latestLog, err := tickets.SharedUserTicketLogDAO.FindLatestTicketLog(tx, int64(ticket.Id))
		if err != nil {
			return nil, err
		}

		var pbLatestLog *pb.UserTicketLog
		if latestLog != nil {
			var pbLogAdmin *pb.Admin
			if latestLog.AdminId > 0 {
				logAdmin, err := models.SharedAdminDAO.FindBasicAdmin(tx, int64(latestLog.AdminId))
				if err != nil {
					return nil, err
				}
				if logAdmin != nil {
					pbLogAdmin = &pb.Admin{
						Id:       int64(logAdmin.Id),
						Username: logAdmin.Username,
						Fullname: logAdmin.Fullname,
					}
				}
			}

			pbLatestLog = &pb.UserTicketLog{
				Id:        int64(latestLog.Id),
				CreatedAt: int64(latestLog.CreatedAt),
				Admin:     pbLogAdmin,
			}
		}

		pbTickets = append(pbTickets, &pb.UserTicket{
			Id:                  int64(ticket.Id),
			CategoryId:          int64(ticket.CategoryId),
			UserId:              int64(ticket.UserId),
			Subject:             ticket.Subject,
			Body:                ticket.Body,
			Status:              ticket.Status,
			CreatedAt:           int64(ticket.CreatedAt),
			LastLogAt:           int64(ticket.LastLogAt),
			UserTicketCategory:  pbCategory,
			User:                pbUser,
			LatestUserTicketLog: pbLatestLog,
		})
	}

	return &pb.ListUserTicketsResponse{
		UserTickets: pbTickets,
	}, nil
}

// FindUserTicket 查找单个工单
func (this *UserTicketService) FindUserTicket(ctx context.Context, req *pb.FindUserTicketRequest) (*pb.FindUserTicketResponse, error) {
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

	ticket, err := tickets.SharedUserTicketDAO.FindEnabledUserTicket(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return &pb.FindUserTicketResponse{
			UserTicket: nil,
		}, nil
	}

	// 分类
	category, err := tickets.SharedUserTicketCategoryDAO.FindEnabledUserTicketCategory(tx, int64(ticket.CategoryId))
	if err != nil {
		return nil, err
	}
	var pbCategory *pb.UserTicketCategory
	if category != nil {
		pbCategory = &pb.UserTicketCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn,
		}
	}

	// 用户
	user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(ticket.UserId))
	if err != nil {
		return nil, err
	}
	var pbUser *pb.User
	if user != nil {
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Fullname: user.Fullname,
			Username: user.Username,
		}
	}

	return &pb.FindUserTicketResponse{
		UserTicket: &pb.UserTicket{
			Id:                 int64(ticket.Id),
			CategoryId:         int64(ticket.CategoryId),
			UserId:             int64(ticket.UserId),
			Subject:            ticket.Subject,
			Body:               ticket.Body,
			Status:             ticket.Status,
			CreatedAt:          int64(ticket.CreatedAt),
			LastLogAt:          int64(ticket.LastLogAt),
			UserTicketCategory: pbCategory,
			User:               pbUser,
		},
	}, nil
}
