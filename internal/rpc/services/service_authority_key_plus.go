// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package services

import (
	"context"
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/systemconfigs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/authority"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	plusutils "github.com/TeaOSLab/EdgePlus/pkg/utils"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

// AuthorityKeyService 版本认证
type AuthorityKeyService struct {
	BaseService
}

// UpdateAuthorityKey 设置Key
func (this *AuthorityKeyService) UpdateAuthorityKey(ctx context.Context, req *pb.UpdateAuthorityKeyRequest) (*pb.RPCSuccess, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAuthority)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()

	key, err := plusutils.DecodeKey([]byte(req.Value))
	if err != nil {
		return nil, err
	}

	var addresses = []string{}
	var macAddresses = key.MacAddresses
	for _, addr := range macAddresses {
		addresses = append(addresses, types.String(addr))
	}

	err = authority.SharedAuthorityKeyDAO.UpdateKey(tx, req.Value, key.DayFrom, key.DayTo, key.Hostname, addresses, key.Company)
	if err != nil {
		return nil, err
	}

	// 设置显示财务管理
	if key.IsValid() {
		adminConfig, err := models.SharedSysSettingDAO.ReadAdminUIConfig(tx, nil)
		if err != nil {
			return nil, err
		}
		if adminConfig != nil {
			adminConfig.ShowFinance = true
			adminConfigJSON, err := json.Marshal(adminConfig)
			if err != nil {
				return nil, err
			}
			err = models.SharedSysSettingDAO.UpdateSetting(tx, systemconfigs.SettingCodeAdminUIConfig, adminConfigJSON)
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// ReadAuthorityKey 读取Key
func (this *AuthorityKeyService) ReadAuthorityKey(ctx context.Context, req *pb.ReadAuthorityKeyRequest) (*pb.ReadAuthorityKeyResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeMonitor, rpcutils.UserTypeProvider, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	key, err := authority.SharedAuthorityKeyDAO.ReadKey(tx)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &pb.ReadAuthorityKeyResponse{AuthorityKey: nil}, nil
	}

	if len(key.Value) == 0 {
		return &pb.ReadAuthorityKeyResponse{AuthorityKey: nil}, nil
	}

	m, err := plusutils.DecodeKey([]byte(key.Value))
	if err != nil {
		return nil, err
	}

	teaconst.MaxNodes = int32(m.Nodes)

	if len(m.Components) == 0 {
		m.Components = []string{"*"}
	}

	return &pb.ReadAuthorityKeyResponse{AuthorityKey: &pb.AuthorityKey{
		Value:        key.Value,
		DayFrom:      m.DayFrom,
		DayTo:        m.DayTo,
		Nodes:        int32(m.Nodes),
		Hostname:     m.Hostname,
		MacAddresses: m.MacAddresses,
		Company:      m.Company,
		UpdatedAt:    m.UpdatedAt,
		Components:   m.Components,
	}}, nil
}

// ResetAuthorityKey 重置Key
func (this *AuthorityKeyService) ResetAuthorityKey(ctx context.Context, req *pb.ResetAuthorityKeyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	err = authority.SharedAuthorityKeyDAO.ResetKey(nil)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ValidateAuthorityKey 校验Key
func (this *AuthorityKeyService) ValidateAuthorityKey(ctx context.Context, req *pb.ValidateAuthorityKeyRequest) (*pb.ValidateAuthorityKeyResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	m, err := plusutils.DecodeKey([]byte(req.Key))
	if err != nil {
		return &pb.ValidateAuthorityKeyResponse{IsOk: false, Error: err.Error()}, nil
	}

	var dayTo = m.DayTo
	if dayTo < timeutil.Format("Y-m-d") {
		return &pb.ValidateAuthorityKeyResponse{IsOk: false, Error: "激活码已于" + dayTo + "过期"}, nil
	}

	return &pb.ValidateAuthorityKeyResponse{IsOk: true}, nil
}
