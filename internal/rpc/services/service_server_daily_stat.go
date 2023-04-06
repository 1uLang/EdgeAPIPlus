package services

import (
	"context"
	"errors"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"math"
	"regexp"
	"time"
)

// ServerDailyStatService 服务统计相关服务
type ServerDailyStatService struct {
	BaseService
}

// UploadServerDailyStats 上传统计
func (this *ServerDailyStatService) UploadServerDailyStats(ctx context.Context, req *pb.UploadServerDailyStatsRequest) (*pb.RPCSuccess, error) {
	role, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 保存统计数据
	err = models.SharedServerDailyStatDAO.SaveStats(tx, req.Stats)
	if err != nil {
		return nil, err
	}

	var clusterId int64
	switch role {
	case rpcutils.UserTypeDNS:
		clusterId, err = models.SharedNSNodeDAO.FindNodeClusterId(tx, nodeId)
		if err != nil {
			return nil, err
		}
	}

	// 写入其他统计表
	// TODO 将来改成每小时入库一次
	for _, stat := range req.Stats {
		if role == rpcutils.UserTypeNode {
			clusterId, err = models.SharedServerDAO.FindServerClusterId(tx, stat.ServerId)
			if err != nil {
				return nil, err
			}
		}

		// 总体流量（按天）
		err = stats.SharedTrafficDailyStatDAO.IncreaseDailyStat(tx, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
		if err != nil {
			return nil, err
		}

		// 总体统计（按小时）
		err = stats.SharedTrafficHourlyStatDAO.IncreaseHourlyStat(tx, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
		if err != nil {
			return nil, err
		}

		// 节点流量
		if nodeId > 0 {
			err = stats.SharedNodeTrafficDailyStatDAO.IncreaseDailyStat(tx, clusterId, role, nodeId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
			if err != nil {
				return nil, err
			}

			err = stats.SharedNodeTrafficHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, role, nodeId, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
			if err != nil {
				return nil, err
			}

			// 集群流量
			if clusterId > 0 {
				err = stats.SharedNodeClusterTrafficDailyStatDAO.IncreaseDailyStat(tx, clusterId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 域名统计
	for _, stat := range req.DomainStats {
		if role == rpcutils.UserTypeNode {
			clusterId, err = models.SharedServerDAO.FindServerClusterId(tx, stat.ServerId)
			if err != nil {
				return nil, err
			}
		}

		err := stats.SharedServerDomainHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, nodeId, stat.ServerId, stat.Domain, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// FindLatestServerHourlyStats 按小时读取统计数据
func (this *ServerDailyStatService) FindLatestServerHourlyStats(ctx context.Context, req *pb.FindLatestServerHourlyStatsRequest) (*pb.FindLatestServerHourlyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	result := []*pb.FindLatestServerHourlyStatsResponse_HourlyStat{}
	if req.Hours > 0 {
		for i := int32(0); i < req.Hours; i++ {
			hourString := timeutil.Format("YmdH", time.Now().Add(-time.Duration(i)*time.Hour))
			stat, err := models.SharedServerDailyStatDAO.SumHourlyStat(tx, req.ServerId, hourString)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				result = append(result, &pb.FindLatestServerHourlyStatsResponse_HourlyStat{
					Hour:                hourString,
					Bytes:               stat.Bytes,
					CachedBytes:         stat.CachedBytes,
					CountRequests:       stat.CountRequests,
					CountCachedRequests: stat.CountCachedRequests,
				})
			}
		}
	}
	return &pb.FindLatestServerHourlyStatsResponse{Stats: result}, nil
}

// FindLatestServerMinutelyStats 按分钟读取统计数据
func (this *ServerDailyStatService) FindLatestServerMinutelyStats(ctx context.Context, req *pb.FindLatestServerMinutelyStatsRequest) (*pb.FindLatestServerMinutelyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	result := []*pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{}
	cache := map[string]*pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{} // minute => stat

	var avgRatio int64 = 5 * 60 // 因为每5分钟记录一次

	if req.Minutes > 0 {
		for i := int32(0); i < req.Minutes; i++ {
			date := time.Now().Add(-time.Duration(i) * time.Minute)
			minuteString := timeutil.Format("YmdHi", date)

			minute := date.Minute()
			roundMinute := minute - minute%5
			if roundMinute != minute {
				date = date.Add(-time.Duration(minute-roundMinute) * time.Minute)
			}
			queryMinuteString := timeutil.Format("YmdHi", date)
			pbStat, ok := cache[queryMinuteString]
			if ok {
				result = append(result, &pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{
					Minute:              minuteString,
					Bytes:               pbStat.Bytes,
					CachedBytes:         pbStat.CachedBytes,
					CountRequests:       pbStat.CountRequests,
					CountCachedRequests: pbStat.CountCachedRequests,
				})
				continue
			}

			stat, err := models.SharedServerDailyStatDAO.SumMinutelyStat(tx, req.ServerId, queryMinuteString)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				pbStat = &pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{
					Minute:              minuteString,
					Bytes:               stat.Bytes / avgRatio,
					CachedBytes:         stat.CachedBytes / avgRatio,
					CountRequests:       int64(math.Ceil(float64(stat.CountRequests) / float64(avgRatio))),
					CountCachedRequests: int64(math.Ceil(float64(stat.CountCachedRequests) / float64(avgRatio))),
				}
				result = append(result, pbStat)
				cache[queryMinuteString] = pbStat
			}
		}
	}
	return &pb.FindLatestServerMinutelyStatsResponse{Stats: result}, nil
}

// FindServer5MinutelyStatsWithDay 读取某天的5分钟间隔流量
func (this *ServerDailyStatService) FindServer5MinutelyStatsWithDay(ctx context.Context, req *pb.FindServer5MinutelyStatsWithDayRequest) (*pb.FindServer5MinutelyStatsWithDayResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if len(req.Day) == 0 {
		req.Day = timeutil.Format("Ymd")
	}

	dailyStats, err := models.SharedServerDailyStatDAO.FindStatsWithDay(tx, req.ServerId, req.Day, req.TimeFrom, req.TimeTo)
	if err != nil {
		return nil, err
	}

	var pbStats = []*pb.FindServer5MinutelyStatsWithDayResponse_Stat{}
	for _, stat := range dailyStats {
		pbStats = append(pbStats, &pb.FindServer5MinutelyStatsWithDayResponse_Stat{
			Day:                 stat.Day,
			TimeFrom:            stat.TimeFrom,
			TimeTo:              stat.TimeTo,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
		})
	}
	return &pb.FindServer5MinutelyStatsWithDayResponse{Stats: pbStats}, nil
}

// FindLatestServerDailyStats 按天读取统计数据
func (this *ServerDailyStatService) FindLatestServerDailyStats(ctx context.Context, req *pb.FindLatestServerDailyStatsRequest) (*pb.FindLatestServerDailyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	result := []*pb.FindLatestServerDailyStatsResponse_DailyStat{}
	if req.Days > 0 {
		for i := int32(0); i < req.Days; i++ {
			dayString := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -int(i)))
			stat, err := models.SharedServerDailyStatDAO.SumDailyStat(tx, req.ServerId, dayString)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				result = append(result, &pb.FindLatestServerDailyStatsResponse_DailyStat{
					Day:                 dayString,
					Bytes:               stat.Bytes,
					CachedBytes:         stat.CachedBytes,
					CountRequests:       stat.CountRequests,
					CountCachedRequests: stat.CountCachedRequests,
				})
			}
		}
	}
	return &pb.FindLatestServerDailyStatsResponse{Stats: result}, nil
}

// SumCurrentServerDailyStats 查找单个服务当前统计数据
func (this *ServerDailyStatService) SumCurrentServerDailyStats(ctx context.Context, req *pb.SumCurrentServerDailyStatsRequest) (*pb.SumCurrentServerDailyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx = this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 按日
	stat, err := models.SharedServerDailyStatDAO.SumCurrentDailyStat(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	var pbStat = &pb.ServerDailyStat{
		ServerId: req.ServerId,
	}
	if stat != nil {
		pbStat = &pb.ServerDailyStat{
			ServerId:            req.ServerId,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		}
	}

	return &pb.SumCurrentServerDailyStatsResponse{ServerDailyStat: pbStat}, nil
}

// SumServerDailyStats 计算单个服务的日统计
func (this *ServerDailyStatService) SumServerDailyStats(ctx context.Context, req *pb.SumServerDailyStatsRequest) (*pb.SumServerDailyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 某日统计
	var day = timeutil.Format("Ymd")
	if regexp.MustCompile(`^\d{8}$`).MatchString(req.Day) {
		day = req.Day
	}

	stat, err := models.SharedServerDailyStatDAO.SumDailyStat(tx, req.ServerId, day)
	if err != nil {
		return nil, err
	}

	var pbStat = &pb.ServerDailyStat{
		ServerId: req.ServerId,
	}
	if stat != nil {
		pbStat = &pb.ServerDailyStat{
			ServerId:            req.ServerId,
			Bytes:               stat.Bytes,
			CachedBytes:         stat.CachedBytes,
			CountRequests:       stat.CountRequests,
			CountCachedRequests: stat.CountCachedRequests,
			CountAttackRequests: stat.CountAttackRequests,
			AttackBytes:         stat.AttackBytes,
		}
	}
	return &pb.SumServerDailyStatsResponse{ServerDailyStat: pbStat}, nil
}

// SumServerMonthlyStats 计算单个服务的月统计
func (this *ServerDailyStatService) SumServerMonthlyStats(ctx context.Context, req *pb.SumServerMonthlyStatsRequest) (*pb.SumServerMonthlyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 某月统计
	var month = timeutil.Format("Ym")
	if regexp.MustCompile(`^\d{6}$`).MatchString(req.Month) {
		month = req.Month
	}

	// 按月
	stat, err := models.SharedServerDailyStatDAO.SumMonthlyStat(tx, req.ServerId, month)
	if err != nil {
		return nil, err
	}

	var pbStat = &pb.ServerDailyStat{
		ServerId: req.ServerId,
	}
	if stat != nil {
		pbStat = &pb.ServerDailyStat{
			ServerId:            req.ServerId,
			Bytes:               stat.Bytes,
			CachedBytes:         stat.CachedBytes,
			CountRequests:       stat.CountRequests,
			CountCachedRequests: stat.CountCachedRequests,
			CountAttackRequests: stat.CountAttackRequests,
			AttackBytes:         stat.AttackBytes,
		}
	}

	return &pb.SumServerMonthlyStatsResponse{ServerMonthlyStat: pbStat}, nil
}

// ------- api 客户定制化接口

type ClusterActualTimeTrafficStatRequest struct {
	ClusterId int64 `json:"clusterId"`
}
type ActualTimeTrafficStatResponse struct {
	Bytes int64 `json:"bytes"`
}
type ListTrafficStatServerDailyStatsResponse struct {
	Stats               []*ListTrafficStatServerDailyStatsResponse_TrafficStat
	WeekTotalBytes      int64 `protobuf:"varint,2,opt,name=weekTotalBytes,proto3" json:"weekTotalBytes,omitempty"`           //本周总流量统计
	YesterdayTotalBytes int64 `protobuf:"varint,3,opt,name=yesterdayTotalBytes,proto3" json:"yesterdayTotalBytes,omitempty"` //昨日总流量统计
	TodayTotalBytes     int64 `protobuf:"varint,4,opt,name=todayTotalBytes,proto3" json:"todayTotalBytes,omitempty"`         //当前总流量统计
}
type ListTrafficStatServerDailyStatsResponse_TrafficStat struct {
	ServerId       int64 `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"`
	WeekBytes      int64 `protobuf:"varint,2,opt,name=weekBytes,proto3" json:"weekBytes,omitempty"`
	YesterdayBytes int64 `protobuf:"varint,3,opt,name=yesterdayBytes,proto3" json:"yesterdayBytes,omitempty"`
	TodayBytes     int64 `protobuf:"varint,4,opt,name=todayBytes,proto3" json:"todayBytes,omitempty"`
}

type ActualTimeBandwitchStatResponse struct {
	Up   string `json:"up,omitempty"` // 上行带宽
	Down string `json:"down"`         // 下行带宽
}

// ListTrafficStat 各服务流量统计
func (this *ServerDailyStatService) ListTrafficStat(ctx context.Context, req *pb.RPCSuccess) (*ListTrafficStatServerDailyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := &ListTrafficStatServerDailyStatsResponse{}

	// 按日流量统计
	this.BeginTag(ctx, "SharedTrafficDailyStatDAO.FindDailyStats")
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -7))
	dailyTrafficStats, err := stats.SharedTrafficDailyStatDAO.FindDailyStats(tx, dayFrom, timeutil.Format("Ymd"))
	this.EndTag(ctx, "SharedTrafficDailyStatDAO.FindDailyStats")
	if err != nil {
		return nil, err
	}

	// 今日流量
	if len(dailyTrafficStats) > 0 {
		result.TodayTotalBytes = int64(dailyTrafficStats[len(dailyTrafficStats)-1].Bytes)
	}
	// 昨日总流量
	if len(dailyTrafficStats) > 1 {
		result.YesterdayTotalBytes = int64(dailyTrafficStats[len(dailyTrafficStats)-2].Bytes)
	}
	// 本周总流量
	if len(dailyTrafficStats) > 0 {
		for _, stat := range dailyTrafficStats {
			result.WeekTotalBytes += int64(stat.Bytes)
		}
	}

	//各服务流量统计
	serverIds, err := models.SharedServerDAO.FindAllEnabledServerIds(tx)
	if err != nil {
		return nil, err
	}
	result.Stats = make([]*ListTrafficStatServerDailyStatsResponse_TrafficStat, len(serverIds))
	for k, v := range serverIds {
		result.Stats[k] = new(ListTrafficStatServerDailyStatsResponse_TrafficStat)
		stat, err := this.FindLatestServerDailyStats(ctx, &pb.FindLatestServerDailyStatsRequest{ServerId: v, Days: 7})
		if err != nil {
			return nil, err
		}
		result.Stats[k].ServerId = v
		// 当日流量
		if len(stat.Stats) > 0 {
			result.Stats[k].TodayBytes = stat.Stats[len(stat.Stats)-1].Bytes
		}
		// 昨日流量
		if len(stat.Stats) > 1 {
			result.Stats[k].YesterdayBytes = stat.Stats[len(stat.Stats)-2].Bytes
		}
		// 本周流量
		if len(stat.Stats) > 0 {
			for _, bytes := range stat.Stats {
				result.Stats[k].WeekBytes += bytes.Bytes
			}
		}
	}
	return result, nil
}

// ServerActualTimeTrafficStat 服务实时流量统计
func (this *ServerDailyStatService) ServerActualTimeTrafficStat(ctx context.Context, req *pb.FindLatestServerDailyStatsRequest) (*ActualTimeTrafficStatResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if req.ServerId == 0 {
		return nil, errors.New("serverId 服务ID不能为空")
	}
	tx := this.NullTx()
	result := &ActualTimeTrafficStatResponse{}

	values, err := models.SharedServerDailyStatDAO.SumLastMinutelyStat(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	result.Bytes = values.Bytes
	return result, nil
}

// ClusterActualTimeTrafficStat 服务实时流量统计
func (this *ServerDailyStatService) ClusterActualTimeTrafficStat(ctx context.Context, req *ClusterActualTimeTrafficStatRequest) (*ActualTimeTrafficStatResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if req.ClusterId == 0 {
		return nil, errors.New("clusterId 集群ID不能为空")
	}
	tx := this.NullTx()

	result := &ActualTimeTrafficStatResponse{}

	values, err := models.SharedNodeValueDAO.FindLatestClusterValue(tx, "node", req.ClusterId, nodeconfigs.NodeValueItemTrafficOut, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		if value != nil {
			result.Bytes += value.DecodeMapValue().GetInt64("total")
		}
	}
	return result, nil
}

// ActualTimeTrafficStat 平台实时流量统计
func (this *ServerDailyStatService) ActualTimeTrafficStat(ctx context.Context, req *pb.RPCSuccess) (*ActualTimeTrafficStatResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := &ActualTimeTrafficStatResponse{}

	values, err := models.SharedNodeValueDAO.FindLatestAllValue(tx, "node", nodeconfigs.NodeValueItemTrafficOut, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		if value != nil {
			result.Bytes += value.DecodeMapValue().GetInt64("total")
		}
	}
	return result, nil
}

// ServerActualTimeBandwitchStat 服务实时流量统计
func (this *ServerDailyStatService) ServerActualTimeBandwitchStat(ctx context.Context, req *pb.FindLatestServerDailyStatsRequest) (*ActualTimeBandwitchStatResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if req.ServerId == 0 {
		return nil, errors.New("serverId 服务ID不能为空")
	}
	tx := this.NullTx()
	result := &ActualTimeBandwitchStatResponse{}

	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	result.Down = showBandwitch(float64(server.BandwidthBytes))
	return result, nil
}

// ClusterActualTimeBandwitchStat 服务实时流量统计
func (this *ServerDailyStatService) ClusterActualTimeBandwitchStat(ctx context.Context, req *ClusterActualTimeTrafficStatRequest) (*ActualTimeBandwitchStatResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if req.ClusterId == 0 {
		return nil, errors.New("clusterId 集群ID不能为空")
	}
	tx := this.NullTx()

	result := &ActualTimeBandwitchStatResponse{}
	downBytes := float64(0)
	values, err := models.SharedNodeValueDAO.FindLatestClusterValue(tx, "node", req.ClusterId, nodeconfigs.NodeValueItemTrafficOut, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		if value != nil {
			downBytes += value.DecodeMapValue().GetFloat64("total")
		}
	}
	upBytes := float64(0)
	values, err = models.SharedNodeValueDAO.FindLatestClusterValue(tx, "node", req.ClusterId, nodeconfigs.NodeValueItemTrafficOut, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		if value != nil {
			upBytes += value.DecodeMapValue().GetFloat64("total")
		}
	}
	result.Up = showBandwitch(upBytes)
	result.Down = showBandwitch(downBytes)
	return result, nil
}

// ActualTimeBandwitchStat 平台实时带宽统计
func (this *ServerDailyStatService) ActualTimeBandwitchStat(ctx context.Context, req *pb.RPCSuccess) (*ActualTimeBandwitchStatResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := &ActualTimeBandwitchStatResponse{}
	downBytes := float64(0)
	values, err := models.SharedNodeValueDAO.FindLatestAllValue(tx, "node", nodeconfigs.NodeValueItemTrafficOut, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		if value != nil {
			downBytes += value.DecodeMapValue().GetFloat64("total")
		}
	}
	upBytes := float64(0)
	values, err = models.SharedNodeValueDAO.FindLatestAllValue(tx, "node", nodeconfigs.NodeValueItemTrafficIn, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		if value != nil {
			upBytes += value.DecodeMapValue().GetFloat64("total")
		}
	}
	result.Up = showBandwitch(upBytes)
	result.Down = showBandwitch(downBytes)
	return result, nil
}

func showBandwitch(bytesize float64) string {
	bandwitch := ""
	if bytesize > 1024*1024 {
		bandwitch = types.String(math.Round(bytesize*100/1024/1024)/100) + "MB/s" // 100 = 两位小数
	} else if bytesize > 1024 {
		bandwitch = types.String(math.Round(bytesize*100/1024)/100) + "KB/s" // 100 = 两位小数
	} else {
		bandwitch = types.String(bytesize) + "B/s"
	}
	return bandwitch
}
