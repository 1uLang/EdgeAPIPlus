// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package services

import (
	"context"
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/dbs"
)

// UpdateServerUAM 修改服务UAM设置
func (this *ServerService) UpdateServerUAM(ctx context.Context, req *pb.UpdateServerUAMRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	var config = &serverconfigs.UAMConfig{}
	err = json.Unmarshal(req.UamJSON, config)
	if err != nil {
		return nil, err
	}

	err = models.SharedServerDAO.UpdateServerUAM(tx, req.ServerId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledServerUAM 查找服务UAM设置
func (this *ServerService) FindEnabledServerUAM(ctx context.Context, req *pb.FindEnabledServerUAMRequest) (*pb.FindEnabledServerUAMResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	uamJSON, err := models.SharedServerDAO.FindServerUAM(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledServerUAMResponse{
		UamJSON: uamJSON,
	}, nil
}
