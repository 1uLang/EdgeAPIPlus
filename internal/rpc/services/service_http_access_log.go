package services

import (
	"context"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/accesslogs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/iplibrary"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
)

// HTTPAccessLogService 访问日志相关服务
type HTTPAccessLogService struct {
	BaseService
}

// CreateHTTPAccessLogs 创建访问日志
func (this *HTTPAccessLogService) CreateHTTPAccessLogs(ctx context.Context, req *pb.CreateHTTPAccessLogsRequest) (*pb.CreateHTTPAccessLogsResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	if len(req.HttpAccessLogs) == 0 {
		return &pb.CreateHTTPAccessLogsResponse{}, nil
	}

	tx := this.NullTx()

	err = models.SharedHTTPAccessLogDAO.CreateHTTPAccessLogs(tx, req.HttpAccessLogs)
	if err != nil {
		return nil, err
	}

	// 发送到访问日志策略
	policyId, err := models.SharedHTTPAccessLogPolicyDAO.FindCurrentPublicPolicyId(tx)
	if err != nil {
		return nil, err
	}
	if policyId > 0 {
		err = accesslogs.SharedStorageManager.Write(policyId, req.HttpAccessLogs)
		if err != nil {
			return nil, err
		}
	}

	return &pb.CreateHTTPAccessLogsResponse{}, nil
}

// ListHTTPAccessLogs 列出单页访问日志
func (this *HTTPAccessLogService) ListHTTPAccessLogs(ctx context.Context, req *pb.ListHTTPAccessLogsRequest) (*pb.ListHTTPAccessLogsResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查服务ID
	if userId > 0 {
		if req.UserId > 0 && userId != req.UserId {
			return nil, this.PermissionError()
		}

		// 这里不用担心serverId <= 0 的情况，因为如果userId>0，则只会查询当前用户下的服务，不会产生安全问题
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	}

	accessLogs, requestId, hasMore, err := models.SharedHTTPAccessLogDAO.ListAccessLogs(tx, req.RequestId, req.Size, req.Day, req.ServerId, req.Reverse, req.HasError, req.FirewallPolicyId, req.FirewallRuleGroupId, req.FirewallRuleSetId, req.HasFirewallPolicy, req.UserId, req.Keyword, req.Ip, req.Domain)
	if err != nil {
		return nil, err
	}

	result := []*pb.HTTPAccessLog{}
	var pbNodeMap = map[int64]*pb.Node{}
	var pbClusterMap = map[int64]*pb.NodeCluster{}
	for _, accessLog := range accessLogs {
		a, err := accessLog.ToPB()
		if err != nil {
			return nil, err
		}

		// 节点 & 集群
		pbNode, ok := pbNodeMap[a.NodeId]
		if ok {
			a.Node = pbNode
		} else {
			node, err := models.SharedNodeDAO.FindEnabledNode(tx, a.NodeId)
			if err != nil {
				return nil, err
			}
			if node != nil {
				pbNode = &pb.Node{Id: int64(node.Id), Name: node.Name}

				var clusterId = int64(node.ClusterId)
				pbCluster, ok := pbClusterMap[clusterId]
				if ok {
					pbNode.NodeCluster = pbCluster
				} else {
					cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, clusterId)
					if err != nil {
						return nil, err
					}
					if cluster != nil {
						pbCluster = &pb.NodeCluster{
							Id:   int64(cluster.Id),
							Name: cluster.Name,
						}
						pbNode.NodeCluster = pbCluster
						pbClusterMap[clusterId] = pbCluster
					}
				}

				pbNodeMap[a.NodeId] = pbNode
				a.Node = pbNode
			}
		}

		result = append(result, a)
	}
	return &pb.ListHTTPAccessLogsResponse{
		HttpAccessLogs: result,
		AccessLogs:     result, // TODO 仅仅为了兼容，当用户节点版本大于0.0.8时可以删除
		HasMore:        hasMore,
		RequestId:      requestId,
	}, nil
}

// SearchHTTPAccessLogs 列出单页访问日志
func (this *HTTPAccessLogService) SearchHTTPAccessLogs(ctx context.Context, req *pb.SearchHTTPAccessLogsRequest) (*pb.SearchHTTPAccessLogsResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查服务ID
	if userId > 0 {
		if req.UserId > 0 && userId != req.UserId {
			return nil, this.PermissionError()
		}

		// 这里不用担心serverId <= 0 的情况，因为如果userId>0，则只会查询当前用户下的服务，不会产生安全问题
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	}
	accessLogs, requestId, err := models.SharedHTTPAccessLogDAO.SearchAccessLogs(tx,
		req.RequestId, req.Day, req.Ip, req.Domain, req.Code, req.RequestMethod, req.Keyword, req.StartAt, req.EndAt, req.UserId, req.Size, req.HasAll, req.HasError)

	if err != nil {
		return nil, err
	}

	result := []*pb.HTTPAccessLog{}
	var pbNodeMap = map[int64]*pb.Node{}
	var pbClusterMap = map[int64]*pb.NodeCluster{}
	for _, accessLog := range accessLogs {
		a, err := accessLog.ToPB()
		if err != nil {
			return nil, err
		}

		// 节点 & 集群
		pbNode, ok := pbNodeMap[a.NodeId]
		if ok {
			a.Node = pbNode
		} else {
			node, err := models.SharedNodeDAO.FindEnabledNode(tx, a.NodeId)
			if err != nil {
				return nil, err
			}
			if node != nil {
				pbNode = &pb.Node{Id: int64(node.Id), Name: node.Name}

				var clusterId = int64(node.ClusterId)
				pbCluster, ok := pbClusterMap[clusterId]
				if ok {
					pbNode.NodeCluster = pbCluster
				} else {
					cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, clusterId)
					if err != nil {

						return nil, err
					}
					if cluster != nil {
						pbCluster = &pb.NodeCluster{
							Id:   int64(cluster.Id),
							Name: cluster.Name,
						}
						pbNode.NodeCluster = pbCluster
						pbClusterMap[clusterId] = pbCluster
					}
				}

				pbNodeMap[a.NodeId] = pbNode
				a.Node = pbNode
			}
		}

		result = append(result, a)
	}
	return &pb.SearchHTTPAccessLogsResponse{
		HttpAccessLogs: result,
		HasMore:        requestId != "",
		RequestId:      requestId,
	}, nil
}

// FindHTTPAccessLog 查找单个日志
func (this *HTTPAccessLogService) FindHTTPAccessLog(ctx context.Context, req *pb.FindHTTPAccessLogRequest) (*pb.FindHTTPAccessLogResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	accessLog, err := models.SharedHTTPAccessLogDAO.FindAccessLogWithRequestId(tx, req.RequestId)
	if err != nil {
		return nil, err
	}
	if accessLog == nil {
		return &pb.FindHTTPAccessLogResponse{HttpAccessLog: nil}, nil
	}

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, int64(accessLog.ServerId))
		if err != nil {
			return nil, err
		}
	}

	a, err := accessLog.ToPB()
	if err != nil {
		return nil, err
	}
	return &pb.FindHTTPAccessLogResponse{HttpAccessLog: a}, nil
}

// StatisticsHTTPAccessTop 统计攻击ip排行
func (this *HTTPAccessLogService) StatisticsHTTPAccessTop(ctx context.Context, req *pb.StatisticsHTTPAccessTopRequest) (*pb.StatisticsHTTPAccessTopResponse, error) {
	// 校验请求
	if len(req.Day) == 0 {
		return &pb.StatisticsHTTPAccessTopResponse{}, nil
	}
	tx := this.NullTx()
	//防止存在 循环包
	stats, err := models.SharedHTTPAccessLogDAO.StatisticsTop(tx, req.Day, req.User, func(s string) (string, string) {
		r, _ := iplibrary.SharedLibrary.Lookup(s)
		//忽略国外的攻击
		if r == nil || r.Country != "中国" {
			return "", ""
		}
		return r.Province, r.City
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.StatisticsHTTPAccessTopResponse{}
	for _, v := range stats.Tops {
		stat := &pb.StatisticsHTTPAccess{ServerId: v.ServerId, Total: v.Total}
		stat.Ip = &pb.AccessTop{Names: v.Ips.IP, Counts: v.Ips.Count}
		stat.Region = &pb.AccessTop{Names: v.Region.Region, Counts: v.Region.Count}
		resp.Stats = append(resp.Stats, stat)
	}
	return resp, nil
}

// StatisticsHTTPAccess 统计指定日期下用户的攻击次数
func (this *HTTPAccessLogService) StatisticsHTTPAccess(ctx context.Context, req *pb.StatisticsHTTPAccessRequest) (*pb.StatisticsHTTPAccessResponse, error) {
	// 校验请求
	if len(req.Days) == 0 {
		return &pb.StatisticsHTTPAccessResponse{}, nil
	}
	tx := this.NullTx()
	//防止存在 循环包
	counts, err := models.SharedHTTPAccessLogDAO.Statistics(tx, req.Days, req.User)
	if err != nil {
		return nil, err
	}
	resp := &pb.StatisticsHTTPAccessResponse{Counts: counts}
	return resp, nil
}

// StatisticsHTTPAccessType 统计各类型的攻击策略条数
func (this *HTTPAccessLogService) StatisticsHTTPAccessType(ctx context.Context, req *pb.StatisticsHTTPAccessTypeRequest) (*pb.StatisticsHTTPAccessTypeResponse, error) {
	// 校验请求
	tx := this.NullTx()
	//防止存在 循环包
	counts, err := models.SharedHTTPAccessLogDAO.StatisticsType(tx, req.Day, req.User)
	if err != nil {
		return nil, err
	}
	resp := &pb.StatisticsHTTPAccessTypeResponse{}
	for _, v := range counts {
		resp.Attacks = append(resp.Attacks, &pb.HTTPAccessType{ServerId: fmt.Sprintf("%d", v.ServerId), Count: v.Count, Code: v.Code, Name: v.Name})
	}
	return resp, nil
}

// StatisticsHTTPAccessLogs 统计指定用户日期下 各访问的 访问条数 访问总次数  防护总次数 访问IP总数 拦截IP总数
func (this *HTTPAccessLogService) StatisticsHTTPAccessLogs(ctx context.Context, req *pb.StatisticsHTTPAccessTypeRequest) (*pb.StatisticsHTTPAccessLogResponse, error) {
	// 校验请求
	tx := this.NullTx()
	//防止存在 循环包
	stats, err := models.SharedHTTPAccessLogDAO.AccessStatistics(tx, req.Day, req.User)
	if err != nil {
		return nil, err
	}
	resp := &pb.StatisticsHTTPAccessLogResponse{}
	for _, v := range stats {
		resp.Attacks = append(resp.Attacks, &pb.HTTPAccessStat{ServerId: v.ServerId, AccessTotal: v.AccessTotal, AttackTotal: v.AttackTotal,
			AccessIpTotal: v.AccessIpTotal, AttackIpTotal: v.AttackIpTotal})
	}

	return resp, nil
}

// StatisticsAttackURLTop 统计最受攻击的域名排行
func (this *HTTPAccessLogService) StatisticsAttackURLTop(ctx context.Context, req *pb.StatisticsHTTPAccessTopRequest) (*pb.StatisticsHTTPAttackURLTopResponse, error) {
	// 校验请求
	tx := this.NullTx()
	//防止存在 循环包
	stats, err := models.SharedHTTPAccessLogDAO.AttackURLTop(tx, req.Day, req.User)
	if err != nil {
		return nil, err
	}
	resp := &pb.StatisticsHTTPAttackURLTopResponse{}
	for _, v := range stats.Tops {
		item := pb.HTTPAttackURL{ServerId: v.ServerId}
		item.Uris = &pb.AttackCount{Values: v.Uris.Value, Counts: v.Uris.Count}
		item.Hosts = &pb.AttackCount{Values: v.Hosts.Value, Counts: v.Hosts.Count}
		resp.Attacks = append(resp.Attacks, &item)
	}

	return resp, nil
}

// StatisticsAccessIPTop 客户端访问IP排行
func (this *HTTPAccessLogService) StatisticsAccessIPTop(ctx context.Context, req *pb.StatisticsHTTPAccessTopRequest) (*pb.StatisticsHTTPAccessIPTopResponse, error) {
	// 校验请求
	tx := this.NullTx()
	//防止存在 循环包
	stats, err := models.SharedHTTPAccessLogDAO.AccessIPTop(tx, req.Day, req.User, int(req.Top))
	if err != nil {
		return nil, err
	}
	resp := &pb.StatisticsHTTPAccessIPTopResponse{}
	for _, v := range stats.Tops {
		resp.Access = append(resp.Access, &pb.HTTPAccessIP{ServerId: v.ServerId, Ip: v.IPs, Count: v.Counts})
	}

	return resp, nil
}

// StatusCodeStatistics 访问状态码统计
func (this *HTTPAccessLogService) StatusCodeStatistics(ctx context.Context, req *pb.StatisticsHTTPAccessTopRequest) (*pb.StatisticsStatusCodeTopResponse, error) {
	// 校验请求
	tx := this.NullTx()
	//防止存在 循环包
	stats, err := models.SharedHTTPAccessLogDAO.StatusCodeStatistics(tx, req.Day, req.User)
	if err != nil {
		return nil, err
	}
	resp := &pb.StatisticsStatusCodeTopResponse{}
	for _, v := range stats.Tops {
		resp.Codes = append(resp.Codes, &pb.StatusCode{ServerId: v.ServerId, Code: v.Codes, Count: v.Counts})
	}

	return resp, nil
}
