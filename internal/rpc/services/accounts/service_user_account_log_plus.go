// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package accounts

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UserAccountLogService 用户账户日志服务
type UserAccountLogService struct {
	services.BaseService
}

// CountUserAccountLogs 计算日志数量
func (this *UserAccountLogService) CountUserAccountLogs(ctx context.Context, req *pb.CountUserAccountLogsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := accounts.SharedUserAccountLogDAO.CountAccountLogs(tx, 0, req.UserAccountId, req.Keyword, req.EventType)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserAccountLogs 列出单页日志
func (this *UserAccountLogService) ListUserAccountLogs(ctx context.Context, req *pb.ListUserAccountLogsRequest) (*pb.ListUserAccountLogsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	logs, err := accounts.SharedUserAccountLogDAO.ListAccountLogs(tx, 0, req.UserAccountId, req.Keyword, req.EventType, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbLogs = []*pb.UserAccountLog{}
	var cacheMap = utils.NewCacheMap()
	for _, log := range logs {
		// 用户
		var pbUser = &pb.User{Id: int64(log.UserId)}
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(log.UserId), cacheMap)
		if err != nil {
			return nil, err
		}
		if user != nil {
			pbUser = &pb.User{Id: int64(user.Id), Fullname: user.Fullname, Username: user.Username}
		}

		// 账户
		var pbAccount = &pb.UserAccount{Id: int64(log.AccountId)}

		pbLogs = append(pbLogs, &pb.UserAccountLog{
			Id:            int64(log.Id),
			UserId:        int64(log.UserId),
			UserAccountId: int64(log.AccountId),
			Delta:         float32(log.Delta),
			DeltaFrozen:   float32(log.DeltaFrozen),
			Total:         float32(log.Total),
			TotalFrozen:   float32(log.TotalFrozen),
			EventType:     log.EventType,
			Description:   log.Description,
			CreatedAt:     int64(log.CreatedAt),
			ParamsJSON:    []byte(log.Params),
			User:          pbUser,
			UserAccount:   pbAccount,
		})
	}
	return &pb.ListUserAccountLogsResponse{UserAccountLogs: pbLogs}, nil
}
