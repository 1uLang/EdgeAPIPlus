// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package services

import (
	"context"
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
)

// ScriptService 脚本相关服务
type ScriptService struct {
	BaseService
}

// CreateScript 添加脚本
func (this *ScriptService) CreateScript(ctx context.Context, req *pb.CreateScriptRequest) (*pb.CreateScriptResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	scriptId, err := models.SharedScriptDAO.CreateScript(tx, userId, req.Name, req.Filename, req.Code)
	if err != nil {
		return nil, err
	}
	return &pb.CreateScriptResponse{ScriptId: scriptId}, nil
}

// DeleteScript 删除脚本
func (this *ScriptService) DeleteScript(ctx context.Context, req *pb.DeleteScriptRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedScriptDAO.CheckUserScript(tx, userId, req.ScriptId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedScriptDAO.DisableScript(tx, req.ScriptId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledScripts 计算脚本数量
func (this *ScriptService) CountAllEnabledScripts(ctx context.Context, req *pb.CountAllEnabledScriptsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := models.SharedScriptDAO.CountAllEnabledScripts(tx, userId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledScripts 列出单页脚本
func (this *ScriptService) ListEnabledScripts(ctx context.Context, req *pb.ListEnabledScriptsRequest) (*pb.ListEnabledScriptsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	scripts, err := models.SharedScriptDAO.ListEnabledScripts(tx, req.UserId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbScripts = []*pb.Script{}
	for _, script := range scripts {
		pbScripts = append(pbScripts, &pb.Script{
			Id:        int64(script.Id),
			UserId:    int64(script.UserId),
			IsOn:      script.IsOn,
			Name:      script.Name,
			Filename:  script.Filename,
			Code:      script.Code,
			UpdatedAt: int64(script.UpdatedAt),
		})
	}

	return &pb.ListEnabledScriptsResponse{Scripts: pbScripts}, nil
}

// PublishScripts 发布脚本
func (this *ScriptService) PublishScripts(ctx context.Context, req *pb.PublishScriptsRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	err = models.SharedScriptHistoryDAO.PublishScripts(tx, userId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CheckScriptUpdates 检查脚本是否需要有更新
func (this *ScriptService) CheckScriptUpdates(ctx context.Context, req *pb.CheckScriptUpdatesRequest) (*pb.CheckScriptUpdatesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	hasUpdates, version, err := models.SharedScriptHistoryDAO.CheckScriptsUpdates(tx, userId)
	if err != nil {
		return nil, err
	}

	return &pb.CheckScriptUpdatesResponse{
		HasUpdates: hasUpdates,
		Version:    version,
	}, nil
}

// FindEnabledScript 查找单个脚本
func (this *ScriptService) FindEnabledScript(ctx context.Context, req *pb.FindEnabledScriptRequest) (*pb.FindEnabledScriptResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedScriptDAO.CheckUserScript(tx, userId, req.ScriptId)
		if err != nil {
			return nil, err
		}
	}

	script, err := models.SharedScriptDAO.FindEnabledScript(tx, req.ScriptId)
	if err != nil {
		return nil, err
	}

	if script == nil {
		return &pb.FindEnabledScriptResponse{
			Script: nil,
		}, nil
	}

	return &pb.FindEnabledScriptResponse{
		Script: &pb.Script{
			Id:       int64(script.Id),
			UserId:   int64(script.UserId),
			IsOn:     script.IsOn,
			Name:     script.Name,
			Filename: script.Filename,
			Code:     script.Code,
		},
	}, nil
}

// UpdateScript 修改脚本
func (this *ScriptService) UpdateScript(ctx context.Context, req *pb.UpdateScriptRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedScriptDAO.CheckUserScript(tx, userId, req.ScriptId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedScriptDAO.UpdateScript(tx, req.ScriptId, req.Name, req.Filename, req.Code, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ComposeScriptConfigs 组合脚本配置
func (this *ScriptService) ComposeScriptConfigs(ctx context.Context, req *pb.ComposeScriptConfigsRequest) (*pb.ComposeScriptConfigsResponse, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	configs, err := models.SharedScriptHistoryDAO.ComposeScriptConfigs(tx, 0, nil)
	if err != nil {
		return nil, err
	}

	configsJSON, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}

	return &pb.ComposeScriptConfigsResponse{
		ScriptConfigsJSON: configsJSON,
	}, nil
}
