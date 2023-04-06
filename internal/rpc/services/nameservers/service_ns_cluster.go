// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package nameservers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/ddosconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
)

// NSClusterService 域名服务集群相关服务
type NSClusterService struct {
	services.BaseService
}

// CreateNSCluster 创建集群
func (this *NSClusterService) CreateNSCluster(ctx context.Context, req *pb.CreateNSClusterRequest) (*pb.CreateNSClusterResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()

	// SOA
	var soaConfig = dnsconfigs.DefaultNSSOAConfig()
	if len(req.SoaJSON) > 0 {
		err = json.Unmarshal(req.SoaJSON, soaConfig)
		if err != nil {
			return nil, err
		}
		err = soaConfig.Init()
		if err != nil {
			return nil, errors.New("validate SOA config failed: " + err.Error())
		}
	}

	// 校验管理员邮箱
	if len(req.Email) == 0 {
		return nil, errors.New("required 'email'")
	}
	if !utils.ValidateEmail(req.Email) {
		return nil, errors.New("invalid email format '" + req.Email + "'")
	}

	// 校验访问日志配置
	var accessLogRef = &dnsconfigs.NSAccessLogRef{}
	if len(req.AccessLogJSON) > 0 {
		err = json.Unmarshal(req.AccessLogJSON, accessLogRef)
		if err != nil {
			return nil, errors.New("invalid accessLogJSON: " + err.Error())
		}
		err = accessLogRef.Init()
		if err != nil {
			return nil, errors.New("validate accessLogJSON failed: " + err.Error())
		}
	}

	clusterId, err := models.SharedNSClusterDAO.CreateCluster(tx, req.Name, req.Email, req.AccessLogJSON, req.Hosts, soaConfig)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSClusterResponse{NsClusterId: clusterId}, nil
}

// UpdateNSCluster 修改集群
func (this *NSClusterService) UpdateNSCluster(ctx context.Context, req *pb.UpdateNSClusterRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()

	// 校验管理员邮箱
	if len(req.Email) == 0 {
		return nil, errors.New("required 'email'")
	}
	if !utils.ValidateEmail(req.Email) {
		return nil, errors.New("invalid email format '" + req.Email + "'")
	}

	err = models.SharedNSClusterDAO.UpdateCluster(tx, req.NsClusterId, req.Name, req.Email, req.Hosts, req.IsOn, req.TimeZone, req.AutoRemoteStart)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSClusterAccessLog 查找集群访问日志配置
func (this *NSClusterService) FindNSClusterAccessLog(ctx context.Context, req *pb.FindNSClusterAccessLogRequest) (*pb.FindNSClusterAccessLogResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	accessLogJSON, err := models.SharedNSClusterDAO.FindClusterAccessLog(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.FindNSClusterAccessLogResponse{AccessLogJSON: accessLogJSON}, nil
}

// UpdateNSClusterAccessLog 修改集群访问日志配置
func (this *NSClusterService) UpdateNSClusterAccessLog(ctx context.Context, req *pb.UpdateNSClusterAccessLogRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 校验访问日志配置
	var accessLogRef = &dnsconfigs.NSAccessLogRef{}
	if len(req.AccessLogJSON) > 0 {
		err = json.Unmarshal(req.AccessLogJSON, accessLogRef)
		if err != nil {
			return nil, errors.New("invalid accessLogJSON: " + err.Error())
		}
		err = accessLogRef.Init()
		if err != nil {
			return nil, errors.New("validate accessLogJSON failed: " + err.Error())
		}
	}

	err = models.SharedNSClusterDAO.UpdateClusterAccessLog(tx, req.NsClusterId, req.AccessLogJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSCluster 删除集群
func (this *NSClusterService) DeleteNSCluster(ctx context.Context, req *pb.DeleteNSCluster) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.DisableNSCluster(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	// 删除任务
	err = models.SharedNodeTaskDAO.DeleteAllClusterTasks(tx, nodeconfigs.NodeRoleDNS, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNSCluster 查找单个可用集群信息
func (this *NSClusterService) FindNSCluster(ctx context.Context, req *pb.FindNSClusterRequest) (*pb.FindNSClusterResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return &pb.FindNSClusterResponse{NsCluster: nil}, nil
	}
	return &pb.FindNSClusterResponse{NsCluster: &pb.NSCluster{
		Id:              int64(cluster.Id),
		IsOn:            cluster.IsOn,
		Name:            cluster.Name,
		Email:           cluster.Email,
		Hosts:           cluster.DecodeHosts(),
		InstallDir:      cluster.InstallDir,
		TcpJSON:         cluster.Tcp,
		TlsJSON:         cluster.Tls,
		UdpJSON:         cluster.Udp,
		TimeZone:        cluster.TimeZone,
		AutoRemoteStart: cluster.AutoRemoteStart,
		AnswerJSON:      cluster.Answer,
		SoaJSON:         cluster.Soa,
	}}, nil
}

// CountAllNSClusters 计算所有可用集群的数量
func (this *NSClusterService) CountAllNSClusters(ctx context.Context, req *pb.CountAllNSClustersRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	count, err := models.SharedNSClusterDAO.CountAllEnabledClusters(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListNSClusters 列出单页可用集群
func (this *NSClusterService) ListNSClusters(ctx context.Context, req *pb.ListNSClustersRequest) (*pb.ListNSClustersResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	clusters, err := models.SharedNSClusterDAO.ListEnabledClusters(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbClusters = []*pb.NSCluster{}
	for _, cluster := range clusters {
		pbClusters = append(pbClusters, &pb.NSCluster{
			Id:         int64(cluster.Id),
			IsOn:       cluster.IsOn,
			Name:       cluster.Name,
			Hosts:      cluster.DecodeHosts(),
			InstallDir: cluster.InstallDir,
		})
	}
	return &pb.ListNSClustersResponse{NsClusters: pbClusters}, nil
}

// FindAllNSClusters 查找所有可用集群
func (this *NSClusterService) FindAllNSClusters(ctx context.Context, req *pb.FindAllNSClustersRequest) (*pb.FindAllNSClustersResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	clusters, err := models.SharedNSClusterDAO.FindAllEnabledClusters(tx)
	if err != nil {
		return nil, err
	}
	var pbClusters = []*pb.NSCluster{}
	for _, cluster := range clusters {
		pbClusters = append(pbClusters, &pb.NSCluster{
			Id:         int64(cluster.Id),
			IsOn:       cluster.IsOn,
			Name:       cluster.Name,
			InstallDir: cluster.InstallDir,
		})
	}
	return &pb.FindAllNSClustersResponse{NsClusters: pbClusters}, nil
}

// UpdateNSClusterRecursionConfig 设置递归DNS配置
func (this *NSClusterService) UpdateNSClusterRecursionConfig(ctx context.Context, req *pb.UpdateNSClusterRecursionConfigRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// 校验配置
	var config = &dnsconfigs.NSRecursionConfig{}
	err = json.Unmarshal(req.RecursionJSON, config)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateRecursion(tx, req.NsClusterId, req.RecursionJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSClusterRecursionConfig 读取递归DNS配置
func (this *NSClusterService) FindNSClusterRecursionConfig(ctx context.Context, req *pb.FindNSClusterRecursionConfigRequest) (*pb.FindNSClusterRecursionConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	recursion, err := models.SharedNSClusterDAO.FindClusterRecursion(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.FindNSClusterRecursionConfigResponse{
		RecursionJSON: recursion,
	}, nil
}

// FindNSClusterTCPConfig 查找集群的TCP设置
func (this *NSClusterService) FindNSClusterTCPConfig(ctx context.Context, req *pb.FindNSClusterTCPConfigRequest) (*pb.FindNSClusterTCPConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	tcpJSON, err := models.SharedNSClusterDAO.FindClusterTCP(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.FindNSClusterTCPConfigResponse{
		TcpJSON: tcpJSON,
	}, nil
}

// UpdateNSClusterTCP 修改集群的TCP设置
func (this *NSClusterService) UpdateNSClusterTCP(ctx context.Context, req *pb.UpdateNSClusterTCPRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var config = &serverconfigs.TCPProtocolConfig{}
	err = json.Unmarshal(req.TcpJSON, config)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateClusterTCP(tx, req.NsClusterId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNSClusterTLSConfig 查找集群的TLS设置
func (this *NSClusterService) FindNSClusterTLSConfig(ctx context.Context, req *pb.FindNSClusterTLSConfigRequest) (*pb.FindNSClusterTLSConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	tlsJSON, err := models.SharedNSClusterDAO.FindClusterTLS(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.FindNSClusterTLSConfigResponse{
		TlsJSON: tlsJSON,
	}, nil
}

// UpdateNSClusterTLS 修改集群的TLS设置
func (this *NSClusterService) UpdateNSClusterTLS(ctx context.Context, req *pb.UpdateNSClusterTLSRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var config = &serverconfigs.TLSProtocolConfig{}
	err = json.Unmarshal(req.TlsJSON, config)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateClusterTLS(tx, req.NsClusterId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNSClusterUDPConfig 查找集群的UDP设置
func (this *NSClusterService) FindNSClusterUDPConfig(ctx context.Context, req *pb.FindNSClusterUDPConfigRequest) (*pb.FindNSClusterUDPConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	udpJSON, err := models.SharedNSClusterDAO.FindClusterUDP(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.FindNSClusterUDPConfigResponse{
		UdpJSON: udpJSON,
	}, nil
}

// UpdateNSClusterUDP 修改集群的UDP设置
func (this *NSClusterService) UpdateNSClusterUDP(ctx context.Context, req *pb.UpdateNSClusterUDPRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var config = &serverconfigs.UDPProtocolConfig{}
	err = json.Unmarshal(req.UdpJSON, config)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateClusterUDP(tx, req.NsClusterId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllNSClustersWithSSLCertId 计算使用某个SSL证书的集群数量
func (this *NSClusterService) CountAllNSClustersWithSSLCertId(ctx context.Context, req *pb.CountAllNSClustersWithSSLCertIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}
	if len(policyIds) == 0 {
		return this.SuccessCount(0)
	}

	count, err := models.SharedNSClusterDAO.CountAllClustersWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// FindNSClusterDDoSProtection 获取集群的DDoS设置
func (this *NSClusterService) FindNSClusterDDoSProtection(ctx context.Context, req *pb.FindNSClusterDDoSProtectionRequest) (*pb.FindNSClusterDDoSProtectionResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx
	ddosProtection, err := models.SharedNSClusterDAO.FindClusterDDoSProtection(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	if ddosProtection == nil {
		ddosProtection = ddosconfigs.DefaultProtectionConfig()
	}
	ddosProtectionJSON, err := json.Marshal(ddosProtection)
	if err != nil {
		return nil, err
	}

	var result = &pb.FindNSClusterDDoSProtectionResponse{
		DdosProtectionJSON: ddosProtectionJSON,
	}

	return result, nil
}

// UpdateNSClusterDDoSProtection 修改集群的DDoS设置
func (this *NSClusterService) UpdateNSClusterDDoSProtection(ctx context.Context, req *pb.UpdateNSClusterDDoSProtectionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var ddosProtection = &ddosconfigs.ProtectionConfig{}
	err = json.Unmarshal(req.DdosProtectionJSON, ddosProtection)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx
	err = models.SharedNSClusterDAO.UpdateClusterDDoSProtection(tx, req.NsClusterId, ddosProtection)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSClusterHosts 查找NS集群的主机地址
func (this *NSClusterService) FindNSClusterHosts(ctx context.Context, req *pb.FindNSClusterHostsRequest) (*pb.FindNSClusterHostsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	hosts, err := models.SharedNSClusterDAO.FindClusterHosts(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.FindNSClusterHostsResponse{
		Hosts: hosts,
	}, nil
}

// FindAvailableNSHostsForUser 查找用户可以使用的主机地址
func (this *NSClusterService) FindAvailableNSHostsForUser(ctx context.Context, req *pb.FindAvailableNSHostsForUserRequest) (*pb.FindAvailableNSHostsForUserResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	if req.UserId <= 0 {
		return &pb.FindAvailableNSHostsForUserResponse{
			Hosts: nil,
		}, nil
	}

	// 所属集群
	var tx = this.NullTx()
	userConfig, err := models.SharedSysSettingDAO.ReadNSUserConfig(tx)
	if err != nil {
		return nil, err
	}

	if userConfig == nil || userConfig.DefaultClusterId <= 0 {
		return &pb.FindAvailableNSHostsForUserResponse{
			Hosts: nil,
		}, nil
	}

	hosts, err := models.SharedNSClusterDAO.FindClusterHosts(tx, userConfig.DefaultClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.FindAvailableNSHostsForUserResponse{
		Hosts: hosts,
	}, nil
}

// FindNSClusterAnswerConfig 查找应答模式
func (this *NSClusterService) FindNSClusterAnswerConfig(ctx context.Context, req *pb.FindNSClusterAnswerConfigRequest) (*pb.FindNSClusterAnswerConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	config, err := models.SharedNSClusterDAO.FindClusterAnswer(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindNSClusterAnswerConfigResponse{
		AnswerJSON: configJSON,
	}, nil
}

// UpdateNSClusterAnswerConfig 设置应答模式
func (this *NSClusterService) UpdateNSClusterAnswerConfig(ctx context.Context, req *pb.UpdateNSClusterAnswerConfigRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var config = dnsconfigs.DefaultNSAnswerConfig()
	if len(req.AnswerJSON) > 0 {
		err = json.Unmarshal(req.AnswerJSON, config)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateClusterAnswer(tx, req.NsClusterId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNSClusterSOAConfig 查询SOA配置
func (this *NSClusterService) FindNSClusterSOAConfig(ctx context.Context, req *pb.FindNSClusterSOAConfigRequest) (*pb.FindNSClusterSOAConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	config, err := models.SharedNSClusterDAO.FindClusterSOA(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindNSClusterSOAConfigResponse{
		SoaJSON: configJSON,
	}, nil
}

// UpdateNSClusterSOAConfig 修改SOA配置
func (this *NSClusterService) UpdateNSClusterSOAConfig(ctx context.Context, req *pb.UpdateNSClusterSOAConfigRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var config = dnsconfigs.DefaultNSSOAConfig()
	if len(req.SoaJSON) > 0 {
		err = json.Unmarshal(req.SoaJSON, config)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateClusterSOA(tx, req.NsClusterId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
