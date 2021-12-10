// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package reporters

import (
	"context"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/reporterconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"net/http"
	"strings"
	"time"
)

// ReportResultService 区域监控报告结果
type ReportResultService struct {
	services.BaseService
}

// CountAllReportResults 计算监控结果数量
func (this *ReportResultService) CountAllReportResults(ctx context.Context, req *pb.CountAllReportResultsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedReportResultDAO.CountAllResults(tx, req.ReportNodeId, req.Level, types.Int8(req.OkState))
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListReportResults 列出单页监控结果
func (this *ReportResultService) ListReportResults(ctx context.Context, req *pb.ListReportResultsRequest) (*pb.ListReportResultsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	results, err := models.SharedReportResultDAO.ListResults(tx, req.ReportNodeId, types.Int8(req.OkState), req.Level, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbResults = []*pb.ReportResult{}
	for _, result := range results {
		pbResults = append(pbResults, &pb.ReportResult{
			Id:           int64(result.Id),
			Type:         result.Type,
			TargetId:     int64(result.TargetId),
			TargetDesc:   result.TargetDesc,
			ReportNodeId: int64(result.ReportNodeId),
			IsOk:         result.IsOk == 1,
			CostMs:       float32(result.CostMs),
			Error:        result.Error,
			UpdatedAt:    int64(result.UpdatedAt),
			Level:        result.Level,
		})
	}

	return &pb.ListReportResultsResponse{
		ReportResults: pbResults,
	}, nil
}

// UpdateReportResults 上传报告结果
func (this *ReportResultService) UpdateReportResults(ctx context.Context, req *pb.UpdateReportResultsRequest) (*pb.RPCSuccess, error) {
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeReport)
	if err != nil {
		return nil, err
	}

	if !teaconst.IsPlus {
		return nil, errors.New("the commercial version is expired.")
	}

	var tx = this.NullTx()

	err = validateClient(tx, nodeId, ctx)
	if err != nil {
		return nil, err
	}

	// 设置
	var setting = reporterconfigs.DefaultGlobalSetting()
	settingJSON, err := models.SharedSysSettingDAO.ReadSetting(tx, systemconfigs.SettingCodeReportNodeGlobalSetting)
	if err != nil {
		return nil, err
	}
	if len(settingJSON) > 0 {
		err = json.Unmarshal(settingJSON, setting)
		if err != nil {
			return nil, err
		}
	}

	for _, result := range req.ReportResults {
		// 更新数据
		err := models.SharedReportResultDAO.UpdateResult(tx, result.Type, result.TargetId, result.TargetDesc, nodeId, result.Level, result.IsOk, float64(result.CostMs), result.Error)
		if err != nil {
			return nil, err
		}

		// 更新对象状态
		costMs, err := models.SharedReportResultDAO.FindAvgCostMsWithTarget(tx, result.Type, result.TargetId)
		if err != nil {
			return nil, err
		}

		level, err := models.SharedReportResultDAO.FindAvgLevelWithTarget(tx, result.Type, result.TargetId)
		if err != nil {
			return nil, err
		}

		percent, err := models.SharedReportResultDAO.FindConnectivityWithTargetPercent(tx, result.Type, result.TargetId, 0)
		if err != nil {
			return nil, err
		}

		// 是否应该通知
		if setting != nil && percent < setting.MinNotifyConnectivity {
			switch result.Type {
			case reporterconfigs.TaskTypeIPAddr:
				addr, err := models.SharedNodeIPAddressDAO.FindEnabledAddress(tx, result.TargetId)
				if err != nil {
					return nil, err
				}
				if addr != nil {
					var nodeId = int64(addr.NodeId)
					clusterId, err := models.SharedNodeDAO.FindNodeClusterId(tx, nodeId)
					if err != nil {
						return nil, err
					}

					var messageSubject = "IP地址：" + addr.Ip + "连通性低于" + types.String(setting.MinNotifyConnectivity) + "%"
					err = models.SharedMessageDAO.CreateNodeMessage(tx, addr.Role, clusterId, nodeId, models.MessageTypeConnectivity, models.LevelError, messageSubject, messageSubject, maps.Map{"addrId": addr.Id}.AsJSON(), false)
					if err != nil {
						return nil, err
					}

					err = models.SharedMessageTaskDAO.CreateMessageTasks(tx, addr.Role, clusterId, nodeId, 0, models.MessageTypeConnectivity, messageSubject, messageSubject)
					if err != nil {
						return nil, err
					}

					// 发送外部通知
					if len(setting.NotifyWebHookURL) > 0 {
						var client = utils.SharedHttpClient(10 * time.Second)
						var url = setting.NotifyWebHookURL
						var args = "role=" + addr.Role + "&clusterId=" + types.String(clusterId) + "&nodeId=" + types.String(nodeId) + "&addressId=" + types.String(addr.Id) + "&ip=" + addr.Ip
						var hasQuestionMark = strings.Contains(url, "?")
						if hasQuestionMark {
							url += "&" + args
						} else {
							url += "?" + args
						}
						req, err := http.NewRequest(http.MethodGet, url, nil)
						if err != nil {
							// 不阻断执行
							remotelogs.Error("ReportResultService.UpdateReportResults", "notify url '"+url+"' failed: "+err.Error())
						} else {
							resp, err := client.Do(req)
							if err != nil {
								// 不阻断执行
								remotelogs.Error("ReportResultService.UpdateReportResults", "notify url '"+url+"' failed: "+err.Error())
							} else {
								_ = resp.Body.Close()
							}
						}
					}
				}
			}
		}

		// 保存
		switch result.Type {
		case reporterconfigs.TaskTypeIPAddr:
			err = models.SharedNodeIPAddressDAO.UpdateAddressConnectivity(tx, result.TargetId, &nodeconfigs.Connectivity{
				CostMs:    costMs,
				Level:     level,
				Percent:   percent * 100,
				UpdatedAt: time.Now().Unix(),
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// FindAllReportResults 查询某个对象的监控结果
func (this *ReportResultService) FindAllReportResults(ctx context.Context, req *pb.FindAllReportResultsRequest) (*pb.FindAllReportResultsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	results, err := models.SharedReportResultDAO.FindAllResults(tx, req.Type, req.TargetId)
	if err != nil {
		return nil, err
	}
	var pbResults = []*pb.ReportResult{}
	for _, result := range results {
		pbResults = append(pbResults, &pb.ReportResult{
			Id:           int64(result.Id),
			Type:         result.Type,
			TargetId:     int64(result.TargetId),
			TargetDesc:   result.TargetDesc,
			ReportNodeId: int64(result.ReportNodeId),
			IsOk:         result.IsOk == 1,
			CostMs:       float32(result.CostMs),
			Error:        result.Error,
			UpdatedAt:    int64(result.UpdatedAt),
			Level:        result.Level,
		})
	}
	return &pb.FindAllReportResultsResponse{
		ReportResults: pbResults,
	}, nil
}
