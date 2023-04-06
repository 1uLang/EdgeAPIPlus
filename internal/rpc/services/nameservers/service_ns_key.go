// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package nameservers

import (
	"context"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
)

// NSKeyService NS密钥相关服务
type NSKeyService struct {
	services.BaseService
}

// CreateNSKey 创建密钥
func (this *NSKeyService) CreateNSKey(ctx context.Context, req *pb.CreateNSKeyRequest) (*pb.CreateNSKeyResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	keyId, err := nameservers.SharedNSKeyDAO.CreateKey(tx, req.NsDomainId, req.NsZoneId, req.Name, req.Algo, req.Secret, req.SecretType)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSKeyResponse{NsKeyId: keyId}, nil
}

// UpdateNSKey 修改密钥
func (this *NSKeyService) UpdateNSKey(ctx context.Context, req *pb.UpdateNSKeyRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = nameservers.SharedNSKeyDAO.CheckUserKey(tx, userId, req.NsKeyId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSKeyDAO.UpdateKey(tx, req.NsKeyId, req.Name, req.Algo, req.Secret, req.SecretType, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSKey 删除密钥
func (this *NSKeyService) DeleteNSKey(ctx context.Context, req *pb.DeleteNSKeyRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = nameservers.SharedNSKeyDAO.CheckUserKey(tx, userId, req.NsKeyId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSKeyDAO.DisableNSKey(tx, req.NsKeyId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSKey 查找单个密钥
func (this *NSKeyService) FindNSKey(ctx context.Context, req *pb.FindNSKeyRequest) (*pb.FindNSKeyResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = nameservers.SharedNSKeyDAO.CheckUserKey(tx, userId, req.NsKeyId)
		if err != nil {
			return nil, err
		}
	}

	key, err := nameservers.SharedNSKeyDAO.FindEnabledNSKey(tx, req.NsKeyId)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &pb.FindNSKeyResponse{NsKey: nil}, nil
	}
	return &pb.FindNSKeyResponse{
		NsKey: &pb.NSKey{
			Id:         int64(key.Id),
			IsOn:       key.IsOn,
			Name:       key.Name,
			Algo:       key.Algo,
			Secret:     key.Secret,
			SecretType: key.SecretType,
		},
	}, nil
}

// CountAllNSKeys 计算密钥数量
func (this *NSKeyService) CountAllNSKeys(ctx context.Context, req *pb.CountAllNSKeysRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	count, err := nameservers.SharedNSKeyDAO.CountEnabledKeys(tx, req.NsDomainId, req.NsZoneId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListNSKeys 列出单页密钥
func (this *NSKeyService) ListNSKeys(ctx context.Context, req *pb.ListNSKeysRequest) (*pb.ListNSKeysResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	keys, err := nameservers.SharedNSKeyDAO.ListEnabledKeys(tx, req.NsDomainId, req.NsZoneId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbKeys = []*pb.NSKey{}
	for _, key := range keys {
		pbKeys = append(pbKeys, &pb.NSKey{
			Id:         int64(key.Id),
			IsOn:       key.IsOn,
			Name:       key.Name,
			Algo:       key.Algo,
			Secret:     key.Secret,
			SecretType: key.SecretType,
		})
	}
	return &pb.ListNSKeysResponse{NsKeys: pbKeys}, nil
}

// ListNSKeysAfterVersion 根据版本列出一组密钥
func (this *NSKeyService) ListNSKeysAfterVersion(ctx context.Context, req *pb.ListNSKeysAfterVersionRequest) (*pb.ListNSKeysAfterVersionResponse, error) {
	_, err := this.ValidateNSNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if req.Size <= 0 {
		req.Size = 2000
	}
	keys, err := nameservers.SharedNSKeyDAO.ListKeysAfterVersion(tx, req.Version, req.Size)
	if err != nil {
		return nil, err
	}
	var pbKeys = []*pb.NSKey{}
	for _, key := range keys {
		var pbDomain *pb.NSDomain
		var pbZone *pb.NSZone

		if key.DomainId > 0 {
			pbDomain = &pb.NSDomain{Id: int64(key.DomainId)}
		}
		if key.ZoneId > 0 {
			pbZone = &pb.NSZone{Id: int64(key.ZoneId)}
		}

		pbKeys = append(pbKeys, &pb.NSKey{
			Id:         int64(key.Id),
			IsOn:       key.IsOn,
			Name:       "",
			Algo:       key.Algo,
			Secret:     key.Secret,
			SecretType: key.SecretType,
			IsDeleted:  key.State == nameservers.NSKeyStateDisabled,
			Version:    int64(key.Version),
			NsDomain:   pbDomain,
			NsZone:     pbZone,
		})
	}
	return &pb.ListNSKeysAfterVersionResponse{NsKeys: pbKeys}, nil
}
