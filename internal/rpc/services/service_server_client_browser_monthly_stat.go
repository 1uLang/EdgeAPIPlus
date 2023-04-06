package services

import (
	"context"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
)

// ServerClientBrowserMonthlyStatService 操作系统统计
type ServerClientBrowserMonthlyStatService struct {
	BaseService
}

// FindTopServerClientBrowserMonthlyStats 查找前N个操作系统
func (this *ServerClientBrowserMonthlyStatService) FindTopServerClientBrowserMonthlyStats(ctx context.Context, req *pb.FindTopServerClientBrowserMonthlyStatsRequest) (*pb.FindTopServerClientBrowserMonthlyStatsResponse, error) {
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
	statList, err := stats.SharedServerClientBrowserMonthlyStatDAO.ListStats(tx, req.ServerId, req.Month, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbStats := []*pb.FindTopServerClientBrowserMonthlyStatsResponse_Stat{}
	for _, stat := range statList {
		pbStat := &pb.FindTopServerClientBrowserMonthlyStatsResponse_Stat{
			Count:   int64(stat.Count),
			Version: stat.Version,
		}
		browser, err := models.SharedClientBrowserDAO.FindEnabledClientBrowser(tx, int64(stat.BrowserId))
		if err != nil {
			return nil, err
		}
		if browser == nil {
			continue
		}
		pbStat.ClientBrowser = &pb.ClientBrowser{
			Id:   int64(browser.Id),
			Name: browser.Name,
		}

		pbStats = append(pbStats, pbStat)
	}
	return &pb.FindTopServerClientBrowserMonthlyStatsResponse{Stats: pbStats}, nil
}
