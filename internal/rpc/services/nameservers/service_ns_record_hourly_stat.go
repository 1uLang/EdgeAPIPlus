// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus

package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"time"
)

// NSRecordHourlyStatService NS记录小时统计
type NSRecordHourlyStatService struct {
	services.BaseService
}

// UploadNSRecordHourlyStats 上传统计
func (this *NSRecordHourlyStatService) UploadNSRecordHourlyStats(ctx context.Context, req *pb.UploadNSRecordHourlyStatsRequest) (*pb.RPCSuccess, error) {
	_, nodeId, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}
	if nodeId <= 0 {
		return nil, errors.New("invalid nodeId")
	}
	if len(req.Stats) == 0 {
		return this.Success()
	}

	var tx = this.NullTx()
	clusterId, err := models.SharedNSNodeDAO.FindNodeClusterId(tx, nodeId)
	if err != nil {
		return nil, err
	}

	// 增加小时统计
	for _, stat := range req.Stats {
		err := nameservers.SharedNSRecordHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, nodeId, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.NsDomainId, stat.NsRecordId, stat.CountRequests, stat.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// FindNSRecordHourlyStat 获取单个记录单个小时的统计
func (this *NSRecordHourlyStatService) FindNSRecordHourlyStat(ctx context.Context, req *pb.FindNSRecordHourlyStatRequest) (*pb.FindNSRecordHourlyStatResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(req.Hour) == 0 {
		req.Hour = timeutil.Format("YmdH")
	} else if !regexp.MustCompile(`^\d{10}$`).MatchString(req.Hour) {
		return nil, errors.New("invalid hour '" + req.Hour + "'")
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = nameservers.SharedNSRecordDAO.CheckUserRecord(tx, userId, req.NsRecordId)
		if err != nil {
			return nil, err
		}
	}

	stat, err := nameservers.SharedNSRecordHourlyStatDAO.FindHourlyStatWithRecordId(tx, req.NsRecordId, req.Hour)
	if err != nil {
		return nil, err
	}
	if stat == nil {
		return &pb.FindNSRecordHourlyStatResponse{NsRecordHourlyStat: nil}, nil
	}

	return &pb.FindNSRecordHourlyStatResponse{
		NsRecordHourlyStat: &pb.NSRecordHourlyStat{
			NsRecordId:    req.NsRecordId,
			Bytes:         int64(stat.Bytes),
			CountRequests: int64(stat.CountRequests),
			Hour:          req.Hour,
		},
	}, nil
}

// FindLatestNSRecordsHourlyStats 获取单个记录24小时内的统计
func (this *NSRecordHourlyStatService) FindLatestNSRecordsHourlyStats(ctx context.Context, req *pb.FindLatestNSRecordsHourlyStatsRequest) (*pb.FindLatestNSRecordsHourlyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSRecordDAO.CheckUserRecord(tx, userId, req.NsRecordId)
		if err != nil {
			return nil, err
		}
	}

	stats, err := nameservers.SharedNSRecordHourlyStatDAO.FindHourlyStatsWithRecordId(tx, req.NsRecordId, timeutil.Format("YmdH", time.Now().Add(-23*time.Hour)), timeutil.Format("YmdH"))
	if err != nil {
		return nil, err
	}
	var pbStats = []*pb.NSRecordHourlyStat{}
	for _, stat := range stats {
		pbStats = append(pbStats, &pb.NSRecordHourlyStat{
			NsRecordId:    req.NsRecordId,
			Bytes:         int64(stat.Bytes),
			CountRequests: int64(stat.CountRequests),
			Hour:          stat.Hour,
		})
	}
	return &pb.FindLatestNSRecordsHourlyStatsResponse{
		NsRecordHourlyStats: pbStats,
	}, nil
}

// FindNSRecordHourlyStatWithRecordIds 批量获取一组记录的统计
func (this *NSRecordHourlyStatService) FindNSRecordHourlyStatWithRecordIds(ctx context.Context, req *pb.FindNSRecordHourlyStatWithRecordIdsRequest) (*pb.FindNSRecordHourlyStatWithRecordIdsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(req.Hour) == 0 {
		req.Hour = timeutil.Format("YmdH")
	} else if !regexp.MustCompile(`^\d{10}$`).MatchString(req.Hour) {
		return nil, errors.New("invalid hour '" + req.Hour + "'")
	}

	var tx = this.NullTx()
	if userId > 0 {
		for _, recordId := range req.NsRecordIds {
			err = nameservers.SharedNSRecordDAO.CheckUserRecord(tx, userId, recordId)
			if err != nil {
				return nil, err
			}
		}
	}

	var pbStats = []*pb.NSRecordHourlyStat{}
	for _, recordId := range req.NsRecordIds {
		stat, err := nameservers.SharedNSRecordHourlyStatDAO.FindHourlyStatWithRecordId(tx, recordId, req.Hour)
		if err != nil {
			return nil, err
		}
		if stat == nil {
			continue
		}
		pbStats = append(pbStats, &pb.NSRecordHourlyStat{
			NsRecordId:    recordId,
			Bytes:         int64(stat.Bytes),
			CountRequests: int64(stat.CountRequests),
			Hour:          stat.Hour,
		})
	}

	return &pb.FindNSRecordHourlyStatWithRecordIdsResponse{
		NsRecordHourlyStats: pbStats,
	}, nil
}
