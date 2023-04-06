package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/rpc/dao"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/schedulingconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/sslconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/domainutils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type ServerService struct {
	BaseService
}

// CreateServer 创建服务
func (this *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.CreateServerResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	// 校验用户相关数据
	if userId > 0 {
		// HTTPS
		if len(req.HttpsJSON) > 0 {
			httpsConfig := &serverconfigs.HTTPSProtocolConfig{}
			err = json.Unmarshal(req.HttpsJSON, httpsConfig)
			if err != nil {
				return nil, err
			}
			if httpsConfig.SSLPolicyRef != nil && httpsConfig.SSLPolicyRef.SSLPolicyId > 0 {
				err := models.SharedSSLPolicyDAO.CheckUserPolicy(tx, userId, httpsConfig.SSLPolicyRef.SSLPolicyId)
				if err != nil {
					return nil, err
				}
			}
		}

		// TLS
		if len(req.TlsJSON) > 0 {
			tlsConfig := &serverconfigs.TLSProtocolConfig{}
			err = json.Unmarshal(req.TlsJSON, tlsConfig)
			if err != nil {
				return nil, err
			}
			if tlsConfig.SSLPolicyRef != nil && tlsConfig.SSLPolicyRef.SSLPolicyId > 0 {
				err := models.SharedSSLPolicyDAO.CheckUserPolicy(tx, userId, tlsConfig.SSLPolicyRef.SSLPolicyId)
				if err != nil {
					return nil, err
				}
			}
		}

		// 集群
		nodeClusterId, err := models.SharedUserDAO.FindUserClusterId(tx, userId)
		if err != nil {
			return nil, err
		}
		if nodeClusterId > 0 {
			req.NodeClusterId = nodeClusterId
		}

		// 服务分组
		for _, groupId := range req.ServerGroupIds {
			err := models.SharedServerGroupDAO.CheckUserGroup(tx, userId, groupId)
			if err != nil {
				return nil, err
			}
		}

		// 增加默认分组
		config, err := models.SharedSysSettingDAO.ReadUserServerConfig(tx)
		if err == nil && config.GroupId > 0 && !lists.ContainsInt64(req.ServerGroupIds, config.GroupId) {
			req.ServerGroupIds = append(req.ServerGroupIds, config.GroupId)
		}
	} else if req.UserId > 0 {
		// 集群
		nodeClusterId, err := models.SharedUserDAO.FindUserClusterId(tx, req.UserId)
		if err != nil {
			return nil, err
		}
		if nodeClusterId > 0 {
			req.NodeClusterId = nodeClusterId
		}
	}

	// 是否需要审核
	isAuditing := false
	serverNamesJSON := req.ServerNamesJON
	auditingServerNamesJSON := []byte("[]")
	if userId > 0 {
		// 如果域名不为空的时候需要审核
		if len(serverNamesJSON) > 0 && string(serverNamesJSON) != "[]" {
			globalConfig, err := models.SharedSysSettingDAO.ReadGlobalConfig(tx)
			if err != nil {
				return nil, err
			}
			if globalConfig != nil && globalConfig.HTTPAll.DomainAuditingIsOn {
				isAuditing = true
				serverNamesJSON = []byte("[]")
				auditingServerNamesJSON = req.ServerNamesJON
			}
		}
	}

	// 检查用户套餐
	if req.UserPlanId > 0 {
		userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, req.UserPlanId, nil)
		if err != nil {
			return nil, err
		}
		if userPlan == nil {
			return nil, errors.New("can not find user plan with id '" + types.String(req.UserPlanId) + "'")
		}
		if userId > 0 && int64(userPlan.UserId) != userId {
			return nil, errors.New("invalid user plan")
		}
		if req.UserId > 0 && int64(userPlan.UserId) != req.UserId {
			return nil, errors.New("invalid user plan")
		}

		// 套餐
		plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
		if err != nil {
			return nil, err
		}
		if plan == nil {
			return nil, errors.New("invalid plan: " + types.String(userPlan.PlanId))
		}
		if plan.ClusterId > 0 {
			req.NodeClusterId = int64(plan.ClusterId)
		}

		// 检查是否已经被别的服务所占用
		planServerId, err := models.SharedServerDAO.FindEnabledServerIdWithUserPlanId(tx, req.UserPlanId)
		if err != nil {
			return nil, err
		}
		if planServerId > 0 {
			return nil, errors.New("the user plan is used by another server '" + types.String(planServerId) + "'")
		}
	}

	serverId, err := models.SharedServerDAO.CreateServer(tx, req.AdminId, req.UserId, req.Type, req.Name, req.Description, serverNamesJSON, isAuditing, auditingServerNamesJSON, req.HttpJSON, req.HttpsJSON, req.TcpJSON, req.TlsJSON, req.UnixJSON, req.UdpJSON, req.WebId, req.ReverseProxyJSON, req.NodeClusterId, req.IncludeNodesJSON, req.ExcludeNodesJSON, req.ServerGroupIds, req.UserPlanId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

// UpdateServerBasic 修改服务基本信息
func (this *ServerService) UpdateServerBasic(ctx context.Context, req *pb.UpdateServerBasicRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	var tx = this.NullTx()

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	err = models.SharedServerDAO.UpdateServerBasic(tx, req.ServerId, req.Name, req.Description, req.NodeClusterId, req.KeepOldConfigs, req.IsOn, req.ServerGroupIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerGroupIds 修改服务所在分组
func (this *ServerService) UpdateServerGroupIds(ctx context.Context, req *pb.UpdateServerGroupIdsRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 检查分组IDs
	var serverGroupIds = []int64{}
	for _, groupId := range req.ServerGroupIds {
		if userId > 0 {
			err = models.SharedServerGroupDAO.CheckUserGroup(tx, userId, groupId)
			if err != nil {
				return nil, err
			}
		} else {
			b, err := models.SharedServerGroupDAO.ExistsGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if !b {
				continue
			}
		}
		serverGroupIds = append(serverGroupIds, groupId)
	}

	// 增加默认分组
	if userId > 0 {
		config, err := models.SharedSysSettingDAO.ReadUserServerConfig(tx)
		if err == nil && config.GroupId > 0 && !lists.ContainsInt64(serverGroupIds, config.GroupId) {
			serverGroupIds = append(serverGroupIds, config.GroupId)
		}
	}

	// 修改
	err = models.SharedServerDAO.UpdateServerGroupIds(tx, req.ServerId, serverGroupIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerIsOn 修改服务是否启用
func (this *ServerService) UpdateServerIsOn(ctx context.Context, req *pb.UpdateServerIsOnRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}
	err = models.SharedServerDAO.UpdateServerIsOn(tx, req.ServerId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateServerHTTP 修改HTTP服务
func (this *ServerService) UpdateServerHTTP(ctx context.Context, req *pb.UpdateServerHTTPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerHTTP(tx, req.ServerId, req.HttpJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerHTTPS 修改HTTPS服务
func (this *ServerService) UpdateServerHTTPS(ctx context.Context, req *pb.UpdateServerHTTPSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerHTTPS(tx, req.ServerId, req.HttpsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerTCP 修改TCP服务
func (this *ServerService) UpdateServerTCP(ctx context.Context, req *pb.UpdateServerTCPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTCP(tx, req.ServerId, req.TcpJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerTLS 修改TLS服务
func (this *ServerService) UpdateServerTLS(ctx context.Context, req *pb.UpdateServerTLSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTLS(tx, req.ServerId, req.TlsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerUnix 修改Unix服务
func (this *ServerService) UpdateServerUnix(ctx context.Context, req *pb.UpdateServerUnixRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	var tx = this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerUnix(tx, req.ServerId, req.UnixJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerUDP 修改UDP服务
func (this *ServerService) UpdateServerUDP(ctx context.Context, req *pb.UpdateServerUDPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	var tx = this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerUDP(tx, req.ServerId, req.UdpJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerWeb 修改Web服务
func (this *ServerService) UpdateServerWeb(ctx context.Context, req *pb.UpdateServerWebRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerWeb(tx, req.ServerId, req.WebId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerReverseProxy 修改反向代理服务
func (this *ServerService) UpdateServerReverseProxy(ctx context.Context, req *pb.UpdateServerReverseProxyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerReverseProxy(tx, req.ServerId, req.ReverseProxyJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindServerNames 查找服务的域名设置
func (this *ServerService) FindServerNames(ctx context.Context, req *pb.FindServerNamesRequest) (*pb.FindServerNamesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	serverNamesJSON, isAuditing, auditingAt, auditingServerNamesJSON, auditingResultJSON, err := models.SharedServerDAO.FindServerServerNames(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	// 审核结果
	auditingResult := &pb.ServerNameAuditingResult{}
	if len(auditingResultJSON) > 0 {
		err = json.Unmarshal(auditingResultJSON, auditingResult)
		if err != nil {
			return nil, err
		}
	} else {
		auditingResult.IsOk = true
	}

	return &pb.FindServerNamesResponse{
		ServerNamesJSON:         serverNamesJSON,
		IsAuditing:              isAuditing,
		AuditingAt:              auditingAt,
		AuditingServerNamesJSON: auditingServerNamesJSON,
		AuditingResult:          auditingResult,
	}, nil
}

// UpdateServerNames 修改域名服务
func (this *ServerService) UpdateServerNames(ctx context.Context, req *pb.UpdateServerNamesRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 转换为小写
	var serverNameConfigs = []*serverconfigs.ServerNameConfig{}
	if len(req.ServerNamesJSON) > 0 {
		err = json.Unmarshal(req.ServerNamesJSON, &serverNameConfigs)
		if err != nil {
			return nil, err
		}
		if len(serverNameConfigs) > 0 {
			for _, serverName := range serverNameConfigs {
				serverName.Normalize()
			}
			req.ServerNamesJSON, err = json.Marshal(serverNameConfigs)
			if err != nil {
				return nil, err
			}
		}
	}

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}

		// 是否需要审核
		globalConfig, err := models.SharedSysSettingDAO.ReadGlobalConfig(tx)
		if err != nil {
			return nil, err
		}
		if globalConfig != nil && globalConfig.HTTPAll.DomainAuditingIsOn {
			err = models.SharedServerDAO.UpdateAuditingServerNames(tx, req.ServerId, true, req.ServerNamesJSON)
			if err != nil {
				return nil, err
			}

			// 发送审核通知
			err = models.SharedMessageDAO.CreateMessage(tx, 0, 0, models.MessageTypeServerNamesRequireAuditing, models.MessageLevelWarning, "有新的网站域名需要审核", "有新的网站域名需要审核", maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())

			return this.Success()
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerNames(tx, req.ServerId, req.ServerNamesJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerNamesAuditing 审核服务的域名设置
func (this *ServerService) UpdateServerNamesAuditing(ctx context.Context, req *pb.UpdateServerNamesAuditingRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.AuditingResult == nil {
		return nil, errors.New("'result' should not be nil")
	}

	var tx = this.NullTx()

	err = models.SharedServerDAO.UpdateServerAuditing(tx, req.ServerId, req.AuditingResult)
	if err != nil {
		return nil, err
	}

	// 发送消息提醒
	_, userId, err := models.SharedServerDAO.FindServerAdminIdAndUserId(tx, req.ServerId)
	if userId > 0 {
		if req.AuditingResult.IsOk {
			subject := "服务域名审核通过"
			msg := "服务域名审核通过"
			err = models.SharedMessageDAO.CreateMessage(tx, 0, userId, models.MessageTypeServerNamesAuditingSuccess, models.MessageLevelSuccess, subject, msg, maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())
			if err != nil {
				return nil, err
			}
		} else {
			subject := "服务域名审核失败"
			msg := "服务域名审核失败，原因：" + req.AuditingResult.Reason
			err = models.SharedMessageDAO.CreateMessage(tx, 0, userId, models.MessageTypeServerNamesAuditingFailed, models.LevelError, subject, msg, maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// UpdateServerDNS 修改服务的DNS相关设置
func (this *ServerService) UpdateServerDNS(ctx context.Context, req *pb.UpdateServerDNSRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedServerDAO.UpdateServerDNS(tx, req.ServerId, req.SupportCNAME)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// RegenerateServerDNSName 重新生成CNAME
func (this *ServerService) RegenerateServerDNSName(ctx context.Context, req *pb.RegenerateServerDNSNameRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	_, err = models.SharedServerDAO.GenerateServerDNSName(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateServerDNSName 修改服务的CNAME
func (this *ServerService) UpdateServerDNSName(ctx context.Context, req *pb.UpdateServerDNSNameRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var dnsName = req.DnsName

	if req.ServerId <= 0 {
		return nil, errors.New("invalid 'serverId'")
	}

	if len(dnsName) == 0 {
		return nil, errors.New("'dnsName' must not be empty")
	}

	// 处理格式
	dnsName = strings.ToLower(dnsName)
	const maxLen = 30
	if len(dnsName) > maxLen {
		return nil, errors.New("'dnsName' too long than " + types.String(maxLen))
	}
	if !regexp.MustCompile(`^[a-z0-9]{1,` + types.String(maxLen) + `}$`).MatchString(dnsName) {
		return nil, errors.New("invalid 'dnsName': contains invalid character(s)")
	}

	// 检查是否被使用
	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, errors.New("the server is not belong to any cluster")
	}

	serverId, err := models.SharedServerDAO.FindServerIdWithDNSName(tx, clusterId, dnsName)
	if err != nil {
		return nil, err
	}
	if serverId > 0 && serverId != req.ServerId {
		return nil, errors.New("the 'dnsName': " + dnsName + " has already been used")
	}

	err = models.SharedServerDAO.UpdateServerDNSName(tx, req.ServerId, dnsName)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindServerIdWithDNSName 使用CNAME查找服务
func (this *ServerService) FindServerIdWithDNSName(ctx context.Context, req *pb.FindServerIdWithDNSNameRequest) (*pb.FindServerIdWithDNSNameResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.DnsName) == 0 {
		return nil, errors.New("'dnsName' must not be empty")
	}

	var tx = this.NullTx()
	serverId, err := models.SharedServerDAO.FindServerIdWithDNSName(tx, req.NodeClusterId, req.DnsName)
	if err != nil {
		return nil, err
	}

	return &pb.FindServerIdWithDNSNameResponse{
		ServerId: serverId,
	}, nil
}

// CountAllEnabledServersMatch 计算服务数量
func (this *ServerService) CountAllEnabledServersMatch(ctx context.Context, req *pb.CountAllEnabledServersMatchRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersMatch(tx, req.ServerGroupId, req.Keyword, req.UserId, req.NodeClusterId, types.Int8(req.AuditingFlag), utils.SplitStrings(req.ProtocolFamily, ","))
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledServersMatch 列出单页服务
func (this *ServerService) ListEnabledServersMatch(ctx context.Context, req *pb.ListEnabledServersMatchRequest) (*pb.ListEnabledServersMatchResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId
	}

	var order = ""
	if req.TrafficOutAsc {
		order = "trafficOutAsc"
	} else if req.TrafficOutDesc {
		order = "trafficOutDesc"
	}

	servers, err := models.SharedServerDAO.ListEnabledServersMatch(tx, req.Offset, req.Size, req.ServerGroupId, req.Keyword, req.UserId, req.NodeClusterId, req.AuditingFlag, utils.SplitStrings(req.ProtocolFamily, ","), order)
	if err != nil {
		return nil, err
	}
	var result = []*pb.Server{}
	for _, server := range servers {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(server.ClusterId))
		if err != nil {
			return nil, err
		}

		// 分组信息
		var pbGroups = []*pb.ServerGroup{}
		if models.IsNotNull(server.GroupIds) {
			var groupIds = []int64{}
			err = json.Unmarshal(server.GroupIds, &groupIds)
			if err != nil {
				return nil, err
			}
			for _, groupId := range groupIds {
				group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, groupId)
				if err != nil {
					return nil, err
				}
				if group == nil {
					continue
				}
				pbGroups = append(pbGroups, &pb.ServerGroup{
					Id:     int64(group.Id),
					Name:   group.Name,
					UserId: int64(group.UserId),
				})
			}
		}

		// 用户
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(server.UserId))
		if err != nil {
			return nil, err
		}
		var pbUser *pb.User = nil
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Fullname: user.Fullname,
			}
		}

		// 审核结果
		var auditingResult = &pb.ServerNameAuditingResult{}
		if len(server.AuditingResult) > 0 {
			err = json.Unmarshal(server.AuditingResult, auditingResult)
			if err != nil {
				return nil, err
			}
		} else {
			auditingResult.IsOk = true
		}

		// 配置
		config, err := models.SharedServerDAO.ComposeServerConfig(tx, server, nil, false)
		if err != nil {
			return nil, err
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.Server{
			Id:                      int64(server.Id),
			IsOn:                    server.IsOn,
			Type:                    server.Type,
			Config:                  configJSON,
			Name:                    server.Name,
			Description:             server.Description,
			HttpJSON:                server.Http,
			HttpsJSON:               server.Https,
			TcpJSON:                 server.Tcp,
			TlsJSON:                 server.Tls,
			UnixJSON:                server.Unix,
			UdpJSON:                 server.Udp,
			IncludeNodes:            server.IncludeNodes,
			ExcludeNodes:            server.ExcludeNodes,
			ServerNamesJSON:         server.ServerNames,
			IsAuditing:              server.IsAuditing,
			AuditingAt:              int64(server.AuditingAt),
			AuditingServerNamesJSON: server.AuditingServerNames,
			AuditingResult:          auditingResult,
			CreatedAt:               int64(server.CreatedAt),
			DnsName:                 server.DnsName,
			UserPlanId:              int64(server.UserPlanId),
			NodeCluster: &pb.NodeCluster{
				Id:   int64(server.ClusterId),
				Name: clusterName,
			},
			ServerGroups:   pbGroups,
			User:           pbUser,
			BandwidthTime:  server.BandwidthTime,
			BandwidthBytes: int64(server.BandwidthBytes),
		})
	}

	return &pb.ListEnabledServersMatchResponse{Servers: result}, nil
}

// DeleteServer 禁用某服务
func (this *ServerService) DeleteServer(ctx context.Context, req *pb.DeleteServerRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 禁用服务
	err = models.SharedServerDAO.DisableServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledServer 查找单个服务
func (this *ServerService) FindEnabledServer(ctx context.Context, req *pb.FindEnabledServerRequest) (*pb.FindEnabledServerResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if server == nil {
		return &pb.FindEnabledServerResponse{}, nil
	}

	// 集群信息
	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	// 分组信息
	pbGroups := []*pb.ServerGroup{}
	if len(server.GroupIds) > 0 {
		groupIds := []int64{}
		err = json.Unmarshal(server.GroupIds, &groupIds)
		if err != nil {
			return nil, err
		}
		for _, groupId := range groupIds {
			group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if group == nil {
				continue
			}
			pbGroups = append(pbGroups, &pb.ServerGroup{
				Id:     int64(group.Id),
				Name:   group.Name,
				UserId: int64(group.UserId),
			})
		}
	}

	// 用户信息
	var pbUser *pb.User = nil
	if server.UserId > 0 {
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(server.UserId))
		if err != nil {
			return nil, err
		}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}
	}

	// 配置
	config, err := models.SharedServerDAO.ComposeServerConfig(tx, server, nil, userId > 0)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerResponse{Server: &pb.Server{
		Id:           int64(server.Id),
		IsOn:         server.IsOn,
		Type:         server.Type,
		Name:         server.Name,
		Description:  server.Description,
		DnsName:      server.DnsName,
		SupportCNAME: server.SupportCNAME == 1,
		UserPlanId:   int64(server.UserPlanId),

		Config:           configJSON,
		ServerNamesJSON:  server.ServerNames,
		HttpJSON:         server.Http,
		HttpsJSON:        server.Https,
		TcpJSON:          server.Tcp,
		TlsJSON:          server.Tls,
		UnixJSON:         server.Unix,
		UdpJSON:          server.Udp,
		WebId:            int64(server.WebId),
		ReverseProxyJSON: server.ReverseProxy,

		IncludeNodes: server.IncludeNodes,
		ExcludeNodes: server.ExcludeNodes,
		CreatedAt:    int64(server.CreatedAt),
		NodeCluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
		ServerGroups: pbGroups,
		User:         pbUser,
	}}, nil
}

// FindEnabledServerConfig 查找服务配置
func (this *ServerService) FindEnabledServerConfig(ctx context.Context, req *pb.FindEnabledServerConfigRequest) (*pb.FindEnabledServerConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, req.ServerId, false)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return &pb.FindEnabledServerConfigResponse{ServerJSON: nil}, nil
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledServerConfigResponse{ServerJSON: configJSON}, nil
}

// FindEnabledServerType 查找服务的服务类型
func (this *ServerService) FindEnabledServerType(ctx context.Context, req *pb.FindEnabledServerTypeRequest) (*pb.FindEnabledServerTypeResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	serverType, err := models.SharedServerDAO.FindEnabledServerType(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerTypeResponse{Type: serverType}, nil
}

// FindAndInitServerReverseProxyConfig 查找反向代理设置
func (this *ServerService) FindAndInitServerReverseProxyConfig(ctx context.Context, req *pb.FindAndInitServerReverseProxyConfigRequest) (*pb.FindAndInitServerReverseProxyConfigResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	reverseProxyRef, err := models.SharedServerDAO.FindReverseProxyRef(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, userId, nil, nil, nil)
		if err != nil {
			return nil, err
		}

		reverseProxyRef = &serverconfigs.ReverseProxyRef{
			IsOn:           false,
			ReverseProxyId: reverseProxyId,
		}
		refJSON, err := json.Marshal(reverseProxyRef)
		if err != nil {
			return nil, err
		}
		err = models.SharedServerDAO.UpdateServerReverseProxy(tx, req.ServerId, refJSON)
		if err != nil {
			return nil, err
		}
	}

	reverseProxyConfig, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, nil)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(reverseProxyConfig)
	if err != nil {
		return nil, err
	}

	refJSON, err := json.Marshal(reverseProxyRef)
	if err != nil {
		return nil, err
	}

	return &pb.FindAndInitServerReverseProxyConfigResponse{ReverseProxyJSON: configJSON, ReverseProxyRefJSON: refJSON}, nil
}

// FindAndInitServerWebConfig 初始化Web设置
func (this *ServerService) FindAndInitServerWebConfig(ctx context.Context, req *pb.FindAndInitServerWebConfigRequest) (*pb.FindAndInitServerWebConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindAndInitServerWebConfigResponse{WebJSON: configJSON}, nil
}

// CountAllEnabledServersWithSSLCertId 计算使用某个SSL证书的服务数量
func (this *ServerService) CountAllEnabledServersWithSSLCertId(ctx context.Context, req *pb.CountAllEnabledServersWithSSLCertIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	if userId > 0 {
		// TODO 校验权限
	}

	var tx = this.NullTx()

	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}

	if len(policyIds) == 0 {
		return this.SuccessCount(0)
	}

	count, err := models.SharedServerDAO.CountAllEnabledServersWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// FindAllEnabledServersWithSSLCertId 查找使用某个SSL证书的所有服务
func (this *ServerService) FindAllEnabledServersWithSSLCertId(ctx context.Context, req *pb.FindAllEnabledServersWithSSLCertIdRequest) (*pb.FindAllEnabledServersWithSSLCertIdResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}

	var tx = this.NullTx()

	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}
	if len(policyIds) == 0 {
		return &pb.FindAllEnabledServersWithSSLCertIdResponse{Servers: nil}, nil
	}

	servers, err := models.SharedServerDAO.FindAllEnabledServersWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return nil, err
	}
	result := []*pb.Server{}
	for _, server := range servers {
		result = append(result, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
			IsOn: server.IsOn,
			Type: server.Type,
		})
	}
	return &pb.FindAllEnabledServersWithSSLCertIdResponse{Servers: result}, nil
}

// CountAllEnabledServersWithNodeClusterId 计算运行在某个集群上的所有服务数量
func (this *ServerService) CountAllEnabledServersWithNodeClusterId(ctx context.Context, req *pb.CountAllEnabledServersWithNodeClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersWithNodeClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountAllEnabledServersWithServerGroupId 计算使用某个分组的服务数量
func (this *ServerService) CountAllEnabledServersWithServerGroupId(ctx context.Context, req *pb.CountAllEnabledServersWithServerGroupIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersWithGroupId(tx, req.ServerGroupId, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// NotifyServersChange 通知更新
func (this *ServerService) NotifyServersChange(ctx context.Context, _ *pb.NotifyServersChangeRequest) (*pb.NotifyServersChangeResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	clusterIds, err := models.SharedNodeClusterDAO.FindAllEnableClusterIds(tx)
	if err != nil {
		return nil, err
	}
	for _, clusterId := range clusterIds {
		err = models.SharedNodeClusterDAO.NotifyUpdate(tx, clusterId)
		if err != nil {
			return nil, err
		}
	}

	return &pb.NotifyServersChangeResponse{}, nil
}

// FindAllEnabledServersDNSWithNodeClusterId 取得某个集群下的所有服务相关的DNS
func (this *ServerService) FindAllEnabledServersDNSWithNodeClusterId(ctx context.Context, req *pb.FindAllEnabledServersDNSWithNodeClusterIdRequest) (*pb.FindAllEnabledServersDNSWithNodeClusterIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	servers, err := models.SharedServerDAO.FindAllServersDNSWithClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.ServerDNSInfo{}
	for _, server := range servers {
		// 如果子域名为空
		if len(server.DnsName) == 0 {
			// 自动生成子域名
			dnsName, err := models.SharedServerDAO.GenerateServerDNSName(tx, int64(server.Id))
			if err != nil {
				return nil, err
			}
			server.DnsName = dnsName
		}

		result = append(result, &pb.ServerDNSInfo{
			Id:      int64(server.Id),
			Name:    server.Name,
			DnsName: server.DnsName,
		})
	}

	return &pb.FindAllEnabledServersDNSWithNodeClusterIdResponse{Servers: result}, nil
}

// FindEnabledServerDNS 查找单个服务的DNS信息
func (this *ServerService) FindEnabledServerDNS(ctx context.Context, req *pb.FindEnabledServerDNSRequest) (*pb.FindEnabledServerDNSResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	dnsName, err := models.SharedServerDAO.FindServerDNSName(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	supportCNAME, err := models.SharedServerDAO.FindServerSupportCNAME(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	var pbDomain *pb.DNSDomain = nil
	if clusterId > 0 {
		clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId, nil)
		if err != nil {
			return nil, err
		}
		if clusterDNS != nil {
			domainId := int64(clusterDNS.DnsDomainId)
			if domainId > 0 {
				domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId, nil)
				if err != nil {
					return nil, err
				}
				if domain != nil {
					pbDomain = &pb.DNSDomain{
						Id:   domainId,
						Name: domain.Name,
					}
				}
			}
		}
	}

	return &pb.FindEnabledServerDNSResponse{
		DnsName:      dnsName,
		Domain:       pbDomain,
		SupportCNAME: supportCNAME,
	}, nil
}

// CheckUserServer 检查服务是否属于某个用户
func (this *ServerService) CheckUserServer(ctx context.Context, req *pb.CheckUserServerRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllEnabledServerNamesWithUserId 查找一个用户下的所有域名列表
func (this *ServerService) FindAllEnabledServerNamesWithUserId(ctx context.Context, req *pb.FindAllEnabledServerNamesWithUserIdRequest) (*pb.FindAllEnabledServerNamesWithUserIdResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	servers, err := models.SharedServerDAO.FindAllEnabledServersWithUserId(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	serverNames := []string{}
	for _, server := range servers {
		if models.IsNotNull(server.ServerNames) {
			serverNameConfigs := []*serverconfigs.ServerNameConfig{}
			err = json.Unmarshal(server.ServerNames, &serverNameConfigs)
			if err != nil {
				return nil, err
			}
			for _, config := range serverNameConfigs {
				if len(config.SubNames) == 0 {
					serverNames = append(serverNames, config.Name)
				} else {
					serverNames = append(serverNames, config.SubNames...)
				}
			}
		}
	}
	return &pb.FindAllEnabledServerNamesWithUserIdResponse{ServerNames: serverNames}, nil
}

// FindEnabledUserServerBasic 查找服务基本信息
func (this *ServerService) FindEnabledUserServerBasic(ctx context.Context, req *pb.FindEnabledUserServerBasicRequest) (*pb.FindEnabledUserServerBasicResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	server, err := models.SharedServerDAO.FindEnabledServerBasic(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return &pb.FindEnabledUserServerBasicResponse{Server: nil}, nil
	}

	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledUserServerBasicResponse{Server: &pb.Server{
		Id:          int64(server.Id),
		Name:        server.Name,
		Description: server.Description,
		IsOn:        server.IsOn,
		Type:        server.Type,
		NodeCluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
	}}, nil
}

// UpdateEnabledUserServerBasic 修改用户服务基本信息
func (this *ServerService) UpdateEnabledUserServerBasic(ctx context.Context, req *pb.UpdateEnabledUserServerBasicRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedServerDAO.UpdateUserServerBasic(tx, req.ServerId, req.Name)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UploadServerHTTPRequestStat 上传待统计数据
func (this *ServerService) UploadServerHTTPRequestStat(ctx context.Context, req *pb.UploadServerHTTPRequestStatRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	month := req.Month
	if len(month) == 0 {
		month = timeutil.Format("Ym")
	}

	day := req.Day
	if len(day) == 0 {
		day = timeutil.Format("Ymd")
	}

	// 区域
	for _, result := range req.RegionCities {
		// IP => 地理位置
		err := func() error {
			// 区域
			if len(result.CountryName) > 0 {
				countryId, err := regions.SharedRegionCountryDAO.FindCountryIdWithNameCacheable(tx, result.CountryName)
				if err != nil {
					return err
				}
				if countryId > 0 {
					countryKey := fmt.Sprintf("%d@%d@%s", result.ServerId, countryId, day)
					serverStatLocker.Lock()
					stat, ok := serverHTTPCountryStatMap[countryKey]
					if !ok {
						stat = &TrafficStat{}
						serverHTTPCountryStatMap[countryKey] = stat
					}
					stat.CountRequests += result.CountRequests
					stat.Bytes += result.Bytes
					stat.CountAttackRequests += result.CountAttackRequests
					stat.AttackBytes += result.AttackBytes
					serverStatLocker.Unlock()

					// 省份
					if len(result.ProvinceName) > 0 {
						provinceId, err := regions.SharedRegionProvinceDAO.FindProvinceIdWithNameCacheable(tx, countryId, result.ProvinceName)
						if err != nil {
							return err
						}
						if provinceId > 0 {
							key := fmt.Sprintf("%d@%d@%s", result.ServerId, provinceId, month)
							serverStatLocker.Lock()
							serverHTTPProvinceStatMap[key] += result.CountRequests
							serverStatLocker.Unlock()

							// 城市
							if len(result.CityName) > 0 {
								cityId, err := regions.SharedRegionCityDAO.FindCityIdWithNameCacheable(tx, provinceId, result.CityName)
								if err != nil {
									return err
								}
								if cityId > 0 {
									key := fmt.Sprintf("%d@%d@%s", result.ServerId, cityId, month)
									serverStatLocker.Lock()
									serverHTTPCityStatMap[key] += result.CountRequests
									serverStatLocker.Unlock()
								}
							}

						}
					}
				}
			}

			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// 运营商
	for _, result := range req.RegionProviders {
		// IP => 地理位置
		err := func() error {
			if len(result.Name) == 0 {
				return nil
			}
			providerId, err := regions.SharedRegionProviderDAO.FindProviderIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if providerId > 0 {
				key := fmt.Sprintf("%d@%d@%s", result.ServerId, providerId, month)
				serverStatLocker.Lock()
				serverHTTPProviderStatMap[key] += result.Count
				serverStatLocker.Unlock()
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// OS
	for _, result := range req.Systems {
		err := func() error {
			if len(result.Name) == 0 {
				return nil
			}

			systemId, err := models.SharedClientSystemDAO.FindSystemIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if systemId == 0 {
				systemId, err = models.SharedClientSystemDAO.CreateSystem(tx, result.Name)
				if err != nil {
					return err
				}
			}
			key := fmt.Sprintf("%d@%d@%s@%s", result.ServerId, systemId, result.Version, month)
			serverStatLocker.Lock()
			serverHTTPSystemStatMap[key] += result.Count
			serverStatLocker.Unlock()
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// Browser
	for _, result := range req.Browsers {
		err := func() error {
			if len(result.Name) == 0 {
				return nil
			}

			browserId, err := models.SharedClientBrowserDAO.FindBrowserIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if browserId == 0 {
				browserId, err = models.SharedClientBrowserDAO.CreateBrowser(tx, result.Name)
				if err != nil {
					return err
				}
			}
			key := fmt.Sprintf("%d@%d@%s@%s", result.ServerId, browserId, result.Version, month)
			serverStatLocker.Lock()
			serverHTTPBrowserStatMap[key] += result.Count
			serverStatLocker.Unlock()
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// 防火墙
	for _, result := range req.HttpFirewallRuleGroups {
		err := func() error {
			if result.HttpFirewallRuleGroupId <= 0 {
				return nil
			}
			key := fmt.Sprintf("%d@%d@%s@%s", result.ServerId, result.HttpFirewallRuleGroupId, result.Action, day)
			serverStatLocker.Lock()
			serverHTTPFirewallRuleGroupStatMap[key] += result.Count
			serverStatLocker.Unlock()
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// CheckServerNameDuplicationInNodeCluster 检查域名是否已经存在
func (this *ServerService) CheckServerNameDuplicationInNodeCluster(ctx context.Context, req *pb.CheckServerNameDuplicationInNodeClusterRequest) (*pb.CheckServerNameDuplicationInNodeClusterResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(req.ServerNames) == 0 {
		return &pb.CheckServerNameDuplicationInNodeClusterResponse{DuplicatedServerNames: nil}, nil
	}

	var tx = this.NullTx()
	var checkFunc func(tx *dbs.Tx, clusterId int64, serverName string, excludeServerId int64, supportWildcard bool) (bool, error)
	if req.All {
		checkFunc = models.SharedServerDAO.ExistServerNameInClusterAll
	} else {
		checkFunc = models.SharedServerDAO.ExistServerNameInCluster
	}
	var duplicatedServerNames = []string{}
	for _, serverName := range req.ServerNames {
		exist, err := checkFunc(tx, req.NodeClusterId, serverName, req.ExcludeServerId, req.SupportWildcard)
		if err != nil {
			return nil, err
		}
		if exist {
			duplicatedServerNames = append(duplicatedServerNames, serverName)
		}
	}

	return &pb.CheckServerNameDuplicationInNodeClusterResponse{DuplicatedServerNames: duplicatedServerNames}, nil
}

// FindLatestServers 查找最近访问的服务
func (this *ServerService) FindLatestServers(ctx context.Context, req *pb.FindLatestServersRequest) (*pb.FindLatestServersResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	servers, err := models.SharedServerDAO.FindLatestServers(tx, req.Size)
	if err != nil {
		return nil, err
	}
	pbServers := []*pb.Server{}
	for _, server := range servers {
		pbServers = append(pbServers, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
		})
	}
	return &pb.FindLatestServersResponse{Servers: pbServers}, nil
}

// FindNearbyServers 查找某个服务附近的服务
func (this *ServerService) FindNearbyServers(ctx context.Context, req *pb.FindNearbyServersRequest) (*pb.FindNearbyServersResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 查询服务的Group
	groupIds, err := models.SharedServerDAO.FindServerGroupIds(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if len(groupIds) > 0 {
		var pbGroups = []*pb.FindNearbyServersResponse_GroupInfo{}
		for _, groupId := range groupIds {
			group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if group == nil {
				continue
			}

			var pbGroup = &pb.FindNearbyServersResponse_GroupInfo{
				Name: group.Name,
			}
			servers, err := models.SharedServerDAO.FindNearbyServersInGroup(tx, groupId, req.ServerId, 10)
			if err != nil {
				return nil, err
			}
			for _, server := range servers {
				pbGroup.Servers = append(pbGroup.Servers, &pb.Server{
					Id:   int64(server.Id),
					Name: server.Name,
					IsOn: server.IsOn,
				})
			}
			pbGroups = append(pbGroups, pbGroup)
		}

		if len(pbGroups) > 0 {
			return &pb.FindNearbyServersResponse{
				Scope:  "group",
				Groups: pbGroups,
			}, nil
		}
	}

	// 集群
	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	servers, err := models.SharedServerDAO.FindNearbyServersInCluster(tx, clusterId, req.ServerId, 10)
	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return &pb.FindNearbyServersResponse{
			Scope:  "cluster",
			Groups: nil,
		}, nil
	}

	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, clusterId)
	if err != nil {
		return nil, err
	}
	var pbGroup = &pb.FindNearbyServersResponse_GroupInfo{
		Name: clusterName,
	}
	for _, server := range servers {
		pbGroup.Servers = append(pbGroup.Servers, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
			IsOn: server.IsOn,
		})
	}

	return &pb.FindNearbyServersResponse{
		Scope:  "cluster",
		Groups: []*pb.FindNearbyServersResponse_GroupInfo{pbGroup},
	}, nil
}

// PurgeServerCache 清除缓存
func (this *ServerService) PurgeServerCache(ctx context.Context, req *pb.PurgeServerCacheRequest) (*pb.PurgeServerCacheResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		// 检查是否为节点
		_, err = this.ValidateNode(ctx)
		if err != nil {
			return nil, err
		}
	}

	if len(req.Keys) == 0 && len(req.Prefixes) == 0 {
		return &pb.PurgeServerCacheResponse{IsOk: true}, nil
	}

	var purgeResponse = &pb.PurgeServerCacheResponse{}

	var tx = this.NullTx()

	var taskType = "purge"

	var tasks = []*pb.CreateHTTPCacheTaskRequest{}
	if len(req.Keys) > 0 {
		tasks = append(tasks, &pb.CreateHTTPCacheTaskRequest{
			Type:    taskType,
			KeyType: "key",
			Keys:    req.Keys,
		})
	}
	if len(req.Prefixes) > 0 {
		tasks = append(tasks, &pb.CreateHTTPCacheTaskRequest{
			Type:    taskType,
			KeyType: "prefix",
			Keys:    req.Prefixes,
		})
	}

	var domainMap = map[string]*models.Server{} // domain name => *Server

	for _, pbTask := range tasks {
		// 创建任务
		taskId, err := models.SharedHTTPCacheTaskDAO.CreateTask(tx, 0, pbTask.Type, pbTask.KeyType, "调用PURGE API")
		if err != nil {
			return nil, err
		}

		var countKeys = 0

		for _, key := range req.Keys {
			if len(key) == 0 {
				continue
			}

			// 获取域名
			var domain = utils.ParseDomainFromKey(key)
			if len(domain) == 0 {
				continue
			}

			// 查询所在集群
			server, ok := domainMap[domain]
			if !ok {
				server, err = models.SharedServerDAO.FindEnabledServerWithDomain(tx, domain)
				if err != nil {
					return nil, err
				}
				if server == nil {
					continue
				}
				domainMap[domain] = server
			}

			var serverClusterId = int64(server.ClusterId)
			if serverClusterId == 0 {
				continue
			}

			_, err = models.SharedHTTPCacheTaskKeyDAO.CreateKey(tx, taskId, key, pbTask.Type, pbTask.KeyType, serverClusterId)
			if err != nil {
				return nil, err
			}

			countKeys++
		}

		if countKeys == 0 {
			// 如果没有有效的Key，则直接完成
			err = models.SharedHTTPCacheTaskDAO.UpdateTaskStatus(tx, taskId, true, true)
		} else {
			err = models.SharedHTTPCacheTaskDAO.UpdateTaskReady(tx, taskId)
		}
		if err != nil {
			return nil, err
		}
	}

	purgeResponse.IsOk = true

	return purgeResponse, nil
}

// FindEnabledServerTrafficLimit 查找流量限制
func (this *ServerService) FindEnabledServerTrafficLimit(ctx context.Context, req *pb.FindEnabledServerTrafficLimitRequest) (*pb.FindEnabledServerTrafficLimitResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	// TODO 检查用户权限

	var tx = this.NullTx()
	limitConfig, err := models.SharedServerDAO.FindServerTrafficLimitConfig(tx, req.ServerId, nil)
	if err != nil {
		return nil, err
	}
	limitConfigJSON, err := json.Marshal(limitConfig)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledServerTrafficLimitResponse{
		TrafficLimitJSON: limitConfigJSON,
	}, nil
}

// UpdateServerTrafficLimit 设置流量限制
func (this *ServerService) UpdateServerTrafficLimit(ctx context.Context, req *pb.UpdateServerTrafficLimitRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var config = &serverconfigs.TrafficLimitConfig{}
	err = json.Unmarshal(req.TrafficLimitJSON, config)
	if err != nil {
		return nil, err
	}

	err = models.SharedServerDAO.UpdateServerTrafficLimitConfig(tx, req.ServerId, config)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateServerUserPlan 修改服务套餐
func (this *ServerService) UpdateServerUserPlan(ctx context.Context, req *pb.UpdateServerUserPlanRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		// 检查服务
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 检查套餐
	if req.UserPlanId < 0 {
		req.UserPlanId = 0
	}

	// 检查是否有变化
	oldUserPlanId, err := models.SharedServerDAO.FindServerUserPlanId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if req.UserPlanId == oldUserPlanId {
		return this.Success()
	}

	if req.UserPlanId > 0 {
		userId, err := models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
		if userId == 0 {
			return nil, errors.New("the server is not belong to any user")
		}

		userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, req.UserPlanId, nil)
		if err != nil {
			return nil, err
		}
		if userPlan == nil {
			return nil, errors.New("can not find user plan with id '" + types.String(req.UserPlanId) + "'")
		}
		if int64(userPlan.UserId) != userId {
			return nil, errors.New("can not find user plan with id '" + types.String(req.UserPlanId) + "'")
		}

		// 检查是否已经被别的服务所使用
		serverId, err := models.SharedServerDAO.FindEnabledServerIdWithUserPlanId(tx, req.UserPlanId)
		if err != nil {
			return nil, err
		}
		if serverId > 0 && serverId != req.ServerId {
			return nil, errors.New("the user plan is used by other server")
		}
	}

	err = models.SharedServerDAO.UpdateServerUserPlanId(tx, req.ServerId, req.UserPlanId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindServerUserPlan 获取服务套餐信息
func (this *ServerService) FindServerUserPlan(ctx context.Context, req *pb.FindServerUserPlanRequest) (*pb.FindServerUserPlanResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		// 检查服务
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	userPlanId, err := models.SharedServerDAO.FindServerUserPlanId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if userPlanId <= 0 {
		return &pb.FindServerUserPlanResponse{UserPlan: nil}, nil
	}

	userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, userPlanId, nil)
	if err != nil {
		return nil, err
	}
	if userPlan == nil {
		return &pb.FindServerUserPlanResponse{UserPlan: nil}, nil
	}

	plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return &pb.FindServerUserPlanResponse{UserPlan: nil}, nil
	}

	return &pb.FindServerUserPlanResponse{
		UserPlan: &pb.UserPlan{
			Id:     int64(userPlan.Id),
			UserId: int64(userPlan.UserId),
			PlanId: int64(userPlan.PlanId),
			Name:   userPlan.Name,
			IsOn:   userPlan.IsOn,
			DayTo:  userPlan.DayTo,
			User:   nil,
			Plan: &pb.Plan{
				Id:               int64(plan.Id),
				Name:             plan.Name,
				PriceType:        plan.PriceType,
				TrafficPriceJSON: plan.TrafficPrice,
				TrafficLimitJSON: plan.TrafficLimit,
			},
		},
	}, nil
}

// ComposeServerConfig 获取服务配置
func (this *ServerService) ComposeServerConfig(ctx context.Context, req *pb.ComposeServerConfigRequest) (*pb.ComposeServerConfigResponse, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	//读取节点的所有集群
	clusterIds, err := models.SharedNodeDAO.FindEnabledNodeClusterIds(tx, nodeId)
	if err != nil {
		return nil, err
	}

	// 读取服务所在集群
	serverClusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	// 如果不在当前节点的集群中，则返回nil
	if !lists.ContainsInt64(clusterIds, serverClusterId) {
		return &pb.ComposeServerConfigResponse{ServerConfigJSON: nil}, nil
	}

	serverConfig, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, req.ServerId, true)
	if err != nil {
		if err == models.ErrNotFound {
			return &pb.ComposeServerConfigResponse{ServerConfigJSON: nil}, nil
		}
		return nil, err
	}
	if serverConfig == nil {
		return &pb.ComposeServerConfigResponse{ServerConfigJSON: nil}, nil
	}

	configJSON, err := json.Marshal(serverConfig)
	if err != nil {
		return nil, err
	}
	return &pb.ComposeServerConfigResponse{ServerConfigJSON: configJSON}, nil
}

// ------- api 客户定制化接口

type CreateDetailedServerRequest struct {
	Username     string   `json:"username,omitempty"`    //指定用户账号
	ClusterId    int64    `json:"clusterId"`             // 指定集群
	Type         int32    `json:"type,omitempty"`        //回源主机类型：0/1/2 跟随CDN服务，跟随源站，自定义
	Host         string   `json:"host,omitempty"`        //回源主机名
	Domains      []string `json:"domains,omitempty"`     //域名列表
	Http         bool     `json:"http,omitempty"`        //http开关
	Https        bool     `json:"https,omitempty"`       //https开关
	Http2Enabled bool     `json:"http2Enabled"`          //启用http2
	CertIds      []int64  `json:"certIds,omitempty"`     //证书id列表
	OriginsJSON  string   `json:"originsJSON,omitempty"` //源站信息列表
	Name         string   `json:"name"`                  // 服务名称
	Address      []struct {
		Port     string `json:"port"`     //端口 0-65535
		Protocol string `json:"protocol"` // 协议 http/https
	}

	AccessLogIsOn  bool `json:"accessLogIsOn"`  //访问日志
	WebsocketIsOn  bool `json:"websocketIsOn"`  //websocket
	CacheIsOn      bool `json:"cacheIsOn"`      //缓存
	WafIsOn        bool `json:"wafIsOn"`        //waf
	RemoteAddrIsOn bool `json:"remoteAddrIsOn"` //从上级代理中读取IP
	StatIsOn       bool `json:"statIsOn"`       //统计
}

type UpdateServerWAFRequest struct {
	ServerId int64 `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"` //域名服务id
	Enable   bool  `protobuf:"varint,2,opt,name=enable,proto3" json:"enable,omitempty"`     //启用Web防火墙
}

type UpdateServerNamesAPIRequest struct {
	ServerId    int64    `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"`
	ServerNames []string `protobuf:"bytes,2,rep,name=serverNames,proto3" json:"serverNames,omitempty"`
}

type UpdateServerCertsAPIRequest struct {
	ServerId int64   `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"`
	CertIds  []int64 `protobuf:"varint,2,rep,packed,name=certIds,proto3" json:"certIds,omitempty"`
}

type FindServerDomainsRequest struct {
	ServerId int64 `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"` //服务id
}

type FindServerDomainsResponse struct {
	Domains []string `protobuf:"bytes,1,rep,name=domains,proto3" json:"domains,omitempty"` //域名数组
}
type origin struct {
	Id       int64  `json:"id"`       // 源站ID
	IP       string `json:"ip"`       // 源站IP
	Port     string `json:"port"`     // 源站端口
	Protocol string `json:"protocol"` // 源站协议
}
type FindServerOriginsResponse struct {
	Origins struct {
		Primary []origin `json:"primary"` //主要源站
		Backup  []origin `json:"backup"`  // 备用源站
	} `json:"origins,omitempty"` //源站数组
}

type UpdateDefensiveActionRequest struct {
	ServerId int64  `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"`
	Action   string `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"` //防御动作
}

type UpdateServerHTTPSRequest struct {
	ServerId              int64    `json:"serverId"`
	Https                 bool     `json:"https"`
	Http2                 bool     `json:"http2"`
	HstsOn                bool     `json:"hstsOn"`
	TlsVersion            string   `json:"tlsVersion"`
	Certs                 []int64  `json:"certs"`
	CACerts               []int64  `json:"caCerts"`
	CipherSuites          []string `json:"cipherSuites"`
	HstsMaxAge            int      `json:"hstsMaxAge"`
	HstsIncludeSubDomains int      `json:"hstsIncludeSubDomains"`
	HstsPreload           int      `json:"hstsPreload"`
	HstsDomains           []string `json:"hstsDomains"`
	OcspIsOn              bool     `json:"ocspIsOn"`
	ClientAuthType        int32    `json:"clientAuthType"`
	Ports                 []int    `json:"ports"`
}

// CreateDetailedServer 创建详细域名服务
func (this *ServerService) CreateDetailedServer(ctx context.Context, req *CreateDetailedServerRequest) (*pb.CreateServerResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	tx := this.NullTx()

	var userId int64
	//用户所属集群id
	var nodeClusterId int64
	var httpJSON, httpsJSON []byte
	var listenAddress []*serverconfigs.NetworkAddressConfig
	// 端口地址
	var httpConfig *serverconfigs.HTTPProtocolConfig = nil
	var httpsConfig *serverconfigs.HTTPSProtocolConfig = nil
	if !req.Http && !req.Https {
		return nil, fmt.Errorf("请设置http/https,不能同时为空")
	}
	//if len(req.Domains) == 0 {
	//	return nil, fmt.Errorf("域名列表不能为空")
	//}
	for k, v := range req.Domains {
		req.Domains[k] = strings.ToLower(v)
	}
	if req.Username == "" {
		if req.ClusterId == 0 {
			return nil, fmt.Errorf("username或者clusterId不能为空")
		}
		nodeClusterId = req.ClusterId
	} else {
		userId, err = models.SharedUserDAO.FindUserByUserName(tx, req.Username)
		if err != nil {
			return nil, fmt.Errorf("查询用户名%s失败：%s", req.Username, err.Error())
		}
		if userId == 0 {
			return nil, fmt.Errorf("该用户名%s不存在", req.Username)
		}
		// 集群
		nodeClusterId, err = models.SharedUserDAO.FindUserClusterId(tx, userId)
		if err != nil {
			return nil, err
		}
	}
	if req.Address != nil || len(req.Address) > 0 {
		for _, address := range req.Address {
			listenAddress = append(listenAddress, &serverconfigs.NetworkAddressConfig{Host: "", PortRange: address.Port, Protocol: serverconfigs.Protocol(address.Protocol)})
		}
	}
	if len(listenAddress) > 0 {
		for _, addr := range listenAddress {
			switch addr.Protocol.Primary() {
			case serverconfigs.ProtocolHTTP:
				if httpConfig == nil {
					httpConfig = &serverconfigs.HTTPProtocolConfig{
						BaseProtocol: serverconfigs.BaseProtocol{
							IsOn: true,
						},
					}
				}
				httpConfig.IsOn = true
				httpConfig.AddListen(addr)
			case serverconfigs.ProtocolHTTPS:
				if httpsConfig == nil {
					httpsConfig = &serverconfigs.HTTPSProtocolConfig{
						BaseProtocol: serverconfigs.BaseProtocol{
							IsOn: true,
						},
					}
				}
				httpsConfig.IsOn = true
				httpsConfig.AddListen(addr)
			default:
				return nil, fmt.Errorf("请输入正确的绑定端口协议类型（http/https）")
			}
		}
		// 开始保存
		httpJSON, err = httpConfig.AsJSON()
		if err != nil {
			return nil, err
		}

	} else {
		// HTTP
		if req.Http {
			httpConfig = &serverconfigs.HTTPProtocolConfig{
				BaseProtocol: serverconfigs.BaseProtocol{
					IsOn: true,
				}}
			httpConfig.IsOn = true
			httpConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  serverconfigs.ProtocolHTTP,
					Host:      "",
					PortRange: "80",
				},
			}
			// 开始保存
			httpJSON, err = httpConfig.AsJSON()
			if err != nil {
				return nil, err
			}
		}

		// HTTPS
		if req.Https {
			httpsConfig = &serverconfigs.HTTPSProtocolConfig{}
			httpsConfig.IsOn = true
			httpsConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  serverconfigs.ProtocolHTTPS,
					Host:      "",
					PortRange: "443",
				},
			}

		} else {

			httpsConfig = &serverconfigs.HTTPSProtocolConfig{}
			httpsConfig.IsOn = false
			httpsConfig.Listen = []*serverconfigs.NetworkAddressConfig{}

		}
	}
	// HTTP
	if !req.Http {
		httpJSON = nil
	}
	// HTTPS
	if req.Https {

		//if len(req.CertIds) == 0 {
		//	return nil, fmt.Errorf("请选择或者上传HTTPS证书")
		//}

		certRefs := []*sslconfigs.SSLCertRef{}
		for _, certId := range req.CertIds {
			certRefs = append(certRefs, &sslconfigs.SSLCertRef{
				IsOn:   true,
				CertId: certId,
			})
		}
		certRefsJSON, err := json.Marshal(certRefs)
		if err != nil {
			return nil, err
		}

		// 创建策略
		sslPolicyId, err := models.SharedSSLPolicyDAO.
			CreatePolicy(tx, adminId, userId, req.Http2Enabled, "TLS 1.2", certRefsJSON, nil, false, 0, nil, false, nil)
		if err != nil {
			return nil, err
		}
		httpsConfig.SSLPolicyRef = &sslconfigs.SSLPolicyRef{
			IsOn:        true,
			SSLPolicyId: sslPolicyId,
		}
	} else {
		if httpsConfig == nil {
			httpsConfig = &serverconfigs.HTTPSProtocolConfig{
				BaseProtocol: serverconfigs.BaseProtocol{
					IsOn: false,
				},
			}
		}
		// 创建策略
		sslPolicyId, err := models.SharedSSLPolicyDAO.
			CreatePolicy(tx, adminId, userId, false, "TLS 1.2", []byte("[]"), nil, false, 0, []byte("[]"), false, nil)
		if err != nil {
			return nil, err
		}
		httpsConfig.SSLPolicyRef = &sslconfigs.SSLPolicyRef{
			IsOn:        true,
			SSLPolicyId: sslPolicyId,
		}
	}

	httpsJSON, err = httpsConfig.AsJSON()
	if err != nil {
		return nil, err
	}
	serverNames := []*serverconfigs.ServerNameConfig{}
	for _, domainName := range req.Domains {
		if !domainutils.ValidateDomainFormat(domainName) {
			return nil, fmt.Errorf("域名'" + domainName + "'输入错误")
		}
		serverNames = append(serverNames, &serverconfigs.ServerNameConfig{
			Name:     domainName,
			Type:     "",
			SubNames: nil,
		})
	}
	serverconfigs.NormalizeServerNames(serverNames)
	var duplicatedServerNames []string
	// 检查域名是否已经存在
	for _, serverName := range req.Domains {
		exist, err := models.SharedServerDAO.ExistServerNameInCluster(tx, nodeClusterId, serverName, 0, false)
		if err != nil {
			return nil, err
		}
		if exist {
			duplicatedServerNames = append(duplicatedServerNames, serverName)
		}
	}
	if len(duplicatedServerNames) > 0 {
		return nil, fmt.Errorf(strings.Join(duplicatedServerNames, ", ") + " 已经被其他服务所占用，不能重复使用")
	}

	serverNamesJSON, err := json.Marshal(serverNames)
	if err != nil {
		return nil, err
	}
	// 是否需要审核
	isAuditing := false
	auditingServerNamesJSON := []byte("[]")
	if userId > 0 {
		// 如果域名不为空的时候需要审核
		if len(serverNamesJSON) > 0 && string(serverNamesJSON) != "[]" {
			globalConfig, err := models.SharedSysSettingDAO.ReadGlobalConfig(tx)
			if err != nil {
				return nil, err
			}
			if globalConfig != nil && globalConfig.HTTPAll.DomainAuditingIsOn {
				isAuditing = true
				auditingServerNamesJSON = serverNamesJSON
				serverNamesJSON = []byte("[]")
			}
		}
	}

	// 源站信息
	originMaps := []maps.Map{}
	if len(req.OriginsJSON) == 0 {
		return nil, fmt.Errorf("请输入源站信息")
	}
	err = json.Unmarshal([]byte(req.OriginsJSON), &originMaps)
	if err != nil {
		return nil, err
	}
	if len(originMaps) == 0 {
		return nil, fmt.Errorf("请输入源站信息")
	}
	primaryOriginRefs := []*serverconfigs.OriginRef{}
	backupOriginRefs := []*serverconfigs.OriginRef{}
	for _, originMap := range originMaps {
		host := originMap.GetString("host")
		isPrimary := originMap.GetBool("isPrimary")
		scheme := originMap.GetString("scheme")

		if len(host) == 0 {
			return nil, fmt.Errorf("源站地址不能为空")
		}
		if strings.Index(host, ":") > 0 {
			_, _, err := net.SplitHostPort(host)
			if err != nil {
				return nil, fmt.Errorf("源站地址'" + host + "'格式错误'")
			}
		} else if !domainutils.ValidateDomainFormat(host) {
			return nil, fmt.Errorf("源站地址'" + host + "'格式错误")
		}

		if scheme != "http" && scheme != "https" {
			return nil, fmt.Errorf("错误的源站协议")
		}

		addrHost, addrPort, err := net.SplitHostPort(host)
		if err != nil {
			addrHost = host
			if scheme == "http" {
				addrPort = "80"
			} else if scheme == "https" {
				addrPort = "443"
			}
		}
		addrMap := maps.Map{
			"protocol":  scheme,
			"portRange": addrPort,
			"host":      addrHost,
		}

		originId, err := models.SharedOriginDAO.CreateOrigin(tx, 0, userId, "", string(addrMap.AsJSON()), "", 10, true, nil, nil, nil, 0, 0, nil, nil, "", false)
		if err != nil {
			return nil, err
		}
		if isPrimary {
			primaryOriginRefs = append(primaryOriginRefs, &serverconfigs.OriginRef{
				IsOn:     true,
				OriginId: originId,
			})
		} else {
			backupOriginRefs = append(backupOriginRefs, &serverconfigs.OriginRef{
				IsOn:     true,
				OriginId: originId,
			})
		}
	}
	primaryOriginsJSON, err := json.Marshal(primaryOriginRefs)
	if err != nil {
		return nil, err
	}

	backupOriginsJSON, err := json.Marshal(backupOriginRefs)
	if err != nil {
		return nil, err
	}

	scheduling := &serverconfigs.SchedulingConfig{
		Code:    "random",
		Options: nil,
	}
	schedulingJSON, err := json.Marshal(scheduling)
	if err != nil {
		return nil, err
	}

	// 反向代理
	reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, userId, schedulingJSON, primaryOriginsJSON, backupOriginsJSON)
	if err != nil {
		return nil, err
	}
	reverseProxyRef := &serverconfigs.ReverseProxyRef{
		IsPrior:        false,
		IsOn:           true,
		ReverseProxyId: reverseProxyId,
	}
	reverseProxyRefJSON, err := json.Marshal(reverseProxyRef)
	if err != nil {
		return nil, err
	}
	if req.Type > 0 {
		//todo ::: ====== requestHostExcludingPort
		err = models.SharedReverseProxyDAO.UpdateReverseProxy(tx, reverseProxyId, int8(req.Type), req.Host, false, "", "", false, nil, nil, nil, nil, 0, 0, nil, false)
		if err != nil {
			return nil, err
		}
	}
	servername := req.Name
	if servername == "" {
		servername = req.Domains[0]
	}
	serverId, err := models.SharedServerDAO.CreateServer(tx, adminId, userId, serverconfigs.ServerTypeHTTPProxy, servername,
		"", serverNamesJSON, isAuditing, auditingServerNamesJSON, httpJSON, httpsJSON, nil, nil, nil, nil,
		0, reverseProxyRefJSON, nodeClusterId, nil, nil, nil, 0)
	if err != nil {
		return nil, err
	}

	// 开启访问日志和Websocket
	webId, err := models.SharedServerDAO.FindServerWebId(tx, serverId)
	if err != nil {
		return nil, err
	}
	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, serverId)
		if err != nil {
			return nil, err
		}
	}
	webConfig, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	} else {
		// 访问日志
		if req.AccessLogIsOn {
			err = models.SharedHTTPWebDAO.UpdateWebAccessLogConfig(tx,
				webConfig.Id,
				[]byte(`{
			"isPrior": false,
			"isOn": true,
			"fields": [1, 2, 6, 7],
			"status1": true,
			"status2": true,
			"status3": true,
			"status4": true,
			"status5": true,

			"storageOnly": false,
			"storagePolicies": [],

            "firewallOnly": false
		}`))

			if err != nil {
				return nil, err
			}
		}

		// websocket
		if req.WebsocketIsOn {
			websocketId, err := models.SharedHTTPWebsocketDAO.CreateWebsocket(tx,
				[]byte(`{
					"count": 30,
					"unit": "second"
				}`),
				true,
				nil,
				true,
				"",
			)
			if err != nil {
				return nil, err
			} else {
				err = models.SharedHTTPWebDAO.UpdateWebsocket(tx,
					webConfig.Id,
					[]byte(` {
				"isPrior": false,
				"isOn": true,
				"websocketId": `+types.String(websocketId)+`
			}`))
				if err != nil {
					return nil, err
				}
			}
		}

		// cache
		if req.CacheIsOn {
			var cacheConfig = &serverconfigs.HTTPCacheConfig{
				IsPrior:         false,
				IsOn:            true,
				AddStatusHeader: true,
				PurgeIsOn:       false,
				PurgeKey:        "",
				CacheRefs:       []*serverconfigs.HTTPCacheRef{},
			}
			cacheConfigJSON, _ := json.Marshal(cacheConfig)

			err = models.SharedHTTPWebDAO.UpdateWebCache(tx,
				webConfig.Id,
				cacheConfigJSON,
			)
			if err != nil {
				return nil, err
			}
		}

		// waf
		if req.WafIsOn {
			var firewallRef = &firewallconfigs.HTTPFirewallRef{
				IsPrior:          false,
				IsOn:             true,
				FirewallPolicyId: 0,
			}
			firewallRefJSON, _ := json.Marshal(firewallRef)
			err = models.SharedHTTPWebDAO.UpdateWebFirewall(tx,
				webConfig.Id,
				firewallRefJSON,
			)
			if err != nil {
				return nil, err
			}
		}

		// remoteAddr
		var remoteAddrConfig = &serverconfigs.HTTPRemoteAddrConfig{
			IsOn:  true,
			Value: "${rawRemoteAddr}",
		}
		if req.RemoteAddrIsOn {
			remoteAddrConfig.Value = "${remoteAddr}"
		}
		remoteAddrConfigJSON, _ := json.Marshal(remoteAddrConfig)
		err = models.SharedHTTPWebDAO.UpdateWebRemoteAddr(tx,
			webConfig.Id,
			remoteAddrConfigJSON,
		)
		if err != nil {
			return nil, err
		}

		// 统计
		if req.StatIsOn {
			var statConfig = &serverconfigs.HTTPStatRef{
				IsPrior: false,
				IsOn:    true,
			}
			statJSON, _ := json.Marshal(statConfig)

			err = models.SharedHTTPWebDAO.UpdateWebStat(tx,
				webConfig.Id,
				statJSON,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

// UpdateServerWAF 修改服务WAF状态
func (this *ServerService) UpdateServerWAF(ctx context.Context, req *UpdateServerWAFRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	// 当前的Server独立设置
	if config.FirewallRef == nil || config.FirewallRef.FirewallPolicyId == 0 {
		firewallPolicyId, err := dao.SharedHTTPWebDAO.InitEmptyHTTPFirewallPolicy(ctx, 0, req.ServerId, config.Id, config.FirewallRef != nil && config.FirewallRef.IsOn)
		if err != nil {
			return nil, err
		}
		config.FirewallRef = &firewallconfigs.HTTPFirewallRef{FirewallPolicyId: firewallPolicyId}
	}
	config.FirewallRef.IsOn = req.Enable
	firewallJSON, err := json.Marshal(config.FirewallRef)
	if err != nil {
		return nil, err
	}
	err = models.SharedHTTPWebDAO.UpdateWebFirewall(tx, config.Id, firewallJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerHTTPSAPI 修改服务HTTPS配置
func (this *ServerService) UpdateServerHTTPSAPI(ctx context.Context, req *UpdateServerHTTPSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	var httpsConfig = &serverconfigs.HTTPSProtocolConfig{}
	server, err := models.SharedServerDAO.FindEnabledServerBasic(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if req.TlsVersion == "" {
		req.TlsVersion = "TLS 1.1"
	}
	if len(server.Https) > 0 {
		err := json.Unmarshal(server.Https, httpsConfig)
		if err != nil {
			return nil, err
		}
	} else {
		httpsConfig.IsOn = true
	}
	if len(req.Ports) > 0 {
		httpsConfig.Listen = []*serverconfigs.NetworkAddressConfig{}
		for _, v := range req.Ports {
			if v < 0 || v > 65535 {
				return nil, fmt.Errorf("无效的端口(0~65535)")
			}
			httpsConfig.Listen = append(httpsConfig.Listen, &serverconfigs.NetworkAddressConfig{PortRange: strconv.Itoa(v), Protocol: "https"})
		}
	}
	certRefs := []*sslconfigs.SSLCertRef{}
	for _, certId := range req.Certs {
		certRefs = append(certRefs, &sslconfigs.SSLCertRef{
			IsOn:   true,
			CertId: certId,
		})
	}
	certRefsJSON, err := json.Marshal(certRefs)
	if err != nil {
		return nil, err
	}
	certCARefs := []*sslconfigs.SSLCertRef{}
	for _, certId := range req.CACerts {
		certCARefs = append(certRefs, &sslconfigs.SSLCertRef{
			IsOn:   true,
			CertId: certId,
		})
	}
	certCARefsJSON, err := json.Marshal(certCARefs)
	if err != nil {
		return nil, err
	}
	hsts := &sslconfigs.HSTSConfig{IsOn: req.HstsOn, MaxAge: req.HstsMaxAge, IncludeSubDomains: req.HstsIncludeSubDomains == 1, Preload: req.HstsPreload == 1, Domains: req.HstsDomains}
	hstsJSON, err := json.Marshal(hsts)
	if err != nil {
		return nil, err
	}
	var sslPolicyId int64
	if httpsConfig.SSLPolicyRef != nil && httpsConfig.SSLPolicyRef.SSLPolicyId > 0 {
		config, err := models.SharedSSLPolicyDAO.ComposePolicyConfig(tx, httpsConfig.SSLPolicyRef.SSLPolicyId, nil)
		if err != nil {
			return nil, err
		}
		config.IsOn = req.Https
		config.HTTP2Enabled = req.Http2
		config.MinVersion = req.TlsVersion
		config.OCSPIsOn = req.OcspIsOn
		config.ClientAuthType = sslconfigs.SSLClientAuthType(req.ClientAuthType)
		config.CipherSuitesIsOn = len(req.CipherSuites) > 0
		config.CipherSuites = req.CipherSuites
		config.CertRefs = certRefs
		config.ClientCARefs = certCARefs
		config.HSTS = hsts
		err = models.SharedSSLPolicyDAO.UpdatePolicy(tx, config.Id, req.Http2, req.TlsVersion, certRefsJSON, hstsJSON, req.OcspIsOn, req.ClientAuthType, certCARefsJSON, len(req.CipherSuites) > 0, req.CipherSuites)
		if err != nil {
			return nil, err
		}
		sslPolicyId = config.Id
	} else {
		sslPolicyId, err = models.SharedSSLPolicyDAO.
			CreatePolicy(tx, adminId, int64(server.UserId), req.Http2, req.TlsVersion, certRefsJSON, hstsJSON, req.OcspIsOn, req.ClientAuthType, certCARefsJSON, len(req.CipherSuites) > 0, req.CipherSuites)
		if err != nil {
			return nil, err
		}
	}
	httpsConfig.SSLPolicyRef = &sslconfigs.SSLPolicyRef{
		IsOn:        req.Https,
		SSLPolicyId: sslPolicyId,
	}
	configData, err := json.Marshal(httpsConfig)
	if err != nil {
		return nil, err
	}
	err = models.SharedServerDAO.UpdateServerHTTPS(tx, req.ServerId, configData)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateServerNamesAPI 修改域名服务及其新增对应证书
func (this *ServerService) UpdateServerNamesAPI(ctx context.Context, req *UpdateServerNamesAPIRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	if req.ServerId == 0 {
		return nil, fmt.Errorf("请输入服务id")
	}
	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, fmt.Errorf("该域名服务不存在")
	}
	clusterId := server.ClusterId
	// 清空该服务绑定的所有域名
	if len(req.ServerNames) == 0 {
		// 修改配置
		err = models.SharedServerDAO.UpdateServerNames(tx, req.ServerId, nil)
		if err != nil {
			return nil, err
		}
		return this.Success()
	}
	var serverNames []*serverconfigs.ServerNameConfig
	for _, domainName := range req.ServerNames {
		if !domainutils.ValidateDomainFormat(domainName) {
			return nil, fmt.Errorf("域名'" + domainName + "'输入错误")
		}
		serverNames = append(serverNames, &serverconfigs.ServerNameConfig{
			Name:     domainName,
			Type:     "",
			SubNames: nil,
		})
	}
	// 检查域名是否已经存在
	allServerNames := serverconfigs.PlainServerNames(serverNames)

	duplicatedServerNames := []string{}
	if len(allServerNames) > 0 {
		for _, serverName := range allServerNames {
			exist, err := models.SharedServerDAO.ExistServerNameInCluster(tx, int64(clusterId), serverName, req.ServerId, false)
			if err != nil {
				return nil, err
			}
			if exist {
				duplicatedServerNames = append(duplicatedServerNames, serverName)
			}
		}
		if len(duplicatedServerNames) > 0 {
			return nil, fmt.Errorf("域名 " + strings.Join(duplicatedServerNames, ", ") + " 已经被其他服务所占用，不能重复使用")
		}
	}

	ServerNamesJSON, err := json.Marshal(serverNames)
	if err != nil {
		return nil, err
	}
	err = models.SharedServerDAO.UpdateAuditingServerNames(tx, req.ServerId, false, nil)
	if err != nil {
		return nil, err
	}
	// 修改配置
	err = models.SharedServerDAO.UpdateServerNames(tx, req.ServerId, ServerNamesJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

func (this *ServerService) UpdateServerCertsAPI(ctx context.Context, req *UpdateServerCertsAPIRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	if req.ServerId == 0 {
		return nil, fmt.Errorf("请输入服务id")
	}
	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, fmt.Errorf("该域名服务不存在")
	}

	if req.CertIds != nil {
		// HTTPS
		sslPolicy := &sslconfigs.SSLPolicy{}
		sslPolicyRef := &sslconfigs.SSLPolicyRef{}
		httpsConfig := &serverconfigs.HTTPSProtocolConfig{}
		if len(server.Https) > 0 && server.Https != nil {
			err := json.Unmarshal(server.Https, httpsConfig)
			if err != nil {
				return nil, err
			}
			if httpsConfig.SSLPolicy != nil {
				sslPolicy = httpsConfig.SSLPolicy
			}
			if httpsConfig.SSLPolicyRef != nil {
				sslPolicyRef = httpsConfig.SSLPolicyRef
			}
			if httpsConfig.SSLPolicy == nil && sslPolicyRef != nil && sslPolicyRef.IsOn && sslPolicyRef.SSLPolicyId > 0 {
				policy, err := models.SharedSSLPolicyDAO.FindEnabledSSLPolicy(tx, sslPolicyRef.SSLPolicyId)
				if err != nil {
					return nil, err
				}
				sslPolicy.Id = int64(policy.Id)
				sslPolicy.IsOn = policy.IsOn
				sslPolicy.ClientAuthType = int(policy.ClientAuthType)
				sslPolicy.HTTP2Enabled = policy.Http2Enabled == 1
				sslPolicy.MinVersion = policy.MinVersion

				// certs
				if models.IsNotNull(policy.Certs) {
					refs := []*sslconfigs.SSLCertRef{}
					err = json.Unmarshal(policy.Certs, &refs)
					if err != nil {
						return nil, err
					}
					if len(refs) > 0 {
						for _, ref := range refs {
							certConfig, err := models.SharedSSLCertDAO.ComposeCertConfig(tx, ref.CertId, nil)
							if err != nil {
								return nil, err
							}
							if certConfig == nil {
								continue
							}
							sslPolicy.CertRefs = append(sslPolicy.CertRefs, ref)
							sslPolicy.Certs = append(sslPolicy.Certs, certConfig)
						}
					}
				}

				// client CA certs
				if models.IsNotNull(policy.ClientCACerts) {
					refs := []*sslconfigs.SSLCertRef{}
					err = json.Unmarshal(policy.ClientCACerts, &refs)
					if err != nil {
						return nil, err
					}
					if len(refs) > 0 {
						for _, ref := range refs {
							certConfig, err := models.SharedSSLCertDAO.ComposeCertConfig(tx, ref.CertId, nil)
							if err != nil {
								return nil, err
							}
							if certConfig == nil {
								continue
							}
							sslPolicy.ClientCARefs = append(sslPolicy.ClientCARefs, ref)
							sslPolicy.ClientCACerts = append(sslPolicy.ClientCACerts, certConfig)
						}
					} else {
						sslPolicy.ClientCACerts = []*sslconfigs.SSLCertConfig{}
					}
				}

				// cipher suites
				sslPolicy.CipherSuitesIsOn = policy.CipherSuitesIsOn == 1
				if models.IsNotNull(policy.CipherSuites) {
					cipherSuites := []string{}
					err = json.Unmarshal(policy.CipherSuites, &cipherSuites)
					if err != nil {
						return nil, err
					}
					sslPolicy.CipherSuites = cipherSuites
				}

				// hsts
				if models.IsNotNull(policy.Hsts) {
					hstsConfig := &sslconfigs.HSTSConfig{}
					err = json.Unmarshal(policy.Hsts, hstsConfig)
					if err != nil {
						return nil, err
					}
					sslPolicy.HSTS = hstsConfig
				}
			}
		}
		sslPolicy.CertRefs = []*sslconfigs.SSLCertRef{}
		for _, id := range req.CertIds {
			sslPolicy.CertRefs = append(sslPolicy.CertRefs, &sslconfigs.SSLCertRef{IsOn: true, CertId: id})
		}

		sslPolicyId := sslPolicy.Id

		certsJSON, err := json.Marshal(sslPolicy.CertRefs)
		if err != nil {
			return nil, err
		}

		if sslPolicyId > 0 {

			var hstsJSON []byte
			if sslPolicy.HSTS != nil {
				hstsJSON, err = json.Marshal(sslPolicy.HSTS)
				if err != nil {
					return nil, err
				}
			} else {
				hstsJSON = nil
			}

			clientCACertsJSON, err := json.Marshal(sslPolicy.ClientCACerts)
			if err != nil {
				return nil, err
			}
			err = models.SharedSSLPolicyDAO.UpdatePolicy(tx, sslPolicyId, sslPolicy.HTTP2Enabled, sslPolicy.MinVersion,
				certsJSON, hstsJSON, sslPolicy.OCSPIsOn, types.Int32(sslPolicy.ClientAuthType),
				clientCACertsJSON, sslPolicy.CipherSuitesIsOn, sslPolicy.CipherSuites)
			if err != nil {
				return nil, err
			}
		} else {
			sslPolicyId, err = models.SharedSSLPolicyDAO.CreatePolicy(tx, adminId, 0,
				true,
				"TLS 1.1",
				certsJSON,
				nil,
				sslPolicy.OCSPIsOn,
				types.Int32(sslPolicy.ClientAuthType),
				[]byte("[]"),
				false,
				[]string{})
			if err != nil {
				return nil, err
			}
			httpsConfig.SSLPolicyRef = &sslconfigs.SSLPolicyRef{
				IsOn:        true,
				SSLPolicyId: sslPolicyId,
			}
			httpsConfig.IsOn = true
			configData, err := json.Marshal(httpsConfig)
			if err != nil {
				return nil, err
			}
			err = models.SharedServerDAO.UpdateServerHTTPS(tx, req.ServerId, configData)
			if err != nil {
				return nil, err
			}
		}
	}
	return this.Success()
}

// UpdateBasicServer 修改服务基本信息
func (this *ServerService) UpdateBasicServer(ctx context.Context, req *pb.UpdateServerBasicRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}
	groupIds := []int64{}
	err = json.Unmarshal(server.GroupIds, &groupIds)
	if err != nil {
		return nil, errors.New("parse server config error:" + err.Error())
	}
	if req.Description != "" {
		server.Description = req.Description
	}
	err = models.SharedServerDAO.UpdateServerBasic(tx, req.ServerId, req.Name, server.Description, req.NodeClusterId, req.KeepOldConfigs, server.IsOn, groupIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

type FindAllServerResponse struct {
	Servers []*findAllServerResponse `json:"servers"`
	Total   int64                    `json:"total"`
}
type findAllServerResponse struct {
	Id             int64    `json:"id,omitempty"`              //服务id
	Name           string   `json:"name,omitempty"`            //服务名称
	Description    string   ` json:"description,omitempty"`    //服务描述
	Domains        []string `json:"domains,omitempty"`         //服务包含域名数组
	IsOn           bool     ` json:"isOn,omitempty"`           //是否开启
	IsAuditing     bool     `json:"isAuditing,omitempty"`      //是否正在审核
	Cname          string   ` json:"cname,omitempty"`          //cname
	PrimaryOrigins []string ` json:"primaryOrigins,omitempty"` //主源站列表
	BackupOrigins  []string `json:"backupOrigins,omitempty"`   //备源站列表
	ClusterId      uint32   ` json:"clusterId,omitempty"`      //集群ID
}

// FindAllServer 查询所有域名服务
func (this *ServerService) FindAllServer(ctx context.Context, req *pb.ListEnabledServersMatchRequest) (*FindAllServerResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	total, err := models.SharedServerDAO.CountAllEnabledServersMatch(tx, 0, req.Keyword, 0, 0, 0, []string{"http"})
	if err != nil {
		return nil, err
	}

	if req.Size == 0 {
		req.Size = total
	}
	servers, err := models.SharedServerDAO.ListEnabledServersMatch(tx, req.Offset, req.Size, 0, req.Keyword,
		0, 0, 0, []string{"http"}, "")
	if err != nil {
		return nil, err
	}
	var result []*findAllServerResponse
	for _, server := range servers {
		var domains []string
		var serverNamesConfig []*serverconfigs.ServerNameConfig
		if !server.IsAuditing { //审核结束 显示
			err = json.Unmarshal(server.ServerNames, &serverNamesConfig)
		} else { //正在审核 显示
			err = json.Unmarshal(server.AuditingServerNames, &serverNamesConfig)
		}
		if err != nil {
			return nil, err
		} else {
			for _, domain := range serverNamesConfig {
				if domain.Name != "" {
					domains = append(domains, domain.Name)
				}
				domains = append(domains, domain.SubNames...)
			}
		}
		//cname
		var cname string
		cname = server.DnsName + "."
		if server.ClusterId > 0 {
			clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, int64(server.ClusterId), nil)
			if err != nil {
				return nil, err
			}
			if clusterDNS != nil {
				domainId := int64(clusterDNS.DnsDomainId)
				if domainId > 0 {
					domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId, nil)
					if err != nil {
						return nil, err
					}
					if domain != nil {
						cname += domain.Name + "."
					}
				}
			}
		}
		//源站列表
		var primaryOrigins, backupOrigins []string

		// ReverseProxy
		if models.IsNotNull(server.ReverseProxy) {
			reverseProxyRef := &serverconfigs.ReverseProxyRef{}
			err := json.Unmarshal(server.ReverseProxy, reverseProxyRef)
			if err != nil {
				return nil, err
			}

			reverseProxyConfig, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, nil)
			if err != nil {
				return nil, err
			}
			if reverseProxyConfig != nil {
				for _, origin := range reverseProxyConfig.PrimaryOrigins {
					primaryOrigins = append(primaryOrigins, origin.Addr.Protocol.String()+"://"+origin.Addr.Host+":"+origin.Addr.PortRange)
				}
				for _, origin := range reverseProxyConfig.BackupOrigins {
					backupOrigins = append(backupOrigins, origin.Addr.Protocol.String()+"://"+origin.Addr.Host+":"+origin.Addr.PortRange)
				}
			}
		}
		result = append(result, &findAllServerResponse{
			Id:             int64(server.Id),
			IsOn:           server.IsOn,
			Name:           server.Name,
			Description:    server.Description,
			IsAuditing:     server.IsAuditing,
			Domains:        domains,
			Cname:          cname,
			PrimaryOrigins: primaryOrigins,
			BackupOrigins:  backupOrigins,
			ClusterId:      server.ClusterId,
		})
	}

	return &FindAllServerResponse{Servers: result, Total: total}, nil
}

// FindServerOrigins 查询服务的所有源站信息
func (this *ServerService) FindServerOrigins(ctx context.Context, req *FindServerDomainsRequest) (*FindServerOriginsResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	if req.ServerId <= 0 {
		return nil, fmt.Errorf("无效的服务id")
	}

	reverseProxyRef, err := models.SharedServerDAO.FindReverseProxyRef(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		return &FindServerOriginsResponse{}, nil
	}

	primary := []origin{}
	backup := []origin{}
	reverseProxyConfig, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, nil)
	if err != nil {
		return nil, err
	}

	for _, originConfig := range reverseProxyConfig.PrimaryOrigins {
		primary = append(primary, origin{Id: originConfig.Id, IP: originConfig.Addr.Host, Port: originConfig.Addr.PortRange, Protocol: string(originConfig.Addr.Protocol)})
	}
	for _, originConfig := range reverseProxyConfig.BackupOrigins {
		backup = append(backup, origin{Id: originConfig.Id, IP: originConfig.Addr.Host, Port: originConfig.Addr.PortRange, Protocol: string(originConfig.Addr.Protocol)})
	}
	resp := &FindServerOriginsResponse{}
	resp.Origins.Primary = primary
	resp.Origins.Backup = backup
	return resp, nil
}

func (this *ServerService) FindServerDomains(ctx context.Context, req *FindServerDomainsRequest) (*FindServerDomainsResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	if req.ServerId <= 0 {
		return nil, fmt.Errorf("无效的服务id")
	}
	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	domains := []string{}
	serverNamesConfig := []*serverconfigs.ServerNameConfig{}

	if !server.IsAuditing { //审核结束 显示
		err = json.Unmarshal(server.ServerNames, &serverNamesConfig)
	} else { //正在审核 显示
		err = json.Unmarshal(server.AuditingServerNames, &serverNamesConfig)
	}
	if err != nil {
		return nil, err
	} else {
		for _, domain := range serverNamesConfig {
			domains = append(domains, domain.Name)
			domains = append(domains, domain.SubNames...)
		}
	}

	return &FindServerDomainsResponse{Domains: domains}, nil
}

var defensiveActive = map[string]firewallconfigs.HTTPFirewallActionString{
	"BLACK":   firewallconfigs.HTTPFirewallActionBlock,
	"ALLOW":   firewallconfigs.HTTPFirewallActionAllow,
	"LOG":     firewallconfigs.HTTPFirewallActionLog,
	"CAPTCHA": firewallconfigs.HTTPFirewallActionCaptcha,
	"NOTIFY":  firewallconfigs.HTTPFirewallActionNotify,
	"GET302":  firewallconfigs.HTTPFirewallActionGet302,
	"POST307": firewallconfigs.HTTPFirewallActionPost307,
	"RECORD":  firewallconfigs.HTTPFirewallActionRecordIP,
	"DISABLE": "disable", //停用所有规则
}

func (this *ServerService) UpdateDefensiveAction(ctx context.Context, req *UpdateDefensiveActionRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	if req.ServerId == 0 {
		return nil, errors.New("serverId not be empty")
	}
	req.Action = strings.ToUpper(req.Action)
	actionOpt, ok := defensiveActive[req.Action]
	if !ok {
		return nil, errors.New("not support this action")
	}

	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	// 当前的Server独立设置
	if config.FirewallRef == nil || config.FirewallRef.FirewallPolicyId == 0 { //当前无配置策略 无独立的入站规则 直接退出
		return this.Success()
	}
	//列出这个策略下的入站规则分组
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, config.FirewallRef.FirewallPolicyId, nil)
	if err != nil {
		return nil, err
	}
	if firewallPolicy == nil {
		return nil, errors.New(" not found firewallPolicy")
	}
	for _, group := range firewallPolicy.Inbound.Groups {
		// 查询对应分组下的规则集
		groupConfig, err := models.SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(tx, group.Id)
		if err != nil {
			return nil, err
		}
		if groupConfig == nil {
			continue
		}

		for _, ruleSet := range groupConfig.Sets {
			//修改所有action
			err = this.createDefaultAction(actionOpt, ruleSet.Id)
			if err != nil {
				return nil, err
			}
		}
	}
	return this.Success()
}

func (this *ServerService) createDefaultAction(opt firewallconfigs.HTTPFirewallActionString, ruleSetId int64) error {
	actions := make([]firewallconfigs.HTTPFirewallActionConfig, 1)
	tx := this.NullTx()
	actions[0].Code = opt
	if opt == firewallconfigs.HTTPFirewallActionRecordIP {
		listsResp, err := models.SharedIPListDAO.ListEnabledIPLists(tx, "black",
			true,
			"公共黑名单",
			0,
			1,
		)
		if err != nil || len(listsResp) == 0 {
			return errors.New("not found public black ip list")
		}
		actions[0].Options = maps.Map{
			"type": "black", "level": "critical", "timeout": 0, "ipListId": listsResp[0].Id, "ipListName": "公共黑名单",
		}
	}
	actionJSON, err := json.Marshal(actions)
	if err != nil {
		return err
	}
	return models.SharedHTTPFirewallRuleSetDAO.UpdateRuleSetAction(tx, ruleSetId, actionJSON)
}

type UpdateDefensiveEnableRequest struct {
	ServerId int64 `json:"serverId"`
	GroupId  int64 `json:"groupId"`
	Enable   bool  `json:"enable"` //启用
}

func (this *ServerService) UpdateDefensiveEnable(ctx context.Context, req *UpdateDefensiveEnableRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	if req.ServerId == 0 {
		return nil, errors.New("serverId not be empty")
	}

	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	// 当前的Server独立设置
	if config.FirewallRef == nil || config.FirewallRef.FirewallPolicyId == 0 { //当前无配置策略 无独立的入站规则 直接退出
		return this.Success()
	}
	//列出这个策略下的入站规则分组
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, config.FirewallRef.FirewallPolicyId, nil)
	if err != nil {
		return nil, err
	}
	if firewallPolicy == nil {
		return nil, errors.New(" not found firewallPolicy")
	}
	for _, group := range firewallPolicy.Inbound.Groups {
		if req.GroupId == 0 || group.Id == req.GroupId {
			// 启用或者禁用所有分组
			err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(tx, group.Id, req.Enable)

			if err != nil {
				return nil, err
			}
		}
	}
	return this.Success()
}

type CreateInboundGroupRequest struct {
	ServerId int64  `json:"serverId"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Desc     string `json:"desc"`
	IsOn     bool   `json:"isOn"`
}
type CreateInboundGroupResponse struct {
	GroupId int64 `json:"groupId"`
}

// CreateInboundGroup 新增入站规则分组
func (this *ServerService) CreateInboundGroup(ctx context.Context, req *CreateInboundGroupRequest) (*CreateInboundGroupResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	var firewallPolicyId int64
	if req.ServerId == 0 {
		return nil, errors.New("serverId not be empty")
	}

	// 开启访问日志和Websocket
	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}
	webConfig, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, errors.New("find web config error:" + err.Error())
	}

	if webConfig.FirewallRef == nil || webConfig.FirewallRef.FirewallPolicyId == 0 {

		policyId, err := models.SharedHTTPFirewallPolicyDAO.CreateFirewallPolicy(tx, 0, 0, req.ServerId, req.IsOn, req.Name, "", nil, nil)
		if err != nil {
			return nil, err
		}

		// 初始化
		inboundConfig := &firewallconfigs.HTTPFirewallInboundConfig{IsOn: true}
		outboundConfig := &firewallconfigs.HTTPFirewallOutboundConfig{IsOn: true}

		inboundConfigJSON, err := json.Marshal(inboundConfig)
		if err != nil {
			return nil, err
		}

		outboundConfigJSON, err := json.Marshal(outboundConfig)
		if err != nil {
			return nil, err
		}

		err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(tx, policyId, inboundConfigJSON, outboundConfigJSON, false)
		if err != nil {
			return nil, err
		}

		firewallRef := &firewallconfigs.HTTPFirewallRef{
			IsPrior:          false,
			IsOn:             webConfig.FirewallRef != nil && webConfig.FirewallRef.IsOn,
			FirewallPolicyId: policyId,
		}
		firewallRefJSON, err := json.Marshal(firewallRef)
		if err != nil {
			return nil, errors.New("init empty http  firewall policy error:" + err.Error())
		}
		err = models.SharedHTTPWebDAO.UpdateWebFirewall(tx, webConfig.Id, firewallRefJSON)
		if err != nil {
			return nil, errors.New("init empty http  firewall policy error:" + err.Error())
		}
		firewallPolicyId = policyId
	} else {
		firewallPolicyId = webConfig.FirewallPolicy.Id
	}
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, firewallPolicyId, nil)
	if err != nil {
		return nil, errors.New("find http  firewall policy config error:" + err.Error())
	}
	groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroup(tx, req.IsOn, req.Name, req.Code, req.Desc)
	if err != nil {
		return nil, err
	}
	firewallPolicy.Inbound.GroupRefs = append(firewallPolicy.Inbound.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
		IsOn:    true,
		GroupId: groupId,
	})

	inboundJSON, _ := firewallPolicy.InboundJSON()
	outboundJSON, _ := firewallPolicy.OutboundJSON()
	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(tx, firewallPolicyId, inboundJSON, outboundJSON, true)
	if err != nil {
		return nil, err
	}

	return &CreateInboundGroupResponse{GroupId: groupId}, nil
}

var redirectRuleConditionOptions = map[string][]string{
	"url-extension":         {"${requestPathExtension}", "in", "URL 扩展名", "根据URL中的文件路径扩展名进行过滤"},
	"url-prefix":            {"${requestPath}", "prefix", "URL 前端", "根据URL中的文件路径前缀进行过滤"},
	"url-eq":                {"${requestPath}", "eq", "URL 精准匹配", "检查URL中的文件路径是否一致"},
	"url-regexp":            {"${requestPath}", "regexp", "URL 正则匹配", "使用正则表达式检查URL中的文件路径是否一致"},
	"url-agent-regexp":      {"${userAgent}", "regexp", "User-Agent 正则匹配", "使用正则表达式检查User-Agent中是否含有某些浏览器和系统标识"},
	"params":                {"${userAgent}", "regexp", "参数匹配", "根据参数值进行匹配"},
	"url-not-extension":     {"${requestPathExtension}", "not in", "排除：URL 扩展名", "根据URL中的文件路径扩展名进行过滤"},
	"url-not-prefix":        {"${requestPath}", "prefix", "排除：URL 前缀", "根据URL中的文件路径前缀进行过滤"}, //"isReverse":true,
	"url-not-eq":            {"${requestPath}", "eq", "排除：URL 精准匹配", "检查URL中的文件路径是否一致"},     //"isReverse":true,
	"url-not-regexp":        {"${requestPath}", "not regexp", "排除：URL 正则匹配", "使用正则表达式检查URL中的文件路径是否一致，如果一致，则不匹配"},
	"user-agent-not-regexp": {"${userAgent}", "not regexp", "排除：User-Agent 正则匹配", "使用正则表达式检查User-Agent中是否含有某些浏览器和系统标识，如果含有，则不匹配"},
	"mime-type":             {"${response.contentType}", "mime type", "内容 MimeType", "根据服务器返回的内容的MimeType进行过滤。注意：当用于缓存条件时，此条件需要结合别的请求条件使用"},
}
var redirectRuleAllConditions = []maps.Map{
	{
		"code":        "url-extension",
		"name":        "URL 扩展名",
		"description": "根据URL中的文件路径扩展名进行过滤",
	},
	{
		"code":        "url-prefix",
		"name":        "URL 前缀",
		"description": "根据URL中的文件路径前缀进行过滤",
	},
	{
		"code":        "url-eq",
		"name":        "URL 精准匹配",
		"description": "使用正则表达式检查URL中的文件路径是否一致",
	},
	{
		"code":        "url-regexp",
		"name":        "URL 正则匹配",
		"description": "使用正则表达式检查URL中的文件路径是否一致",
	},
	{
		"code":        "url-agent-regexp",
		"name":        "User-Agent 正则匹配",
		"description": "使用正则表达式检查User-Agent中是否含有某些浏览器和系统标识",
	},
	{
		"code":        "params",
		"name":        "参数匹配",
		"description": "根据参数值进行匹配",
	},
	{
		"code":        "url-not-extension",
		"name":        "排除：URL 扩展名",
		"description": "根据URL中的文件路径扩展名进行过滤",
	},
	{
		"code":        "url-not-prefix",
		"name":        "排除：URL 前缀",
		"description": "根据URL中的文件路径前缀进行过滤",
	},
	{
		"code":        "url-not-eq",
		"name":        "排除：URL 精准匹配",
		"description": "检查URL中的文件路径是否一致",
	},
	{
		"code":        "url-not-regexp",
		"name":        "排除：URL 正则匹配",
		"description": "使用正则表达式检查URL中的文件路径是否一致，如果一致，则不匹配",
	},
	{
		"code":        "user-agent-not-regexp",
		"name":        "排除：User-Agent 正则匹配",
		"description": "使用正则表达式检查User-Agent中是否含有某些浏览器和系统标识，如果含有，则不匹配",
	},
	{
		"code":        "mime-type",
		"name":        "内容 MimeType",
		"description": "根据服务器返回的内容的MimeType进行过滤。注意：当用于缓存条件时，此条件需要结合别的请求条件使用",
	},
}

type CreateRedirectRuleRequest struct {
	ServerId       int64  `json:"serverId"`
	BeforeURL      string `json:"beforeURL"`
	Mode           string `json:"mode"`
	AfterURL       string `json:"afterURL"`
	KeepArgs       bool   `json:"keepArgs"`
	KeepRequestURI bool   `json:"keepRequestURI"`
	StatusCode     int    `json:"statusCode"`
	IsON           bool   `json:"isOn"`
	Cond           struct {
		Connector string `json:"connector"` // or/and
		Group     []struct {
			IsON        bool   `json:"isOn"`
			Connector   string `json:"connector"` // or/and
			Description string `json:"description"`
			Conds       []struct {
				Type              string   `json:"type"`              //条件类型
				Values            []string `json:"values"`            //值列表
				IsCaseInsensitive bool     `json:"isCaseInsensitive"` // 是否区分大小写
				Operator          string   `json:"operator"`          //操作法 大于小于等
				Param             string   `json:"param"`             //参数
			} `json:"conds"`
		} `json:"groups"`
	} `json:"conds"`
}
type CreateRedirectRuleResponse struct {
	Id int `json:"id"`
}
type redirectRuleOperatorsItem struct {
	Name        string `json:"name"`
	Operator    string `json:"operator"`
	Description string `json:"description"`
}
type redirectRuleConditionsItem struct {
	Name        string `json:"name"`
	Code        string `json:"Code"`
	Description string `json:"description"`
}
type commonParametersItem struct {
	Param       string `json:"param"`
	Description string `json:"description"`
}
type RedirectRuleOperatorsResponse struct {
	Options []*redirectRuleOperatorsItem `json:"options"`
}

type RedirectRuleConditionsResponse struct {
	Options []*redirectRuleConditionsItem `json:"options"`
}
type CommonParametersResponse struct {
	Options []*commonParametersItem `json:"options"`
}

// RedirectRuleOperators 重定向操作符下拉框
func (this *ServerService) RedirectRuleOperators(ctx context.Context, req *pb.RPCSuccess) (*RedirectRuleOperatorsResponse, error) {
	result := &RedirectRuleOperatorsResponse{}
	for _, operator := range shared.AllRequestOperators() {
		result.Options = append(result.Options, &redirectRuleOperatorsItem{
			operator.GetString("name"),
			operator.GetString("op"),
			operator.GetString("description"),
		})
	}
	return result, nil
}

// RedirectRuleCondition 重定向条件类型下拉框
func (this *ServerService) RedirectRuleCondition(ctx context.Context, req *pb.RPCSuccess) (*RedirectRuleConditionsResponse, error) {
	result := &RedirectRuleConditionsResponse{}
	for _, condition := range redirectRuleAllConditions {
		result.Options = append(result.Options, &redirectRuleConditionsItem{
			condition.GetString("name"),
			condition.GetString("code"),
			condition.GetString("description"),
		})
	}
	return result, nil
}

// CommonParameters 常用参数下拉框
func (this *ServerService) CommonParameters(ctx context.Context, req *pb.RPCSuccess) (*CommonParametersResponse, error) {

	return &CommonParametersResponse{
		Options: []*commonParametersItem{
			{"${edgeVersion}", "边缘节点版本"},
			{"${remoteAddr}", "客户端地址（IP）"},
			{"${rawRemoteAddr}", "客户端地址（IP）"},
			{"${remotePort}", "客户端端口"},
			{"${remoteUser}", "客户端用户名"},
			{"${requestURI}", "请求URI"},
			{"${requestPath}", "请求路径（不包括参数）"},
			{"${requestURL}", "完整的请求URL"},
			{"${requestLength}", "请求内容长度"},
			{"${requestMethod}", "请求方式"},
			{"${requestFilename}", "请求文件路径"},
			{"${scheme}", "请求协议，http或https"},
			{"${proto}", "包含版本的HTTP请求协议"},
			{"${timeISO8601}", "ISO 8601格式的时间"},
			{"${timeLocal}", "本地时间"},
			{"${msec}", "带有毫秒的时间"},
			{"${timestamp}", "unix时间戳，单位秒"},
			{"${host}", "主机名"},
			{"${serverName}", "接收请求的服务器名"},
			{"${serverPort}", "接收请求的服务器端口"},
			{"${refer}", "请求来源URL"},
			{"${refer.host}", "请求来源URL域名"},
			{"${userAgent}", "客户端信息"},
			{"${contentType}", "请求头部的Content-Type"},
			{"${cookies}", "所有cookie组合字符串"},
			{"${cookies.NAME}", "单个cookie值"},
			{"${isArgs}", "问好（?）标记"},
			{"${args}", "所有参数组合字符串"},
			{"${arg.NAME}", "单个参数值"},
			{"${headers}", "所有Header信息组合字符串"},
			{"${header.NAME}", "单个Header值"},
			{"${geo.country.name}", "国家/地区名称"},
			{"${geo.country.id}", "国家/地区ID"},
			{"${geo.province.name}", "省份名称"},
			{"${geo.province.id}", "省份ID"},
			{"${geo.city.name}", "城市名称"},
			{"${geo.city.id}", "城市ID"},
		},
	}, nil
}

// CreateRedirectRule 创建URL跳转规则
func (this *ServerService) CreateRedirectRule(ctx context.Context, req *CreateRedirectRuleRequest) (*CreateRedirectRuleResponse, error) {

	var MatchPrefix bool
	var MatchRegexp bool
	switch req.Mode {
	case "matchPrefix":
		MatchPrefix = true
		MatchRegexp = false
		req.KeepArgs = false
	case "equal":
		MatchPrefix = false
		MatchRegexp = false
		req.KeepRequestURI = false
	case "matchRegexp":
		MatchPrefix = false
		MatchRegexp = true
		req.KeepRequestURI = false
	default:
		return nil, errors.New("mode 请输入正确的匹配模式：equal/matchPrefix/matchRegexp")
	}
	if req.ServerId == 0 {
		return nil, errors.New("serverId 不能为空")
	}
	tx := this.NullTx()
	// 开启访问日志和Websocket
	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}
	webConfig, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	if req.ServerId == 0 {
		return nil, errors.New("serverId not be empty")
	}
	if req.BeforeURL == "" {
		return nil, errors.New("beforeURL not be empty")
	}
	if req.AfterURL == "" {
		return nil, errors.New("afterURL not be empty")
	}
	// 校验格式
	if MatchRegexp {
		_, err := regexp.Compile(req.BeforeURL)
		if err != nil {
			return nil, errors.New("跳转前URL正则表达式错误：" + err.Error())
		}
	} else {
		u, err := url.Parse(req.BeforeURL)
		if err != nil {
			return nil, errors.New("beforeURL 请输入正确的跳转前URL")
		}
		if (u.Scheme != "http" && u.Scheme != "https") ||
			len(u.Host) == 0 {
			return nil, errors.New("beforeURL 请输入正确的跳转前URL")
		}
		u, err = url.Parse(req.AfterURL)
		if err != nil {
			return nil, errors.New("afterURL 请输入正确的跳转后URL")
		}
		if (u.Scheme != "http" && u.Scheme != "https") ||
			len(u.Host) == 0 {
			return nil, errors.New("afterURL 请输入正确的跳转后URL")
		}
	}

	if req.StatusCode < 0 {
		return nil, errors.New("statusCode 请选择正确的跳转状态码")
	}
	// 校验匹配条件
	var conds *shared.HTTPRequestCondsConfig
	if len(req.Cond.Group) > 0 {
		conds = &shared.HTTPRequestCondsConfig{}
		conds.IsOn = true
		conds.Connector = req.Cond.Connector
		for _, group := range req.Cond.Group {
			item := &shared.HTTPRequestCondGroup{IsOn: group.IsON, Description: group.Description, Connector: group.Connector}
			for _, cond := range group.Conds {
				i := &shared.HTTPRequestCond{Type: cond.Type, Operator: cond.Operator, Param: cond.Param, IsCaseInsensitive: cond.IsCaseInsensitive, IsRequest: true}
				if cond.Values == nil || len(cond.Values) == 0 {
					i.Value = ""
				} else if len(cond.Values) > 1 {
					if len(cond.Values) == 1 {
						i.Value = cond.Values[0]
					} else {
						i.Value, _ = object2string(cond.Values)
					}
				} else {
					i.Value = cond.Values[0]
				}

				if cond.Type != "params" {
					i.Param = redirectRuleConditionOptions[i.Type][0]
					i.Operator = redirectRuleConditionOptions[i.Type][1]
				}
				if cond.Type == "url-not-prefix" || cond.Type == "url-not-eq" {
					i.IsReverse = true
				}
				item.Conds = append(item.Conds, i)
			}
			conds.Groups = append(conds.Groups, item)
		}
		err = conds.Init()
		if err != nil {
			return nil, errors.New("匹配条件校验失败:" + err.Error())
		}
	}

	hostRedirects := []*serverconfigs.HTTPHostRedirectConfig{}
	if len(webConfig.HostRedirects) > 0 {
		hostRedirects = append(hostRedirects, webConfig.HostRedirects...)
	}
	hostRedirects = append(hostRedirects, &serverconfigs.HTTPHostRedirectConfig{
		IsOn:           req.IsON,
		Status:         req.StatusCode,
		Mode:           req.Mode,
		BeforeURL:      req.BeforeURL,
		AfterURL:       req.AfterURL,
		MatchPrefix:    MatchPrefix,
		MatchRegexp:    MatchRegexp,
		KeepArgs:       req.KeepArgs,
		KeepRequestURI: req.KeepRequestURI,
		Conds:          conds,
	})
	err = models.SharedHTTPWebDAO.UpdateWebHostRedirects(tx, webConfig.Id, hostRedirects)
	if err != nil {
		return nil, err
	}
	return &CreateRedirectRuleResponse{Id: len(hostRedirects)}, nil
}
func object2string(obj interface{}) (string, error) {

	tmp, err := json.Marshal(obj)
	return string(tmp), err
}

type CreateInboundGroupRuleSetRequest struct {
	GroupId   int64  `json:"groupId"`
	Name      string `json:"name"`
	Connector string `json:"connector"` // or / and
	Rules     []struct {
		Prefix       string   `json:"prefix"`
		Param        string   `json:"param"`
		Operator     string   `json:"operator"`
		Value        string   `json:"value"`
		Threshold    int      `json:"threshold"`
		Period       int      `json:"period"`
		Objects      []string `json:"objects"`
		Length       int      `json:"length"`
		Headers      []string `json:"headers"`
		Description  string   `json:"description"`
		ParamFilters []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"paramFilters"`
		AllowEmpty      bool     `json:"allowEmpty"`
		AllowSameDomain bool     `json:"allowSameDomain"`
		AllowDomains    []string `json:"allowDomains"`
	} `json:"rules"`
	Actions []struct {
		Code             string   `json:"code"`
		Timeout          int      `json:"timeout"`
		Scope            string   `json:"scope"` // global / service
		Life             int      `json:"life"`
		MaxFails         int      `json:"maxFails"`
		FailBlockTimeout int      `json:"failBlockTimeout"`
		Level            string   `json:"level"`
		IpListId         int      `json:"ipListId"`
		Type             string   `json:"type"` // black / white
		Tags             []string `json:"tags"`
		Status           int      `json:"status"`
		Body             string   `json:"body"`
		GroupId          int      `json:"groupId"`
		SetId            int      `json:"setId"`
	} `json:"actions"`
	IgnoreLocal bool `json:"ignoreLocal"` // 忽略局域网IP
}
type CreateInboundGroupRuleSetResponse struct {
	SetId int64 `json:"setId"`
}

// InboundGroupRuleOptions 新增入站规则分组的规则集 - 规则
func (this *ServerService) InboundGroupRuleOptions(ctx context.Context, req *pb.RPCSuccess) (*ReverseProxySchedulingOptionsResponse, error) {
	// check points
	checkpointList := []maps.Map{}
	for _, checkpoint := range firewallconfigs.AllCheckpoints {
		if checkpoint.IsRequest {
			checkpointList = append(checkpointList, maps.Map{
				"name":        checkpoint.Name,
				"prefix":      checkpoint.Prefix,
				"description": checkpoint.Description,
				"isComposed":  checkpoint.IsComposed,
				"params":      checkpoint.Params,
			})
		}
	}
	return &ReverseProxySchedulingOptionsResponse{Options: checkpointList}, nil
}

// InboundGroupOperatorsOptions 新增入站规则分组的规则集 - 操作符
func (this *ServerService) InboundGroupOperatorsOptions(ctx context.Context, req *pb.RPCSuccess) (*ReverseProxySchedulingOptionsResponse, error) {
	// check points
	checkpointList := []maps.Map{}
	for _, checkpoint := range firewallconfigs.AllRuleOperators {

		checkpointList = append(checkpointList, maps.Map{
			"name":        checkpoint.Name,
			"code":        checkpoint.Code,
			"description": checkpoint.Description,
		})

	}
	return &ReverseProxySchedulingOptionsResponse{Options: checkpointList}, nil
}

// InboundGroupActionOptions 新增入站规则分组的规则集 - 动作
func (this *ServerService) InboundGroupActionOptions(ctx context.Context, req *pb.RPCSuccess) (*ReverseProxySchedulingOptionsResponse, error) {

	// 所有可选的动作
	actionMaps := []maps.Map{}
	for _, action := range firewallconfigs.AllActions {
		actionMaps = append(actionMaps, maps.Map{
			"name":        action.Name,
			"description": action.Description,
			"code":        action.Code,
		})
	}
	return &ReverseProxySchedulingOptionsResponse{Options: actionMaps}, nil
}

// StatObjectOptions CC统计 - 统计对象下拉框
func (this *ServerService) StatObjectOptions(ctx context.Context, req *pb.RPCSuccess) (*ReverseProxySchedulingOptionsResponse, error) {

	// 所有可选的动作
	objectMaps := []maps.Map{}
	for _, obj := range serverconfigs.FindAllMetricKeyDefinitions(serverconfigs.MetricItemCategoryHTTP) {
		objectMaps = append(objectMaps, maps.Map{
			"name":        obj.Name,
			"code":        obj.Code,
			"description": obj.Description,
		})
	}
	return &ReverseProxySchedulingOptionsResponse{objectMaps}, nil
}

// CodecOptions 通用参数 - 编译码下拉框
func (this *ServerService) CodecOptions(ctx context.Context, req *pb.RPCSuccess) (*ReverseProxySchedulingOptionsResponse, error) {

	return &ReverseProxySchedulingOptionsResponse{Options: []maps.Map{
		{"name": "MD5", "code": "md5"},
		{"name": "URLEncode", "code": "urlEncode"},
		{"name": "URLDecode", "code": "urlDecode"},
		{"name": "BASE64Encode", "code": "base64Encode"},
		{"name": "BASE64Decode", "code": "base64Decode"},
		{"name": "UNICODE编码", "code": "unicodeEncode"},
		{"name": "UNICODE解码", "code": "unicodeDecode"},
		{"name": "HTML实体编码", "code": "htmlEscape"},
		{"name": "HTML实体解码", "code": "htmlUnescape"},
		{"name": "计算长度", "code": "length"},
		{"name": "十六进制->十进制", "code": "hex2dec"},
		{"name": "十进制->十六进制", "code": "dec2hex"},
		{"name": "SHA1", "code": "sha1"},
		{"name": "SHA256", "code": "sha256"},
	}}, nil
}

// CreateInboundGroupRuleSet 创建入站规则分组规则集
func (this *ServerService) CreateInboundGroupRuleSet(ctx context.Context, req *CreateInboundGroupRuleSetRequest) (*CreateInboundGroupRuleSetResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupConfig, err := models.SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(tx, req.GroupId)
	if err != nil {
		return nil, errors.New("find rule group config error:" + err.Error())
	}
	if groupConfig == nil {
		return nil, errors.New("找不到分组，Id：" + strconv.FormatInt(req.GroupId, 10))
	}
	if req.Name == "" {
		return nil, errors.New("请输入规则集名称")
	}
	if len(req.Rules) == 0 {
		return nil, errors.New("请添加至少一个规则")
	}
	rules := []*firewallconfigs.HTTPFirewallRule{}
	for _, rule := range req.Rules {
		item := &firewallconfigs.HTTPFirewallRule{IsOn: true, CheckpointOptions: map[string]interface{}{}}

		if len(rule.Param) > 0 {
			item.Param = "${" + rule.Prefix + "." + rule.Param + "}"
		} else {
			item.Param = "${" + rule.Prefix + "}"
		}

		item.Operator = rule.Operator
		item.Description = rule.Description
		item.Value = rule.Value
		item.IsCaseInsensitive = true
		if rule.Prefix == "cc2" {
			item.CheckpointOptions["keys"] = rule.Objects
			item.CheckpointOptions["period"] = rule.Period
			item.CheckpointOptions["threshold"] = rule.Threshold
			item.Operator = "gt"
		} else if rule.Prefix == "requestGeneralHeaderLength" {
			item.CheckpointOptions["headers"] = rule.Headers
			item.CheckpointOptions["length"] = rule.Length
			item.Operator = "match"
		} else if rule.Prefix == "refererBlock" {
			item.CheckpointOptions["allowEmpty"] = rule.AllowEmpty
			item.CheckpointOptions["allowSameDomain"] = rule.AllowSameDomain
			item.CheckpointOptions["allowDomains"] = rule.AllowDomains
			item.Operator = "eq"
		} else {
			for _, v := range rule.ParamFilters {
				item.ParamFilters = append(item.ParamFilters, &firewallconfigs.ParamFilter{
					Name: v.Name,
					Code: v.Code,
				})
			}
		}

		// 校验
		err := item.Init()
		if err != nil {
			return nil, errors.New("校验规则 '" + rule.Param + " " + rule.Operator + " " + rule.Value + "' 失败，原因：" + err.Error())
		}
		rules = append(rules, item)
	}
	var actionConfigs = []*firewallconfigs.HTTPFirewallActionConfig{}
	if len(req.Actions) == 0 {
		return nil, errors.New("请添加至少一个动作")
	}
	for _, action := range req.Actions {
		item := &firewallconfigs.HTTPFirewallActionConfig{Code: action.Code,
			Options: maps.Map{},
		}
		switch action.Code {
		case firewallconfigs.HTTPFirewallActionBlock:
			if action.Scope != " service" && action.Scope != "global" {
				return nil, errors.New("请输入有效的封锁范围：service/global")
			}
			item.Options = maps.Map{
				"scope":   action.Scope,
				"timeout": action.Timeout,
			}
		case firewallconfigs.HTTPFirewallActionAllow, firewallconfigs.HTTPFirewallActionLog, firewallconfigs.HTTPFirewallActionNotify:
		case firewallconfigs.HTTPFirewallActionCaptcha:
			item.Options = maps.Map{
				"maxFails":         action.MaxFails,
				"failBlockTimeout": action.FailBlockTimeout,
			}
			if action.Life != 0 {
				item.Options["life"] = action.Life
			}
			if action.FailBlockTimeout != 0 {
				item.Options["failBlockTimeout"] = action.FailBlockTimeout
			}
		case firewallconfigs.HTTPFirewallActionGet302, firewallconfigs.HTTPFirewallActionPost307:
			item.Options = maps.Map{
				"life": action.Life,
			}
		case firewallconfigs.HTTPFirewallActionRecordIP:
			item.Options = maps.Map{
				"type":     action.Type,
				"level":    action.Level,
				"timeout":  action.Timeout,
				"ipListId": action.IpListId,
			}
		case firewallconfigs.HTTPFirewallActionTag:
			item.Options["tags"] = action.Tags
		case firewallconfigs.HTTPFirewallActionPage:
			if action.Status == 0 || len(action.Body) == 0 {
				return nil, errors.New("请设置状态码，以及跳转页面内容")
			}
			item.Options["status"] = action.Status
			item.Options["body"] = action.Body
		case firewallconfigs.HTTPFirewallActionGoGroup:
			item.Options["groupId"] = action.GroupId
		case firewallconfigs.HTTPFirewallActionGoSet:
			item.Options["groupId"] = action.GroupId
			item.Options["setId"] = action.SetId
		default:
			return nil, errors.New("无效的动作类型")
		}
		actionConfigs = append(actionConfigs, item)
	}
	setConfig := &firewallconfigs.HTTPFirewallRuleSet{
		Id:          0,
		IsOn:        true,
		Name:        req.Name,
		Code:        "",
		Description: "",
		Connector:   req.Connector,
		RuleRefs:    nil,
		Rules:       rules,
		Actions:     actionConfigs,
		IgnoreLocal: req.IgnoreLocal,
	}

	setId, err := models.SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(tx, setConfig)
	if err != nil {
		return nil, err
	}
	groupConfig.SetRefs = append(groupConfig.SetRefs, &firewallconfigs.HTTPFirewallRuleSetRef{
		IsOn:  true,
		SetId: setId,
	})

	setRefsJSON, err := json.Marshal(groupConfig.SetRefs)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(tx, req.GroupId, setRefsJSON)
	if err != nil {
		return nil, err
	}

	return &CreateInboundGroupRuleSetResponse{SetId: setId}, nil
}

type UpdateReverseProxySchedulingRequest struct {
	ServerId int64 `json:"serverId"`

	Type        string `json:"type"`
	HashKey     string `json:"hashKey"`
	StickyType  string `json:"stickyType"`
	StickyParam string `json:"stickyParam"`
	Family      string `json:"family"` // http/tcp/udp/unix
}

// UpdateReverseProxyScheduling 修改反向代理的调度算法
func (this *ServerService) UpdateReverseProxyScheduling(ctx context.Context, req *UpdateReverseProxySchedulingRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	var reverseProxyId int64
	reverseProxyRef, err := models.SharedServerDAO.FindReverseProxyRef(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		reverseProxyId, err = models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, userId, nil, nil, nil)
		if err != nil {
			return nil, err
		}

		reverseProxyRef = &serverconfigs.ReverseProxyRef{
			IsOn:           false,
			ReverseProxyId: reverseProxyId,
		}
		refJSON, err := json.Marshal(reverseProxyRef)
		if err != nil {
			return nil, err
		}
		err = models.SharedServerDAO.UpdateServerReverseProxy(tx, req.ServerId, refJSON)
		if err != nil {
			return nil, err
		}
	} else {
		reverseProxyId = reverseProxyRef.ReverseProxyId
	}

	reverseProxy, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyId, nil)
	if err != nil {
		return nil, err
	}

	if reverseProxy.Scheduling == nil {
		reverseProxy.FindSchedulingConfig()
	}
	options := maps.Map{}
	if req.Type == "hash" {
		if req.HashKey == "" {
			return nil, errors.New("请输入key")
		}
		options["key"] = req.HashKey
	} else if req.Type == "sticky" {

		if req.StickyType == "" {
			return nil, errors.New("请选择参数类型")
		}
		if req.StickyParam == "" {
			return nil, errors.New("请输入参数名")
		}
		reg, err := regexp.Compile("^[a-zA-Z0-9]+$")
		if err != nil {
			return nil, err
		}
		if !reg.MatchString(req.StickyParam) {
			return nil, errors.New("参数名只能是英文字母和数字的组合")
		}
		if len([]rune(req.StickyParam)) > 50 {
			return nil, errors.New("参数名长度不能超过50位")
		}
		options["type"] = req.StickyType
		options["param"] = req.StickyParam
	}

	if schedulingconfigs.FindSchedulingType(req.Type) == nil {
		return nil, errors.New("不支持此种算法")
	}

	reverseProxy.Scheduling.Code = req.Type
	reverseProxy.Scheduling.Options = options

	schedulingJSON, err := json.Marshal(reverseProxy.Scheduling)
	if err != nil {
		return nil, err
	}

	err = models.SharedReverseProxyDAO.UpdateReverseProxyScheduling(tx, reverseProxyId, schedulingJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

type ReverseProxySchedulingOptionsResponse struct {
	Options []maps.Map `json:"options"`
}

// ReverseProxySchedulingOptions 反向代理的调度算法下拉框
func (this *ServerService) ReverseProxySchedulingOptions(ctx context.Context, req *UpdateReverseProxySchedulingRequest) (*ReverseProxySchedulingOptionsResponse, error) {

	// 调度类型
	schedulingTypes := []maps.Map{}

	var isHTTPFamily = false
	var isTCPFamily = false
	var isUDPFamily = false
	var isUnixFamily = false
	tx := this.NullTx()
	if req.ServerId > 0 {
		server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
		if err != nil {
			return nil, errors.New("find server config error:" + err.Error())
		}
		// 配置
		serverConfig, err := models.SharedServerDAO.ComposeServerConfig(tx, server, nil, false)
		if err != nil {
			return nil, errors.New("parse server config error:" + err.Error())
		}
		isHTTPFamily = serverConfig.IsHTTPFamily()
		isTCPFamily = serverConfig.IsTCPFamily()
		isUDPFamily = serverConfig.IsUDPFamily()
		isUnixFamily = serverConfig.IsUnixFamily()
	} else {
		switch req.Family {
		case "http":
			isHTTPFamily = true
		case "tcp":
			isTCPFamily = true
		case "udp":
			isUDPFamily = true
		case "unix":
			isUnixFamily = true
		}
	}
	for _, m := range schedulingconfigs.AllSchedulingTypes() {
		networks, ok := m["networks"]
		if !ok {
			continue
		}
		if !types.IsSlice(networks) {
			continue
		}
		if (isHTTPFamily && lists.Contains(networks, "http")) ||
			(isTCPFamily && lists.Contains(networks, "tcp")) ||
			(isUDPFamily && lists.Contains(networks, "udp")) ||
			(isUnixFamily && lists.Contains(networks, "unix")) {
			schedulingTypes = append(schedulingTypes, m)
		}
	}
	return &ReverseProxySchedulingOptionsResponse{Options: schedulingTypes}, nil
}

type DeleteRuleGroupRequest struct {
	ServerId int64 `json:"serverId"`
	GroupId  int64 `json:"groupId"`
}

// DeleteRuleGroup 删除规则分组
func (this *ServerService) DeleteRuleGroup(ctx context.Context, req *DeleteRuleGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if req.ServerId == 0 {
		return nil, errors.New("serverId 不能为空")
	}
	tx := this.NullTx()
	// 开启访问日志和Websocket
	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}
	webConfig, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	var firewallPolicyId int64
	if webConfig.FirewallRef == nil || webConfig.FirewallRef.FirewallPolicyId == 0 {
		firewallPolicyId, err = dao.SharedHTTPWebDAO.InitEmptyHTTPFirewallPolicy(ctx, 0, req.ServerId, webConfig.Id, webConfig.FirewallRef != nil && webConfig.FirewallRef.IsOn)
		if err != nil {
			return nil, errors.New("init empty http  firewall policy error:" + err.Error())
		}
	} else {
		firewallPolicyId = webConfig.FirewallPolicy.Id
	}
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, firewallPolicyId, nil)
	if err != nil {
		return nil, errors.New("find http  firewall policy config error:" + err.Error())
	}
	firewallPolicy.RemoveRuleGroup(req.GroupId)

	inboundJSON, err := firewallPolicy.InboundJSON()
	if err != nil {
		return nil, err
	}

	outboundJSON, err := firewallPolicy.OutboundJSON()
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(tx, firewallPolicyId, inboundJSON, outboundJSON, true)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return this.Success()
}

type UpdateCompressionRequest struct {
	ServerId        int64    `json:"serverId"`
	IsOn            bool     `json:"isOn"`
	UseDefaultTypes bool     `json:"useDefaultTypes"` // 压缩算法
	Types           []string `json:"types"`           // 压缩算法类型：brotli/gzip/deflate
	Level           int      `json:"level"`           // 压缩级别
	DecompressData  bool     `json:"decompressData"`  // 支持已压缩内容
	MinLength       struct {
		Count int    `json:"count"`
		Unit  string `json:"unit"` // 字节/ kb mb gb tb pb eb
	} `json:"minLength"`
	MaxLength struct {
		Count int    `json:"count"`
		Unit  string `json:"unit"`
	} `json:"maxLength"`
	MimeTypes  []string `json:"mimeTypes"`  //支持的MimeType
	Extensions []string `json:"extensions"` //支持的扩展名
	Cond       struct {
		IsON      bool            `json:"isOn"`
		Connector string          `json:"connector"` // or/and
		Group     []condGroupItem `json:"groups"`
	} `json:"conds"` // 条件
}
type condGroupItem struct {
	IsON        bool   `json:"isOn"`
	Connector   string `json:"connector"` // or/and
	Description string `json:"description"`
	Conds       []struct {
		Type              string   `json:"type"`              //条件类型
		Values            []string `json:"values"`            //值列表
		IsCaseInsensitive bool     `json:"isCaseInsensitive"` // 是否区分大小写
		Operator          string   `json:"operator"`          //操作法 大于小于等
		Param             string   `json:"param"`             //参数
	} `json:"conds"`
}

// UpdateCompression 修改内容压缩配置
func (this *ServerService) UpdateCompression(ctx context.Context, req *UpdateCompressionRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if req.ServerId == 0 {
		return nil, errors.New("serverId 不能为空")
	}
	tx := this.NullTx()
	// 开启访问日志和Websocket
	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}
	if len(req.Cond.Group) > 0 {
		req.Cond.IsON = true
	}
	// 默认加1个前缀'.'
	for k, v := range req.Extensions {
		if !strings.HasPrefix(v, ".") {
			req.Extensions[k] = "." + v
		}
	}
	if len(req.Cond.Group) == 0 {
		req.Cond.Group = make([]condGroupItem, 0)
		req.Cond.Connector = "or"
	}
	compressionJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = models.SharedHTTPWebDAO.UpdateWebCompression(tx, webId, compressionJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateUAM 开启/关闭指定服务的5秒盾
func (this *ServerService) UpdateUAM(ctx context.Context, req *UpdateCompressionRequest) (*pb.RPCSuccess, error) {
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

	var config = &serverconfigs.UAMConfig{IsOn: req.IsOn}

	err = models.SharedServerDAO.UpdateServerUAM(tx, req.ServerId, config)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
