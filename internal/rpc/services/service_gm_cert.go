package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/gmconfigs"
)

// GMCertService 国密证书相关服务
type GMCertService struct {
	BaseService
}

func (this *GMCertService) DeleteGMCert(ctx context.Context, req *pb.DeleteGMCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err := models.SharedGMCertDAO.CheckUserCert(tx, req.GmCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedGMCertDAO.DisableGMCert(tx, req.GmCertId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CreateGMCert 创建证书
func (this *GMCertService) CreateGMCert(ctx context.Context, req *pb.CreateGMCertRequest) (*pb.CreateGMCertResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	// 用户ID
	if adminId > 0 && req.UserId > 0 {
		userId = req.UserId
	}

	var tx = this.NullTx()

	if req.TimeBeginAt < 0 {
		return nil, errors.New("invalid TimeBeginAt")
	}
	if req.TimeEndAt < 0 {
		return nil, errors.New("invalid TimeEndAt")
	}

	certId, err := models.SharedGMCertDAO.CreateCert(tx, adminId, userId, req.IsOn, req.Name, req.Description, req.ServerName, req.SignCertData, req.SignKeyData, req.EncCertData, req.EncKeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return &pb.CreateGMCertResponse{GmCertId: certId}, nil
}

// FindEnabledGMCertConfig 查找证书配置
func (this *GMCertService) FindEnabledGMCertConfig(ctx context.Context, req *pb.FindEnabledGMCertConfigRequest) (*pb.FindEnabledGMCertConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err := models.SharedGMCertDAO.CheckUserCert(tx, req.GmCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedGMCertDAO.ComposeCertConfig(tx, req.GmCertId, false, nil, nil)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledGMCertConfigResponse{GmCertJSON: configJSON}, nil
}

// CountGMCerts 计算匹配的Cert数量
func (this *GMCertService) CountGMCerts(ctx context.Context, req *pb.CountGMCertRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if adminId > 0 {
		userId = req.UserId
	} else if userId <= 0 {
		return nil, errors.New("invalid user")
	}

	count, err := models.SharedGMCertDAO.CountCerts(tx, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, userId, req.Domains)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListGMCerts 列出单页匹配的Cert
func (this *GMCertService) ListGMCerts(ctx context.Context, req *pb.ListGMCertsRequest) (*pb.ListGMCertsResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if adminId > 0 {
		userId = req.UserId
	} else if userId <= 0 {
		return nil, errors.New("invalid user")
	}

	var tx = this.NullTx()

	certIds, err := models.SharedGMCertDAO.ListCertIds(tx, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, userId, req.Domains, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var certConfigs = []*gmconfigs.GMCertConfig{}
	for _, certId := range certIds {
		certConfig, err := models.SharedGMCertDAO.ComposeCertConfig(tx, certId, false, nil, nil)
		if err != nil {
			return nil, err
		}

		// 这里不需要数据内容
		certConfig.SignCertData = nil
		certConfig.SignKeyData = nil
		certConfig.EncCertData = nil
		certConfig.EncKeyData = nil

		certConfigs = append(certConfigs, certConfig)
	}
	certConfigsJSON, err := json.Marshal(certConfigs)
	if err != nil {
		return nil, err
	}
	return &pb.ListGMCertsResponse{GmCertsJSON: certConfigsJSON}, nil
}

// UpdateGMCert 修改Cert
func (this *GMCertService) UpdateGMCert(ctx context.Context, req *pb.UpdateGMCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.TimeBeginAt < 0 {
		return nil, errors.New("invalid TimeBeginAt")
	}
	if req.TimeEndAt < 0 {
		return nil, errors.New("invalid TimeEndAt")
	}

	// 检查权限
	if userId > 0 {
		err := models.SharedGMCertDAO.CheckUserCert(tx, req.GmCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedGMCertDAO.UpdateCert(tx, req.GmCertId, req.IsOn, req.Name, req.Description, req.ServerName, req.SignCertData, req.SignKeyData, req.EncCertData, req.EncKeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
