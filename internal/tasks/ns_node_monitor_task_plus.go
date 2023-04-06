package tasks

import (
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strings"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewNSNodeMonitorTask(1 * time.Minute).Start()
		})
	})
}

// 节点启动尝试
type nsNodeStartingTry struct {
	count     int
	timestamp int64
}

// NSNodeMonitorTask 边缘节点监控任务
type NSNodeMonitorTask struct {
	BaseTask

	ticker *time.Ticker

	inactiveMap map[string]int  // cluster@nodeId => count
	notifiedMap map[int64]int64 // nodeId => timestamp

	recoverMap map[int64]*nsNodeStartingTry // nodeId => *nsNodeStartingTry
}

func NewNSNodeMonitorTask(duration time.Duration) *NSNodeMonitorTask {
	return &NSNodeMonitorTask{
		ticker:      time.NewTicker(duration),
		inactiveMap: map[string]int{},
		notifiedMap: map[int64]int64{},
		recoverMap:  map[int64]*nsNodeStartingTry{},
	}
}

func (this *NSNodeMonitorTask) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("NS_NODE_MONITOR", err.Error())
		}
	}
}

func (this *NSNodeMonitorTask) Loop() error {
	// 检查是否为主节点
	if !models.SharedAPINodeDAO.CheckAPINodeIsPrimaryWithoutErr() {
		return nil
	}

	clusters, err := models.SharedNSClusterDAO.FindAllEnabledClusters(nil)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		err := this.monitorCluster(cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *NSNodeMonitorTask) monitorCluster(cluster *models.NSCluster) error {
	var clusterId = int64(cluster.Id)

	// 检查离线节点
	inactiveNodes, err := models.SharedNSNodeDAO.FindAllNotifyingInactiveNodesWithClusterId(nil, clusterId)
	if err != nil {
		return err
	}

	// 尝试自动远程启动
	if cluster.AutoRemoteStart {
		var nodeQueue = installers.NewNSNodeQueue()
		for _, node := range inactiveNodes {
			var nodeId = int64(node.Id)
			tryInfo, ok := this.recoverMap[nodeId]
			if !ok {
				tryInfo = &nsNodeStartingTry{
					count:     1,
					timestamp: time.Now().Unix(),
				}
				this.recoverMap[nodeId] = tryInfo
			} else {
				if tryInfo.count >= 3 /** 3次 **/ { // N 秒内超过 M 次就暂时不再重新尝试，防止阻塞当前任务
					if tryInfo.timestamp+10*60 /** 10 分钟 **/ > time.Now().Unix() {
						continue
					}
					tryInfo.timestamp = time.Now().Unix()
					tryInfo.count = 0
				}
				tryInfo.count++
			}

			// TODO 如果用户手工安装的位置不在标准位置，需要节点自身记住最近启动的位置
			err = nodeQueue.StartNode(nodeId)
			if err != nil {
				if !installers.IsGrantError(err) {
					_ = models.SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleDNS, nodeId, 0, 0, models.LevelError, "NODE", "start node from remote API failed: "+err.Error(), time.Now().Unix(), "", nil)
				}
			} else {
				_ = models.SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleDNS, nodeId, 0, 0, models.LevelSuccess, "NODE", "start node from remote API successfully", time.Now().Unix(), "", nil)
			}
		}
	}

	var nodeMap = map[int64]*models.NSNode{}
	for _, node := range inactiveNodes {
		var nodeId = int64(node.Id)
		nodeMap[nodeId] = node
		this.inactiveMap[types.String(clusterId)+"@"+types.String(nodeId)]++
	}

	const maxInactiveTries = 5

	// 处理现有的离线状态
	for key, count := range this.inactiveMap {
		var pieces = strings.Split(key, "@")
		if pieces[0] != types.String(clusterId) {
			continue
		}
		var nodeId = types.Int64(pieces[1])
		node, ok := nodeMap[nodeId]
		if ok {
			// 连续 N 次离线发送通知
			// 同时也要确保两次发送通知的时间不会过近
			if count >= maxInactiveTries && time.Now().Unix()-this.notifiedMap[nodeId] > 3600 {
				this.inactiveMap[key] = 0
				this.notifiedMap[nodeId] = time.Now().Unix()

				var subject = "DNS节点\"" + node.Name + "\"已处于离线状态"
				var msg = "DNS节点\"" + node.Name + "\"已处于离线状态"
				err = models.SharedMessageDAO.CreateNodeMessage(nil, nodeconfigs.NodeRoleDNS, clusterId, int64(node.Id), models.MessageTypeNSNodeInactive, models.LevelError, subject, msg, nil, false)
				if err != nil {
					return err
				}

				// 修改在线状态
				err = models.SharedNSNodeDAO.UpdateNodeStatusIsNotified(nil, int64(node.Id))
				if err != nil {
					return err
				}
			}
		} else {
			delete(this.inactiveMap, key)
		}
	}

	// TODO 检查恢复连接

	// 检查CPU、内存、磁盘不足节点，而且离线的节点不再重复提示
	// TODO 需要实现

	// TODO 检查53/tcp、53/udp是否能够访问

	return nil
}
