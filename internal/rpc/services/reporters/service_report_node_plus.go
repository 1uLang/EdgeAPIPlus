// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package reporters

import (
	"context"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/iplibrary"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/reporterconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"google.golang.org/grpc/peer"
	"net"
	"time"
)

// ReportNodeService 监控终端服务
type ReportNodeService struct {
	services.BaseService
}

// CreateReportNode 添加终端
func (this *ReportNodeService) CreateReportNode(ctx context.Context, req *pb.CreateReportNodeRequest) (*pb.CreateReportNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	reporterId, err := models.SharedReportNodeDAO.CreateReportNode(tx, req.Name, req.Location, req.Isp, req.AllowIPs, req.ReportNodeGroupIds)
	if err != nil {
		return nil, err
	}

	return &pb.CreateReportNodeResponse{ReportNodeId: reporterId}, nil
}

// DeleteReportNode 删除终端
func (this *ReportNodeService) DeleteReportNode(ctx context.Context, req *pb.DeleteReportNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedReportNodeDAO.DisableReportNode(tx, req.ReportNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateReportNode 修改终端
func (this *ReportNodeService) UpdateReportNode(ctx context.Context, req *pb.UpdateReportNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedReportNodeDAO.UpdateReportNode(tx, req.ReportNodeId, req.Name, req.Location, req.Isp, req.AllowIPs, req.ReportNodeGroupIds, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllEnabledReportNodes 计算终端数量
func (this *ReportNodeService) CountAllEnabledReportNodes(ctx context.Context, req *pb.CountAllEnabledReportNodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedReportNodeDAO.CountAllEnabledReportNodes(tx, req.ReportNodeGroupId, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledReportNodes 列出单页终端
func (this *ReportNodeService) ListEnabledReportNodes(ctx context.Context, req *pb.ListEnabledReportNodesRequest) (*pb.ListEnabledReportNodesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	ones, err := models.SharedReportNodeDAO.ListEnabledReportNodes(tx, req.ReportNodeGroupId, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbNodes = []*pb.ReportNode{}
	for _, one := range ones {
		var pbGroups = []*pb.ReportNodeGroup{}
		var groupIds = one.DecodeGroupIds()
		for _, groupId := range groupIds {
			group, err := models.SharedReportNodeGroupDAO.FindEnabledReportNodeGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if group == nil {
				continue
			}
			pbGroups = append(pbGroups, &pb.ReportNodeGroup{
				Id:   int64(group.Id),
				Name: group.Name,
				IsOn: group.IsOn == 1,
			})
		}

		pbNodes = append(pbNodes, &pb.ReportNode{
			Id:               int64(one.Id),
			UniqueId:         one.UniqueId,
			Secret:           one.Secret,
			IsOn:             one.IsOn == 1,
			Name:             one.Name,
			Location:         one.Location,
			Isp:              one.Isp,
			IsActive:         one.IsActive == 1,
			StatusJSON:       []byte(one.Status),
			AllowIPs:         one.DecodeAllowIPs(),
			ReportNodeGroups: pbGroups,
		})
	}

	return &pb.ListEnabledReportNodesResponse{
		ReportNodes: pbNodes,
	}, nil
}

// FindEnabledReportNode 查找单个终端
func (this *ReportNodeService) FindEnabledReportNode(ctx context.Context, req *pb.FindEnabledReportNodeRequest) (*pb.FindEnabledReportNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	node, err := models.SharedReportNodeDAO.FindEnabledReportNode(tx, req.ReportNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindEnabledReportNodeResponse{ReportNode: nil}, nil
	}

	var pbGroups = []*pb.ReportNodeGroup{}
	var groupIds = node.DecodeGroupIds()
	for _, groupId := range groupIds {
		group, err := models.SharedReportNodeGroupDAO.FindEnabledReportNodeGroup(tx, groupId)
		if err != nil {
			return nil, err
		}
		if group == nil {
			continue
		}
		pbGroups = append(pbGroups, &pb.ReportNodeGroup{
			Id:   int64(group.Id),
			Name: group.Name,
			IsOn: group.IsOn == 1,
		})
	}

	return &pb.FindEnabledReportNodeResponse{ReportNode: &pb.ReportNode{
		Id:               int64(node.Id),
		UniqueId:         node.UniqueId,
		Secret:           node.Secret,
		IsOn:             node.IsOn == 1,
		Name:             node.Name,
		Location:         node.Location,
		Isp:              node.Isp,
		IsActive:         node.IsActive == 1,
		StatusJSON:       []byte(node.Status),
		AllowIPs:         node.DecodeAllowIPs(),
		ReportNodeGroups: pbGroups,
	}}, nil
}

// UpdateReportNodeStatus 更新节点状态
func (this *ReportNodeService) UpdateReportNodeStatus(ctx context.Context, req *pb.UpdateReportNodeStatusRequest) (*pb.RPCSuccess, error) {
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeReport)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = validateClient(tx, nodeId, ctx)
	if err != nil {
		return nil, err
	}

	var status = &reporterconfigs.Status{}
	err = json.Unmarshal(req.StatusJSON, status)
	if err != nil {
		return nil, err
	}
	status.UpdatedAt = time.Now().Unix()

	p, ok := peer.FromContext(ctx)
	if ok {
		host, _, _ := net.SplitHostPort(p.Addr.String())
		if len(host) > 0 {
			status.IP = host

			result, _ := iplibrary.SharedLibrary.Lookup(host)
			if result != nil {
				status.Location = result.Summary()
				status.ISP = result.ISP
			}
		}
	}

	statusJSON, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}

	err = models.SharedReportNodeDAO.UpdateNodeStatus(tx, nodeId, statusJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindCurrentReportNodeConfig 获取当前节点信息
func (this *ReportNodeService) FindCurrentReportNodeConfig(ctx context.Context, req *pb.FindCurrentReportNodeConfigRequest) (*pb.FindCurrentReportNodeConfigResponse, error) {
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeReport)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = validateClient(tx, nodeId, ctx)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedReportNodeDAO.ComposeConfig(tx, nodeId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindCurrentReportNodeConfigResponse{ReportNodeJSON: configJSON}, nil
}

// FindReportNodeTasks 读取任务
func (this *ReportNodeService) FindReportNodeTasks(ctx context.Context, req *pb.FindReportNodeTasksRequest) (*pb.FindReportNodeTasksResponse, error) {
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeReport)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = validateClient(tx, nodeId, ctx)
	if err != nil {
		return nil, err
	}

	var result = &pb.FindReportNodeTasksResponse{}

	var ipTasks = []*reporterconfigs.IPTask{}

	// 所有的集群
	// TODO 将来支持NS节点
	clusters, err := models.SharedNodeClusterDAO.FindAllEnableClusters(tx)
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		if cluster.IsOn == 0 {
			continue
		}
		var clusterId = int64(cluster.Id)

		port, err := models.SharedServerDAO.FindFirstHTTPOrHTTPSPortWithClusterId(tx, clusterId)
		if err != nil {
			return nil, err
		}
		if port <= 0 {
			continue
		}

		// 读取所有IP地址
		addrList, err := models.SharedNodeIPAddressDAO.FindAllAccessibleIPAddressesWithClusterId(tx, nodeconfigs.NodeRoleNode, clusterId)
		if err != nil {
			return nil, err
		}
		for _, addr := range addrList {
			if addr.IsOn != 1 {
				continue
			}

			var addrIP = addr.Ip
			var backupIP = addr.DecodeBackupIP()
			if len(backupIP) > 0 {
				addrIP = backupIP
			}

			ipTasks = append(ipTasks, &reporterconfigs.IPTask{
				AddrId: int64(addr.Id),
				IP:     addrIP,
				Port:   port,
			})
		}
	}

	ipTasksJSON, err := json.Marshal(ipTasks)
	if err != nil {
		return nil, err
	}
	result.IpAddrTasksJSON = ipTasksJSON

	return result, nil
}

// FindLatestReportNodeVersion 取得最新的版本号
func (this *ReportNodeService) FindLatestReportNodeVersion(ctx context.Context, req *pb.FindLatestReportNodeVersionRequest) (*pb.FindLatestReportNodeVersionResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	return &pb.FindLatestReportNodeVersionResponse{
		Version: teaconst.ReportNodeVersion,
	}, nil
}

// CountAllReportNodeTasks 计算任务数量
func (this *ReportNodeService) CountAllReportNodeTasks(ctx context.Context, req *pb.CountAllReportNodeTasksRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var count int64
	var tx *dbs.Tx

	switch req.Type {
	case reporterconfigs.TaskTypeIPAddr:
		count, err = models.SharedNodeIPAddressDAO.CountAllAccessibleIPAddressesWithClusterId(tx, req.Role, req.NodeClusterId)
	}

	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListReportNodeTasks 列出单页任务
func (this *ReportNodeService) ListReportNodeTasks(ctx context.Context, req *pb.ListReportNodeTasksRequest) (*pb.ListReportNodeTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx
	switch req.Type {
	case reporterconfigs.TaskTypeIPAddr:
		port, err := models.SharedServerDAO.FindFirstHTTPOrHTTPSPortWithClusterId(tx, req.NodeClusterId)
		if err != nil {
			return nil, err
		}

		addrs, err := models.SharedNodeIPAddressDAO.ListAccessibleIPAddressesWithClusterId(tx, req.Role, req.NodeClusterId, req.Offset, req.Size)
		if err != nil {
			return nil, err
		}

		var pbTasks = []*pb.IPAddrReportTask{}
		for _, addr := range addrs {
			var addrIP = addr.Ip
			var backupIP = addr.DecodeBackupIP()
			if len(backupIP) > 0 {
				addrIP = backupIP
			}

			// 地址
			var pbAddr = &pb.NodeIPAddress{
				Id:          int64(addr.Id),
				NodeId:      int64(addr.NodeId),
				Name:        addr.Name,
				Ip:          addrIP,
				Description: addr.Description,
				CanAccess:   addr.CanAccess == 1,
				IsOn:        addr.IsOn == 1,
				IsUp:        addr.IsUp == 1,
				Role:        addr.Role,
			}
			var connectivity = addr.DecodeConnectivity()

			pbTasks = append(pbTasks, &pb.IPAddrReportTask{
				Ip:            addr.Ip,
				Port:          types.Int32(port),
				NodeIPAddress: pbAddr,
				CostMs:        float32(connectivity.CostMs),
				Level:         connectivity.Level,
				Connectivity:  float32(connectivity.Percent),
			})
		}
		return &pb.ListReportNodeTasksResponse{
			IpAddrReportTasks: pbTasks,
		}, nil
	}

	return &pb.ListReportNodeTasksResponse{}, nil
}

// UpdateReportNodeGlobalSetting 修改全局设置
func (this *ReportNodeService) UpdateReportNodeGlobalSetting(ctx context.Context, req *pb.UpdateReportNodeGlobalSetting) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedSysSettingDAO.UpdateSetting(tx, systemconfigs.SettingCodeReportNodeGlobalSetting, req.SettingJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ReadReportNodeGlobalSetting 读取全局设置
func (this *ReportNodeService) ReadReportNodeGlobalSetting(ctx context.Context, req *pb.ReadReportNodeGlobalSettingRequest) (*pb.ReadReportNodeGlobalSettingResponse, error) {
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeReport)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(tx, systemconfigs.SettingCodeReportNodeGlobalSetting)
	if err != nil {
		return nil, err
	}

	var setting = reporterconfigs.DefaultGlobalSetting()
	if len(valueJSON) > 0 {
		err = json.Unmarshal(valueJSON, setting)
		if err != nil {
			return nil, err
		}
	}

	// 重新编码
	valueJSON, err = json.Marshal(setting)
	if err != nil {
		return nil, err
	}
	return &pb.ReadReportNodeGlobalSettingResponse{
		SettingJSON: valueJSON,
	}, nil
}
