// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
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

// UpdateHTTPWebUAM 修改UAM设置
func (this *HTTPWebService) UpdateHTTPWebUAM(ctx context.Context, req *pb.UpdateHTTPWebUAMRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx
	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(tx, userId, req.HttpWebId)
		if err != nil {
			return nil, err
		}
	}

	var config = &serverconfigs.UAMConfig{}
	err = json.Unmarshal(req.UamJSON, config)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebUAM(tx, req.HttpWebId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindHTTPWebUAM 查找UAM设置
func (this *HTTPWebService) FindHTTPWebUAM(ctx context.Context, req *pb.FindHTTPWebUAMRequest) (*pb.FindHTTPWebUAMResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx
	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(tx, userId, req.HttpWebId)
		if err != nil {
			return nil, err
		}
	}

	uamJSON, err := models.SharedHTTPWebDAO.FindWebUAM(tx, req.HttpWebId)
	if err != nil {
		return nil, err
	}
	return &pb.FindHTTPWebUAMResponse{
		UamJSON: uamJSON,
	}, nil
}
