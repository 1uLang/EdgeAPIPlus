// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package accounts

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

// UserAccountService 用户账户相关服务
type UserAccountService struct {
	services.BaseService
}

// CountUserAccounts 计算账户数量
func (this *UserAccountService) CountUserAccounts(ctx context.Context, req *pb.CountUserAccountsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := accounts.SharedUserAccountDAO.CountAllAccounts(tx, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserAccounts 列出单页账户
func (this *UserAccountService) ListUserAccounts(ctx context.Context, req *pb.ListUserAccountsRequest) (*pb.ListUserAccountsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userAccounts, err := accounts.SharedUserAccountDAO.ListAccounts(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbAccounts = []*pb.UserAccount{}
	for _, account := range userAccounts {
		// 用户
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(account.UserId), nil)
		if err != nil {
			return nil, err
		}
		var pbUser = &pb.User{}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}

		pbAccounts = append(pbAccounts, &pb.UserAccount{
			Id:          int64(account.Id),
			UserId:      int64(account.UserId),
			Total:       float32(account.Total),
			TotalFrozen: float32(account.TotalFrozen),
			User:        pbUser,
		})
	}
	return &pb.ListUserAccountsResponse{UserAccounts: pbAccounts}, nil
}

// FindEnabledUserAccountWithUserId 根据用户ID查找单个账户
func (this *UserAccountService) FindEnabledUserAccountWithUserId(ctx context.Context, req *pb.FindEnabledUserAccountWithUserIdRequest) (*pb.FindEnabledUserAccountWithUserIdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	account, err := accounts.SharedUserAccountDAO.FindUserAccountWithUserId(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return &pb.FindEnabledUserAccountWithUserIdResponse{UserAccount: nil}, nil
	}

	// 用户
	user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(account.UserId), nil)
	if err != nil {
		return nil, err
	}
	var pbUser = &pb.User{}
	if user != nil {
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	return &pb.FindEnabledUserAccountWithUserIdResponse{
		UserAccount: &pb.UserAccount{
			Id:          int64(account.Id),
			UserId:      int64(account.UserId),
			Total:       float32(account.Total),
			TotalFrozen: float32(account.TotalFrozen),
			User:        pbUser,
		},
	}, nil
}

// FindEnabledUserAccount 查找单个账户
func (this *UserAccountService) FindEnabledUserAccount(ctx context.Context, req *pb.FindEnabledUserAccountRequest) (*pb.FindEnabledUserAccountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	account, err := accounts.SharedUserAccountDAO.FindUserAccountWithAccountId(tx, req.UserAccountId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return &pb.FindEnabledUserAccountResponse{UserAccount: nil}, nil
	}

	// 用户
	user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(account.UserId), nil)
	if err != nil {
		return nil, err
	}
	var pbUser = &pb.User{}
	if user != nil {
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	return &pb.FindEnabledUserAccountResponse{
		UserAccount: &pb.UserAccount{
			Id:          int64(account.Id),
			UserId:      int64(account.UserId),
			Total:       float32(account.Total),
			TotalFrozen: float32(account.TotalFrozen),
			User:        pbUser,
		},
	}, nil
}

// UpdateUserAccount 修改用户账户
func (this *UserAccountService) UpdateUserAccount(ctx context.Context, req *pb.UpdateUserAccountRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var params = maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}
	err = this.RunTx(func(tx *dbs.Tx) error {
		err := accounts.SharedUserAccountDAO.UpdateUserAccount(tx, req.UserAccountId, req.Delta, req.EventType, req.Description, params)
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
