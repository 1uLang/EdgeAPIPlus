// Copyright 2021-2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus

package nameservers

import (
	"context"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io"
	"path/filepath"
	"time"
)

// NSNodeService 域名服务器节点服务
type NSNodeService struct {
	services.BaseService
}

// FindAllNSNodesWithNSClusterId 根据集群查找所有节点
func (this *NSNodeService) FindAllNSNodesWithNSClusterId(ctx context.Context, req *pb.FindAllNSNodesWithNSClusterIdRequest) (*pb.FindAllNSNodesWithNSClusterIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodes, err := models.SharedNSNodeDAO.FindAllEnabledNodesWithClusterId(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	pbNodes := []*pb.NSNode{}
	for _, node := range nodes {
		pbNodes = append(pbNodes, &pb.NSNode{
			Id:                  int64(node.Id),
			Name:                node.Name,
			IsOn:                node.IsOn,
			UniqueId:            node.UniqueId,
			Secret:              node.Secret,
			IsInstalled:         node.IsInstalled,
			InstallDir:          node.InstallDir,
			IsUp:                node.IsUp,
			ConnectedAPINodeIds: node.DecodeConnectedAPINodes(),
			NsCluster:           nil,
		})
	}
	return &pb.FindAllNSNodesWithNSClusterIdResponse{NsNodes: pbNodes}, nil
}

// CountAllNSNodes 所有可用的节点数量
func (this *NSNodeService) CountAllNSNodes(ctx context.Context, req *pb.CountAllNSNodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNSNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountAllNSNodesMatch 计算匹配的节点数量
func (this *NSNodeService) CountAllNSNodesMatch(ctx context.Context, req *pb.CountAllNSNodesMatchRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNSNodeDAO.CountAllEnabledNodesMatch(tx, req.NsClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListNSNodesMatch 列出单页节点
func (this *NSNodeService) ListNSNodesMatch(ctx context.Context, req *pb.ListNSNodesMatchRequest) (*pb.ListNSNodesMatchResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	nodes, err := models.SharedNSNodeDAO.ListAllEnabledNodesMatch(tx, req.NsClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbNodes := []*pb.NSNode{}
	for _, node := range nodes {
		// 安装信息
		installStatus, err := node.DecodeInstallStatus()
		if err != nil {
			return nil, err
		}
		installStatusResult := &pb.NodeInstallStatus{}
		if installStatus != nil {
			installStatusResult = &pb.NodeInstallStatus{
				IsRunning:  installStatus.IsRunning,
				IsFinished: installStatus.IsFinished,
				IsOk:       installStatus.IsOk,
				Error:      installStatus.Error,
				ErrorCode:  installStatus.ErrorCode,
				UpdatedAt:  installStatus.UpdatedAt,
			}
		}

		pbNodes = append(pbNodes, &pb.NSNode{
			Id:            int64(node.Id),
			Name:          node.Name,
			IsOn:          node.IsOn,
			UniqueId:      node.UniqueId,
			Secret:        node.Secret,
			IsActive:      node.IsActive,
			IsInstalled:   node.IsInstalled,
			InstallDir:    node.InstallDir,
			IsUp:          node.IsUp,
			StatusJSON:    node.Status,
			InstallStatus: installStatusResult,
			NsCluster:     nil,
		})
	}
	return &pb.ListNSNodesMatchResponse{NsNodes: pbNodes}, nil
}

// CountAllUpgradeNSNodesWithNSClusterId 计算需要升级的节点数量
func (this *NSNodeService) CountAllUpgradeNSNodesWithNSClusterId(ctx context.Context, req *pb.CountAllUpgradeNSNodesWithNSClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	deployFiles := installers.SharedDeployManager.LoadNSNodeFiles()
	total := int64(0)
	for _, deployFile := range deployFiles {
		count, err := models.SharedNSNodeDAO.CountAllLowerVersionNodesWithClusterId(tx, req.NsClusterId, deployFile.OS, deployFile.Arch, deployFile.Version)
		if err != nil {
			return nil, err
		}
		total += count
	}

	return this.SuccessCount(total)
}

// CreateNSNode 创建节点
func (this *NSNodeService) CreateNSNode(ctx context.Context, req *pb.CreateNSNodeRequest) (*pb.CreateNSNodeResponse, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodeId, err := models.SharedNSNodeDAO.CreateNode(tx, adminId, req.Name, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	// 增加认证相关
	if req.NodeLogin != nil {
		_, err = models.SharedNodeLoginDAO.CreateNodeLogin(tx, nodeconfigs.NodeRoleDNS, nodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
		if err != nil {
			return nil, err
		}
	}

	return &pb.CreateNSNodeResponse{
		NsNodeId: nodeId,
	}, nil
}

// DeleteNSNode 删除节点
func (this *NSNodeService) DeleteNSNode(ctx context.Context, req *pb.DeleteNSNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNSNodeDAO.DisableNSNode(tx, req.NsNodeId)
	if err != nil {
		return nil, err
	}

	// 删除任务
	err = models.SharedNodeTaskDAO.DeleteNodeTasks(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNSNode 查询单个节点信息
func (this *NSNodeService) FindNSNode(ctx context.Context, req *pb.FindNSNodeRequest) (*pb.FindNSNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	node, err := models.SharedNSNodeDAO.FindEnabledNSNode(tx, req.NsNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindNSNodeResponse{NsNode: nil}, nil
	}

	// 集群信息
	clusterName, err := models.SharedNSClusterDAO.FindEnabledNSClusterName(tx, int64(node.ClusterId))
	if err != nil {
		return nil, err
	}

	// 认证信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId)
	if err != nil {
		return nil, err
	}
	var respLogin *pb.NodeLogin = nil
	if login != nil {
		respLogin = &pb.NodeLogin{
			Id:     int64(login.Id),
			Name:   login.Name,
			Type:   login.Type,
			Params: login.Params,
		}
	}

	// 安装信息
	installStatus, err := node.DecodeInstallStatus()
	if err != nil {
		return nil, err
	}
	var installStatusResult = &pb.NodeInstallStatus{}
	if installStatus != nil {
		installStatusResult = &pb.NodeInstallStatus{
			IsRunning:  installStatus.IsRunning,
			IsFinished: installStatus.IsFinished,
			IsOk:       installStatus.IsOk,
			Error:      installStatus.Error,
			ErrorCode:  installStatus.ErrorCode,
			UpdatedAt:  installStatus.UpdatedAt,
		}
	}

	return &pb.FindNSNodeResponse{NsNode: &pb.NSNode{
		Id:               int64(node.Id),
		Name:             node.Name,
		StatusJSON:       node.Status,
		UniqueId:         node.UniqueId,
		Secret:           node.Secret,
		IsInstalled:      node.IsInstalled,
		InstallDir:       node.InstallDir,
		ApiNodeAddrsJSON: node.ApiNodeAddrs,
		NsCluster: &pb.NSCluster{
			Id:   int64(node.ClusterId),
			Name: clusterName,
		},
		InstallStatus: installStatusResult,
		IsOn:          node.IsOn,
		IsActive:      node.IsActive,
		NodeLogin:     respLogin,
	}}, nil
}

// UpdateNSNode 修改节点
func (this *NSNodeService) UpdateNSNode(ctx context.Context, req *pb.UpdateNSNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNSNodeDAO.UpdateNode(tx, req.NsNodeId, req.Name, req.NsClusterId, req.IsOn)
	if err != nil {
		return nil, err
	}

	// 登录信息
	if req.NodeLogin == nil {
		err = models.SharedNodeLoginDAO.DisableNodeLogins(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId)
		if err != nil {
			return nil, err
		}
	} else {
		if req.NodeLogin.Id > 0 {
			err = models.SharedNodeLoginDAO.UpdateNodeLogin(tx, req.NodeLogin.Id, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = models.SharedNodeLoginDAO.CreateNodeLogin(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// InstallNSNode 安装节点
func (this *NSNodeService) InstallNSNode(ctx context.Context, req *pb.InstallNSNodeRequest) (*pb.InstallNSNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	goman.New(func() {
		err = installers.SharedNSNodeQueue().InstallNodeProcess(req.NsNodeId, false)
		if err != nil {
			logs.Println("[RPC]install dns node:" + err.Error())
		}
	})

	return &pb.InstallNSNodeResponse{}, nil
}

// FindNSNodeInstallStatus 读取节点安装状态
func (this *NSNodeService) FindNSNodeInstallStatus(ctx context.Context, req *pb.FindNSNodeInstallStatusRequest) (*pb.FindNSNodeInstallStatusResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	installStatus, err := models.SharedNSNodeDAO.FindNodeInstallStatus(tx, req.NsNodeId)
	if err != nil {
		return nil, err
	}
	if installStatus == nil {
		return &pb.FindNSNodeInstallStatusResponse{InstallStatus: nil}, nil
	}

	pbInstallStatus := &pb.NodeInstallStatus{
		IsRunning:  installStatus.IsRunning,
		IsFinished: installStatus.IsFinished,
		IsOk:       installStatus.IsOk,
		Error:      installStatus.Error,
		ErrorCode:  installStatus.ErrorCode,
		UpdatedAt:  installStatus.UpdatedAt,
	}
	return &pb.FindNSNodeInstallStatusResponse{InstallStatus: pbInstallStatus}, nil
}

// UpdateNSNodeIsInstalled 修改节点安装状态
func (this *NSNodeService) UpdateNSNodeIsInstalled(ctx context.Context, req *pb.UpdateNSNodeIsInstalledRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNSNodeDAO.UpdateNodeIsInstalled(tx, req.NsNodeId, req.IsInstalled)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNSNodeStatus 更新节点状态
func (this *NSNodeService) UpdateNSNodeStatus(ctx context.Context, req *pb.UpdateNSNodeStatusRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	if req.NodeId > 0 {
		nodeId = req.NodeId
	}

	if nodeId <= 0 {
		return nil, errors.New("'nodeId' should be greater than 0")
	}

	var tx = this.NullTx()

	// 修改时间戳
	var nodeStatus = &nodeconfigs.NodeStatus{}
	err = json.Unmarshal(req.StatusJSON, nodeStatus)
	if err != nil {
		return nil, errors.New("decode node status json failed: " + err.Error())
	}
	nodeStatus.UpdatedAt = time.Now().Unix()

	// 保存
	err = models.SharedNSNodeDAO.UpdateNodeStatus(tx, nodeId, nodeStatus)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindCurrentNSNodeConfig 获取当前节点信息
func (this *NSNodeService) FindCurrentNSNodeConfig(ctx context.Context, req *pb.FindCurrentNSNodeConfigRequest) (*pb.FindCurrentNSNodeConfigResponse, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	config, err := models.SharedNSNodeDAO.ComposeNodeConfig(tx, nodeId)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return &pb.FindCurrentNSNodeConfigResponse{NsNodeJSON: nil}, nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindCurrentNSNodeConfigResponse{NsNodeJSON: configJSON}, nil
}

// CheckNSNodeLatestVersion 检查新版本
func (this *NSNodeService) CheckNSNodeLatestVersion(ctx context.Context, req *pb.CheckNSNodeLatestVersionRequest) (*pb.CheckNSNodeLatestVersionResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	deployFiles := installers.SharedDeployManager.LoadNSNodeFiles()
	for _, file := range deployFiles {
		if file.OS == req.Os && file.Arch == req.Arch && stringutil.VersionCompare(file.Version, req.CurrentVersion) > 0 {
			return &pb.CheckNSNodeLatestVersionResponse{
				HasNewVersion: true,
				NewVersion:    file.Version,
			}, nil
		}
	}
	return &pb.CheckNSNodeLatestVersionResponse{HasNewVersion: false}, nil
}

// FindLatestNSNodeVersion 获取NS节点最新版本
func (this *NSNodeService) FindLatestNSNodeVersion(ctx context.Context, req *pb.FindLatestNSNodeVersionRequest) (*pb.FindLatestNSNodeVersionResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.FindLatestNSNodeVersionResponse{Version: teaconst.DNSNodeVersion}, nil
}

// DownloadNSNodeInstallationFile 下载最新DNS节点安装文件
func (this *NSNodeService) DownloadNSNodeInstallationFile(ctx context.Context, req *pb.DownloadNSNodeInstallationFileRequest) (*pb.DownloadNSNodeInstallationFileResponse, error) {
	nodeId, err := this.ValidateNSNode(ctx)
	if err != nil {
		return nil, err
	}

	var file = installers.SharedDeployManager.FindNSNodeFile(req.Os, req.Arch)
	if file == nil {
		return &pb.DownloadNSNodeInstallationFileResponse{}, nil
	}

	sum, err := file.Sum()
	if err != nil {
		return nil, err
	}

	data, offset, err := file.Read(req.ChunkOffset)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// 增加下载速度监控
	installers.SharedUpgradeLimiter.UpdateNodeBytes(nodeconfigs.NodeRoleDNS, nodeId, int64(len(data)))

	return &pb.DownloadNSNodeInstallationFileResponse{
		Sum:       sum,
		Offset:    offset,
		ChunkData: data,
		Version:   file.Version,
		Filename:  filepath.Base(file.Path),
	}, nil
}

// UpdateNSNodeConnectedAPINodes 更改节点连接的API节点信息
func (this *NSNodeService) UpdateNSNodeConnectedAPINodes(ctx context.Context, req *pb.UpdateNSNodeConnectedAPINodesRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, _, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNSNodeDAO.UpdateNodeConnectedAPINodes(tx, nodeId, req.ApiNodeIds)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return this.Success()
}

// UpdateNSNodeLogin 修改节点登录信息
func (this *NSNodeService) UpdateNSNodeLogin(ctx context.Context, req *pb.UpdateNSNodeLoginRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.NodeLogin.Id <= 0 {
		_, err := models.SharedNodeLoginDAO.CreateNodeLogin(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedNodeLoginDAO.UpdateNodeLogin(tx, req.NodeLogin.Id, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)

	return this.Success()
}

// StartNSNode 启动节点
func (this *NSNodeService) StartNSNode(ctx context.Context, req *pb.StartNSNodeRequest) (*pb.StartNSNodeResponse, error) {
	// 校验节点
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	err = installers.SharedNSNodeQueue().StartNode(req.NsNodeId)
	if err != nil {
		return &pb.StartNSNodeResponse{
			IsOk:  false,
			Error: err.Error(),
		}, nil
	}

	// 修改状态
	var tx = this.NullTx()
	err = models.SharedNSNodeDAO.UpdateNodeActive(tx, req.NsNodeId, true)
	if err != nil {
		return nil, err
	}

	return &pb.StartNSNodeResponse{IsOk: true}, nil
}

// StopNSNode 停止节点
func (this *NSNodeService) StopNSNode(ctx context.Context, req *pb.StopNSNodeRequest) (*pb.StopNSNodeResponse, error) {
	// 校验节点
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	err = installers.SharedNSNodeQueue().StopNode(req.NsNodeId)
	if err != nil {
		return &pb.StopNSNodeResponse{
			IsOk:  false,
			Error: err.Error(),
		}, nil
	}

	// 修改状态
	var tx = this.NullTx()
	err = models.SharedNSNodeDAO.UpdateNodeActive(tx, req.NsNodeId, false)
	if err != nil {
		return nil, err
	}

	return &pb.StopNSNodeResponse{IsOk: true}, nil
}

// FindNSNodeDDoSProtection 获取集群的DDoS设置
func (this *NSNodeService) FindNSNodeDDoSProtection(ctx context.Context, req *pb.FindNSNodeDDoSProtectionRequest) (*pb.FindNSNodeDDoSProtectionResponse, error) {
	var nodeId = req.NsNodeId
	var isFromNode = false

	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		// 检查是否来自节点
		currentNodeId, err2 := this.ValidateNSNode(ctx)
		if err2 != nil {
			return nil, err
		}

		if nodeId > 0 && currentNodeId != nodeId {
			return nil, errors.New("invalid 'nsNodeId'")
		}

		nodeId = currentNodeId
		isFromNode = true
	}

	var tx *dbs.Tx
	ddosProtection, err := models.SharedNSNodeDAO.FindNodeDDoSProtection(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if ddosProtection == nil {
		ddosProtection = ddosconfigs.DefaultProtectionConfig()
	}

	// 组合父级节点配置
	// 只有从节点读取配置时才需要组合
	if isFromNode {
		clusterId, err := models.SharedNSNodeDAO.FindNodeClusterId(tx, nodeId)
		if err != nil {
			return nil, err
		}

		if clusterId > 0 {
			clusterDDoSProtection, err := models.SharedNSClusterDAO.FindClusterDDoSProtection(tx, clusterId)
			if err != nil {
				return nil, err
			}
			if clusterDDoSProtection == nil {
				clusterDDoSProtection = ddosconfigs.DefaultProtectionConfig()
			}

			clusterDDoSProtection.Merge(ddosProtection)
			ddosProtection = clusterDDoSProtection
		}
	}

	ddosProtectionJSON, err := json.Marshal(ddosProtection)
	if err != nil {
		return nil, err
	}

	var result = &pb.FindNSNodeDDoSProtectionResponse{
		DdosProtectionJSON: ddosProtectionJSON,
	}

	return result, nil
}

// UpdateNSNodeDDoSProtection 修改集群的DDoS设置
func (this *NSNodeService) UpdateNSNodeDDoSProtection(ctx context.Context, req *pb.UpdateNSNodeDDoSProtectionRequest) (*pb.RPCSuccess, error) {
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
	err = models.SharedNSNodeDAO.UpdateNodeDDoSProtection(tx, req.NsNodeId, ddosProtection)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSNodeAPIConfig 查找单个节点的API相关配置
func (this *NSNodeService) FindNSNodeAPIConfig(ctx context.Context, req *pb.FindNSNodeAPIConfigRequest) (*pb.FindNSNodeAPIConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	node, err := models.SharedNSNodeDAO.FindNodeAPIConfig(tx, req.NsNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindNSNodeAPIConfigResponse{
			ApiNodeAddrsJSON: nil,
		}, nil
	}

	return &pb.FindNSNodeAPIConfigResponse{
		ApiNodeAddrsJSON: node.ApiNodeAddrs,
	}, nil
}

// UpdateNSNodeAPIConfig 修改某个节点的API相关配置
func (this *NSNodeService) UpdateNSNodeAPIConfig(ctx context.Context, req *pb.UpdateNSNodeAPIConfigRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var apiNodeAddrs = []*serverconfigs.NetworkAddressConfig{}
	if len(req.ApiNodeAddrsJSON) > 0 {
		err = json.Unmarshal(req.ApiNodeAddrsJSON, &apiNodeAddrs)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedNSNodeDAO.UpdateNodeAPIConfig(tx, req.NsNodeId, apiNodeAddrs)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
