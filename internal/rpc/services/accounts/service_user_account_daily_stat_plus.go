// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package accounts

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/lists"
	"strings"
)

// UserAccountDailyStatService 用户账户统计服务
type UserAccountDailyStatService struct {
	services.BaseService
}

// ListUserAccountDailyStats 列出按天统计
func (this *UserAccountDailyStatService) ListUserAccountDailyStats(ctx context.Context, req *pb.ListUserAccountDailyStatsRequest) (*pb.ListUserAccountDailyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var dayFrom = req.DayFrom
	var dayTo = req.DayTo

	dayFrom = strings.ReplaceAll(dayFrom, "-", "")
	dayTo = strings.ReplaceAll(dayTo, "-", "")

	days, err := utils.RangeDays(dayFrom, dayTo)
	if err != nil {
		return nil, err
	}

	if len(days) == 0 {
		return &pb.ListUserAccountDailyStatsResponse{Stats: nil}, nil
	}

	var tx = this.NullTx()
	stats, err := accounts.SharedUserAccountDailyStatDAO.FindDailyStats(tx, dayFrom, dayTo)
	if err != nil {
		return nil, err
	}
	var statMap = map[string]*accounts.UserAccountDailyStat{} // day => Stat
	for _, stat := range stats {
		statMap[stat.Day] = stat
	}

	var pbStats = []*pb.ListUserAccountDailyStatsResponse_Stat{}
	for _, day := range days {
		stat, ok := statMap[day]
		if ok {
			pbStats = append(pbStats, &pb.ListUserAccountDailyStatsResponse_Stat{
				Day:     day,
				Income:  float32(stat.Income),
				Expense: float32(stat.Expense),
			})
		} else {
			pbStats = append(pbStats, &pb.ListUserAccountDailyStatsResponse_Stat{
				Day:     day,
				Income:  0,
				Expense: 0,
			})
		}
	}

	// 反向排序
	lists.Reverse(pbStats)

	return &pb.ListUserAccountDailyStatsResponse{Stats: pbStats}, nil
}

// ListUserAccountMonthlyStats 列出按月统计
func (this *UserAccountDailyStatService) ListUserAccountMonthlyStats(ctx context.Context, req *pb.ListUserAccountMonthlyStatsRequest) (*pb.ListUserAccountMonthlyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var dayFrom = req.DayFrom
	var dayTo = req.DayTo

	dayFrom = strings.ReplaceAll(dayFrom, "-", "")
	dayTo = strings.ReplaceAll(dayTo, "-", "")

	months, err := utils.RangeMonths(dayFrom, dayTo)
	if err != nil {
		return nil, err
	}

	if len(months) == 0 {
		return &pb.ListUserAccountMonthlyStatsResponse{Stats: nil}, nil
	}

	var tx = this.NullTx()
	stats, err := accounts.SharedUserAccountDailyStatDAO.FindMonthlyStats(tx, dayFrom, dayTo)
	if err != nil {
		return nil, err
	}
	var statMap = map[string]*accounts.UserAccountDailyStat{} // month => Stat
	for _, stat := range stats {
		statMap[stat.Month] = stat
	}

	var pbStats = []*pb.ListUserAccountMonthlyStatsResponse_Stat{}

	for _, month := range months {
		stat, ok := statMap[month]
		if ok {
			pbStats = append(pbStats, &pb.ListUserAccountMonthlyStatsResponse_Stat{
				Month:   month,
				Income:  float32(stat.Income),
				Expense: float32(stat.Expense),
			})
		} else {
			pbStats = append(pbStats, &pb.ListUserAccountMonthlyStatsResponse_Stat{
				Month:   month,
				Income:  0,
				Expense: 0,
			})
		}
	}

	// 反向排序
	lists.Reverse(pbStats)

	return &pb.ListUserAccountMonthlyStatsResponse{Stats: pbStats}, nil
}
