// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package services

import (
	"context"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/authority"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
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

	m, err := plusutils.Decode([]byte(req.Value))
	if err != nil {
		return nil, err
	}

	var addresses = []string{}
	var macAddresses = m.GetSlice("macAddresses")
	for _, addr := range macAddresses {
		addresses = append(addresses, types.String(addr))
	}

	err = authority.SharedAuthorityKeyDAO.UpdateKey(tx, req.Value, m.GetString("dayFrom"), m.GetString("dayTo"), m.GetString("hostname"), addresses, m.GetString("company"))
	if err != nil {
		return nil, err
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

	m, err := plusutils.Decode([]byte(key.Value))
	if err != nil {
		return nil, err
	}

	macAddresses := []string{}
	if len(key.MacAddresses) > 0 {
		err = json.Unmarshal([]byte(key.MacAddresses), &macAddresses)
		if err != nil {
			return nil, err
		}
	}

	teaconst.MaxNodes = m.GetInt32("nodes")

	return &pb.ReadAuthorityKeyResponse{AuthorityKey: &pb.AuthorityKey{
		Value:        key.Value,
		DayFrom:      m.GetString("dayFrom"),
		DayTo:        m.GetString("dayTo"),
		Nodes:        m.GetInt32("nodes"),
		Hostname:     key.Hostname,
		MacAddresses: macAddresses,
		Company:      key.Company,
		UpdatedAt:    int64(key.UpdatedAt),
	}}, nil
}

// ResetAuthorityKey 重置Key
func (this *AuthorityKeyService) ResetAuthorityKey(ctx context.Context, req *pb.ResetAuthorityKeyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	m, err := plusutils.Decode([]byte(req.Key))
	if err != nil {
		return &pb.ValidateAuthorityKeyResponse{IsOk: false, Error: err.Error()}, nil
	}

	var dayTo = m.GetString("dayTo")
	if dayTo < timeutil.Format("Y-m-d") {
		return &pb.ValidateAuthorityKeyResponse{IsOk: false, Error: "激活码已于" + dayTo + "过期"}, nil
	}

	return &pb.ValidateAuthorityKeyResponse{IsOk: true}, nil
}
