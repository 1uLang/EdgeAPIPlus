// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/reporterconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

func init() {
	var ticker = time.NewTicker(1 * time.Minute)
	dbs.OnReadyDone(func() {
		go func() {
			for range ticker.C {
				err := SharedNodeIPAddressDAO.loopTask(nil, nodeconfigs.NodeRoleNode)
				if err != nil {
					remotelogs.Error("NodeIPAddressDAO.LoopTasks", err.Error())
				}
			}
		}()
	})
}

// FireThresholds 触发阈值
func (this *NodeIPAddressDAO) FireThresholds(tx *dbs.Tx, role nodeconfigs.NodeRole, nodeId int64) error {
	// 节点是否存在
	node, err := SharedNodeDAO.FindEnabledBasicNode(tx, nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return nil
	}
	if node.IsOn == 0 {
		return nil
	}

	var clusterId = int64(node.ClusterId)
	var groupId = int64(node.GroupId)

	// 检查集群
	b, err := SharedNodeClusterDAO.ExistsEnabledCluster(tx, clusterId)
	if err != nil {
		return err
	}
	if !b {
		return nil
	}

	ones, err := this.Query(tx).
		Attr("state", NodeIPAddressStateEnabled).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("canAccess", true).
		Attr("isOn", true).
		FindAll()
	if err != nil {
		return err
	}
	if len(ones) == 0 {
		return nil
	}
	for _, one := range ones {
		addr := one.(*NodeIPAddress)
		thresholds, err := SharedNodeIPAddressThresholdDAO.FindAllEnabledThresholdsWithAddrId(tx, int64(addr.Id))
		if err != nil {
			return err
		}

		var oldIsUp = addr.IsUp
		var oldBackupThresholdId = addr.BackupThresholdId
		var hasMatched = false

		for _, threshold := range thresholds {
			matched, err := this.runThreshold(tx, role, clusterId, groupId, nodeId, addr, threshold)
			if err != nil {
				return err
			}
			if matched {
				hasMatched = matched
			}
		}

		// 提示还原
		if !hasMatched && (addr.IsUp != oldIsUp || (oldBackupThresholdId > 0 && addr.BackupThresholdId == 0)) {
			err = SharedMessageDAO.CreateNodeMessage(tx, role, clusterId, nodeId, MessageTypeIPAddrUp, MessageLevelSuccess, "节点IP'"+addr.Ip+"'没有匹配到任何阈值，状态已还原", "节点IP'"+addr.Ip+"'没有匹配到任何阈值，状态已还原。", maps.Map{
				"addrId": addr.Id,
			}.AsJSON(), true)
			if err != nil {
				return err
			}
			err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, int64(addr.Id), "没有匹配到任何阈值，状态已还原。")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 执行单个阈值
func (this *NodeIPAddressDAO) runThreshold(tx *dbs.Tx, role string, clusterId int64, groupId int64, nodeId int64, addr *NodeIPAddress, threshold *NodeIPAddressThreshold) (matched bool, err error) {
	var addrId = int64(addr.Id)

	var items = threshold.DecodeItems()
	if len(items) == 0 {
		return false, nil
	}

	var actions = threshold.DecodeActions()
	if len(actions) == 0 {
		actions = []*nodeconfigs.IPAddressThresholdActionConfig{
			{
				Action: nodeconfigs.IPAddressThresholdActionNotify,
			},
		}
	}

	var summaryList = []string{}

	for _, item := range items {
		if item.Item != nodeconfigs.IPAddressThresholdItemNodeHealthCheck && (item.Value < 0 || item.Duration <= 0) {
			continue
		}

		var value = float64(0)
		var summary = ""
		var op = nodeconfigs.FindNodeValueOperatorName(item.Operator)
		switch item.Item {
		case nodeconfigs.IPAddressThresholdItemNodeAvgRequests:
			value, err = SharedNodeValueDAO.SumNodeValues(tx, role, nodeId, nodeconfigs.NodeValueItemRequests, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value / 60)
			summary = "节点平均请求数：" + types.String(value) + "/s，阈值：" + op + " " + types.String(item.Value) + "/s"
		case nodeconfigs.IPAddressThresholdItemNodeAvgTrafficOut:
			value, err = SharedNodeValueDAO.SumNodeValues(tx, role, nodeId, nodeconfigs.NodeValueItemTrafficOut, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
			summary = "节点平均下行流量：" + types.String(value) + "MB/s，阈值：" + op + " " + types.String(item.Value) + "MB/s"
		case nodeconfigs.IPAddressThresholdItemNodeAvgTrafficIn:
			value, err = SharedNodeValueDAO.SumNodeValues(tx, role, nodeId, nodeconfigs.NodeValueItemTrafficIn, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
			summary = "节点平均上行流量：" + types.String(value) + "MB/s，阈值：" + op + " " + types.String(item.Value) + "MB/s"
		case nodeconfigs.IPAddressThresholdItemGroupAvgRequests:
			if groupId <= 0 {
				continue
			}
			value, err = SharedNodeValueDAO.SumNodeGroupValues(tx, role, groupId, nodeconfigs.NodeValueItemRequests, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value / 60)
			summary = "节点分组平均请求数：" + types.String(value) + "/s，阈值：" + op + " " + types.String(item.Value) + "/s"
		case nodeconfigs.IPAddressThresholdItemGroupAvgTrafficOut:
			if groupId <= 0 {
				continue
			}
			value, err = SharedNodeValueDAO.SumNodeGroupValues(tx, role, groupId, nodeconfigs.NodeValueItemTrafficOut, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
			summary = "节点分组平均下行流量：" + types.String(value) + "MB/s，阈值：" + op + " " + types.String(item.Value) + "MB/s"
		case nodeconfigs.IPAddressThresholdItemGroupAvgTrafficIn:
			if groupId <= 0 {
				continue
			}
			value, err = SharedNodeValueDAO.SumNodeGroupValues(tx, role, groupId, nodeconfigs.NodeValueItemTrafficIn, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
			summary = "节点分组平均上行流量：" + types.String(value) + "MB/s，阈值：" + op + " " + types.String(item.Value) + "MB/s"
		case nodeconfigs.IPAddressThresholdItemClusterAvgRequests:
			value, err = SharedNodeValueDAO.SumNodeClusterValues(tx, role, clusterId, nodeconfigs.NodeValueItemRequests, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value / 60)
			summary = "集群平均请求数：" + types.String(value) + "/s，阈值：" + op + " " + types.String(item.Value) + "/s"
		case nodeconfigs.IPAddressThresholdItemClusterAvgTrafficOut:
			value, err = SharedNodeValueDAO.SumNodeClusterValues(tx, role, clusterId, nodeconfigs.NodeValueItemTrafficOut, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
			summary = "集群平均下行流量：" + types.String(value) + "MB/s，阈值：" + op + " " + types.String(item.Value) + "MB/s"
		case nodeconfigs.IPAddressThresholdItemClusterAvgTrafficIn:
			value, err = SharedNodeValueDAO.SumNodeClusterValues(tx, role, clusterId, nodeconfigs.NodeValueItemTrafficIn, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(item.Duration), item.DurationUnit)
			value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
			summary = "集群平均上行流量：" + types.String(value) + "MB/s，阈值：" + op + " " + types.String(item.Value) + "MB/s"

		case nodeconfigs.IPAddressThresholdItemConnectivity:
			var groups = item.Options.GetSlice("groups")
			if len(groups) > 0 {
				var lastValue float64 = -1
				for _, group := range groups {
					var groupMap = maps.NewMap(group)
					var groupId = groupMap.GetInt64("id")
					if groupId <= 0 {
						continue
					}
					v, err := SharedReportResultDAO.FindConnectivityWithTargetPercent(tx, reporterconfigs.TaskTypeIPAddr, int64(addr.Id), groupId)
					if err != nil {
						return false, err
					}
					if lastValue < 0 || v < lastValue {
						// 取最小的值
						lastValue = v
					}
				}
				value = lastValue
			} else {
				value, err = SharedReportResultDAO.FindConnectivityWithTargetPercent(tx, reporterconfigs.TaskTypeIPAddr, int64(addr.Id), 0)
				if err != nil {
					return false, err
				}
			}
			summary = "连通性：" + fmt.Sprintf("%.2f", value) + "%，阈值：" + op + " " + types.String(item.Value) + "%"
		case nodeconfigs.IPAddressThresholdItemNodeHealthCheck:
			item.Operator = nodeconfigs.NodeValueOperatorEq
			isHealthy, err := SharedNodeIPAddressDAO.FindAddressIsHealthy(tx, addrId)
			if err != nil {
				return false, err
			}
			if !isHealthy {
				value = 0
			} else {
				value = 1
			}
			if item.Value == 0 {
				summary = "节点健康检查失败"
			} else {
				summary = "节点健康检查成功"
			}
		default:
			// TODO 支持更多阈值参数
			err = errors.New("threshold item '" + item.Item + "' not supported")
		}
		if err != nil {
			return false, err
		}

		if !nodeconfigs.CompareNodeValue(item.Operator, value, item.Value) {
			// 设置不匹配
			err = SharedNodeIPAddressThresholdDAO.UpdateThresholdIsMatched(tx, int64(threshold.Id), false)
			if err != nil {
				return false, err
			}

			// 还原动作
			for _, action := range actions {
				switch action.Action {
				case nodeconfigs.IPAddressThresholdActionUp:
					if addr.IsUp == 1 {
						addr.IsUp = 0
						err = this.UpdateAddressIsUp(tx, int64(addr.Id), false)
						if err != nil {
							return false, err
						}
					}
				case nodeconfigs.IPAddressThresholdActionDown:
					if addr.IsUp == 0 {
						addr.IsUp = 1
						err = this.UpdateAddressIsUp(tx, int64(addr.Id), true)
						if err != nil {
							return false, err
						}
					}
				case nodeconfigs.IPAddressThresholdActionSwitch:
					if int64(addr.BackupThresholdId) == int64(threshold.Id) {
						err = this.UpdateAddressBackupIP(tx, addrId, 0, "")
						if err != nil {
							return false, err
						}
						addr.BackupThresholdId = 0
					}
				case nodeconfigs.IPAddressThresholdActionWebHook:
					if threshold.IsMatched == 1 {
						if action.Options != nil {
							var url = action.Options.GetString("url")
							if len(url) > 0 {
								err = this.doWebHookAction(tx, role, clusterId, groupId, nodeId, addr, threshold, false, summaryList, url)
								if err != nil {
									// 记录日志
									err1 := SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "完成WebHook动作："+err.Error())
									if err1 != nil {
										return false, err1
									}

									return false, err
								} else {
									// 记录日志
									err1 := SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "完成WebHook动作")
									if err1 != nil {
										return false, err1
									}
								}
							}
						}
					}
				}
			}

			return false, nil
		}
		summaryList = append(summaryList, summary)
	}

	// 设置匹配
	err = SharedNodeIPAddressThresholdDAO.UpdateThresholdIsMatched(tx, int64(threshold.Id), true)
	if err != nil {
		return false, err
	}

	for _, action := range actions {
		switch action.Action {
		case nodeconfigs.IPAddressThresholdActionNotify:
			err := this.doNotifyAction(tx, role, clusterId, nodeId, addr, threshold, summaryList)
			if err != nil {
				return true, err
			}
		case nodeconfigs.IPAddressThresholdActionUp:
			err := this.doUpAction(tx, role, clusterId, nodeId, addr, threshold, summaryList)
			if err != nil {
				return true, err
			}
		case nodeconfigs.IPAddressThresholdActionDown:
			err := this.doDownAction(tx, role, clusterId, nodeId, addr, threshold, summaryList)
			if err != nil {
				return true, err
			}
		case nodeconfigs.IPAddressThresholdActionSwitch:
			err := this.doSwitchAction(tx, role, clusterId, nodeId, addr, threshold, action, summaryList)
			if err != nil {
				return true, err
			}
		case nodeconfigs.IPAddressThresholdActionWebHook:
			if threshold.IsMatched == 0 {
				if action.Options != nil {
					var url = action.Options.GetString("url")
					if len(url) > 0 {
						err = this.doWebHookAction(tx, role, clusterId, groupId, nodeId, addr, threshold, true, summaryList, url)
						if err != nil {
							// 记录日志
							err1 := SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "完成WebHook动作："+err.Error())
							if err1 != nil {
								return true, err1
							}

							return true, err
						} else {
							// 记录日志
							err1 := SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "完成WebHook动作")
							if err1 != nil {
								return true, err1
							}
						}
					}
				}
			}
		}
	}
	return true, nil
}

// 执行任务
func (this *NodeIPAddressDAO) loopTask(tx *dbs.Tx, role string) error {
	// 检查上次运行时间，防止重复运行
	settingKey := "nodeIPAddressDAOLoopTask"
	timestamp := time.Now().Unix()
	seconds := int64(60)
	c, err := SharedSysSettingDAO.CompareInt64Setting(nil, settingKey, timestamp-seconds)
	if err != nil {
		return err
	}
	if c > 0 {
		return nil
	}

	// 记录时间
	err = SharedSysSettingDAO.UpdateSetting(nil, settingKey, []byte(numberutils.FormatInt64(timestamp)))
	if err != nil {
		return err
	}

	// 查找所有任务
	addrs, err := this.Query(tx).
		Attr("role", role).
		Result("DISTINCT nodeId").
		Where("id IN (SELECT addressId FROM " + SharedNodeIPAddressThresholdDAO.Table + " WHERE state=1)").
		FindAll()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		var nodeId = int64(addr.(*NodeIPAddress).NodeId)
		err := this.FireThresholds(tx, role, nodeId)
		if err != nil {
			return err
		}
	}

	return nil
}

// 上线
func (this *NodeIPAddressDAO) doUpAction(tx *dbs.Tx, role string, clusterId int64, nodeId int64, addr *NodeIPAddress, threshold *NodeIPAddressThreshold, summaryList []string) error {
	if addr.IsUp == 1 {
		return nil
	}
	_, err := this.Query(tx).
		Pk(addr.Id).
		Set("isUp", true).
		Update()
	if err != nil {
		return err
	}

	// 增加日志
	var description = "触发阈值：" + strings.Join(summaryList, "；") + "。"
	err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, int64(addr.Id), "[上线]"+description)
	if err != nil {
		return err
	}

	err = SharedMessageDAO.CreateNodeMessage(tx, role, clusterId, nodeId, MessageTypeIPAddrUp, MessageLevelSuccess, "[上线]节点IP'"+addr.Ip+"'因为达到阈值而上线", "[上线]节点IP'"+addr.Ip+"'因为达到阈值而上线。"+description, maps.Map{
		"addrId": addr.Id,
	}.AsJSON(), true)
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, int64(addr.Id))
	if err != nil {
		return err
	}

	return nil
}

// 下线
func (this *NodeIPAddressDAO) doDownAction(tx *dbs.Tx, role string, clusterId int64, nodeId int64, addr *NodeIPAddress, threshold *NodeIPAddressThreshold, summaryList []string) error {
	if addr.IsUp == 0 {
		return nil
	}
	_, err := this.Query(tx).
		Pk(addr.Id).
		Set("isUp", false).
		Update()
	if err != nil {
		return err
	}

	// 增加日志
	var description = "触发阈值：" + strings.Join(summaryList, "；") + "。"

	err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, int64(addr.Id), "[下线]"+description)
	if err != nil {
		return err
	}

	err = SharedMessageDAO.CreateNodeMessage(tx, role, clusterId, nodeId, MessageTypeIPAddrDown, MessageLevelWarning, "[下线]节点IP'"+addr.Ip+"'因为达到阈值而下线", "[下线]节点IP'"+addr.Ip+"'因为达到阈值而下线。"+description, maps.Map{
		"addrId": addr.Id,
	}.AsJSON(), true)
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, int64(addr.Id))
	if err != nil {
		return err
	}

	return nil
}

// 切换
func (this *NodeIPAddressDAO) doSwitchAction(tx *dbs.Tx, role string, clusterId int64, nodeId int64, addr *NodeIPAddress, threshold *NodeIPAddressThreshold, action *nodeconfigs.IPAddressThresholdActionConfig, summaryList []string) error {
	var ipStrings = []string{}
	if action.Options != nil {
		var ips = action.Options.GetSlice("ips")
		if len(ips) > 0 {
			for _, ip := range ips {
				ipStrings = append(ipStrings, types.String(ip))
			}
		}
	}
	if len(ipStrings) > 0 {
		// 检查是否有变化
		if int64(addr.BackupThresholdId) == int64(threshold.Id) && lists.ContainsString(ipStrings, addr.BackupIP) {
			return nil
		}

		var ip = ""
		if len(ipStrings) == 1 {
			ip = types.String(ipStrings[0])
		} else {
			ip = types.String(ipStrings[rands.Int(0, len(ipStrings)-1)])
		}
		err := this.UpdateAddressBackupIP(tx, int64(addr.Id), int64(threshold.Id), ip)
		if err != nil {
			return err
		}

		// 增加日志
		var description = "触发阈值：" + strings.Join(summaryList, "；") + "。"

		err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, int64(addr.Id), description)
		if err != nil {
			return err
		}

		err = SharedMessageDAO.CreateNodeMessage(tx, role, clusterId, nodeId, MessageTypeIPAddrDown, MessageLevelWarning, "[切换]节点IP'"+addr.Ip+"'因为达到阈值而切换到备用IP", "[切换]节点IP'"+addr.Ip+"'因为达到阈值而切换到备用IP '"+ip+"'。"+description, maps.Map{
			"addrId": addr.Id,
		}.AsJSON(), true)
		if err != nil {
			return err
		}

		err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, int64(addr.Id), "[切换]触发阈值："+strings.Join(summaryList, "；")+"。")
		if err != nil {
			return err
		}
	} else {
		if int64(addr.BackupThresholdId) == int64(threshold.Id) {
			err := this.UpdateAddressBackupIP(tx, int64(addr.Id), 0, "")
			if err != nil {
				return err
			}
			addr.BackupThresholdId = 0
		}
	}

	return nil
}

// 通知
func (this *NodeIPAddressDAO) doNotifyAction(tx *dbs.Tx, role string, clusterId int64, nodeId int64, addr *NodeIPAddress, threshold *NodeIPAddressThreshold, summaryList []string) error {
	// 在一个小时内不重复提醒
	if time.Now().Unix()-int64(threshold.NotifiedAt) < 3600 {
		return nil
	}

	err := SharedMessageDAO.CreateNodeMessage(tx, role, clusterId, nodeId, MessageTypeIPAddrUp, MessageLevelSuccess, "[通知]节点IP'"+addr.Ip+"'达到阈值", "[通知]节点IP'"+addr.Ip+"'达到阈值。触发阈值："+strings.Join(summaryList, "；")+"。", maps.Map{
		"addrId": addr.Id,
	}.AsJSON(), true)
	if err != nil {
		return err
	}
	err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, int64(addr.Id), "触发阈值："+strings.Join(summaryList, "；")+"。")
	if err != nil {
		return err
	}

	return nil
}

// 调用WebHook
func (this *NodeIPAddressDAO) doWebHookAction(tx *dbs.Tx, role string, clusterId int64, groupId int64, nodeId int64, addr *NodeIPAddress, threshold *NodeIPAddressThreshold, isMatched bool, summaryList []string, url string) error {
	var addressId = int64(addr.Id)
	var ip = addr.Ip

	var resultString = ""
	if isMatched {
		resultString = "true"
	} else {
		resultString = "false"
	}
	var args = "role=" + role + "&clusterId=" + types.String(clusterId) + "&groupId=" + types.String(groupId) + "&nodeId=" + types.String(nodeId) + "&addressId=" + types.String(addressId) + "&ip=" + ip + "&result=" + resultString
	var hasQuestionMark = strings.Contains(url, "?")
	if hasQuestionMark {
		url += "&" + args
	} else {
		url += "?" + args
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)

	// TODO 可以重试
	var client = utils.SharedHttpClient(10 * time.Second)
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	data = bytes.TrimSpace(data)

	if len(data) == 0 {
		return nil
	}

	// 读取动作
	var actions = []*nodeconfigs.IPAddressThresholdActionConfig{}
	err = json.Unmarshal(data, &actions)
	if err != nil {
		return err
	}

	// 执行动作
	if len(summaryList) == 0 {
		summaryList = []string{"无"}
	}

	for index, summary := range summaryList {
		summaryList[index] = "[WebHook]" + summary
	}

	for _, action := range actions {
		switch action.Action {
		case nodeconfigs.IPAddressThresholdActionNotify:
			err := this.doNotifyAction(tx, role, clusterId, nodeId, addr, threshold, summaryList)
			if err != nil {
				return err
			}
		case nodeconfigs.IPAddressThresholdActionUp:
			err := this.doUpAction(tx, role, clusterId, nodeId, addr, threshold, summaryList)
			if err != nil {
				return err
			}
		case nodeconfigs.IPAddressThresholdActionDown:
			err := this.doDownAction(tx, role, clusterId, nodeId, addr, threshold, summaryList)
			if err != nil {
				return err
			}
		case nodeconfigs.IPAddressThresholdActionSwitch:
			err := this.doSwitchAction(tx, role, clusterId, nodeId, addr, threshold, action, summaryList)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
