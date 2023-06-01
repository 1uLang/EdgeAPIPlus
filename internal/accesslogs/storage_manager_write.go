// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

//go:build plus

package accesslogs

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"runtime"
	"time"
)

var lastHTTPAccessLogPolicyId int64
var lastHTTPAccessLogPolicyUpdatedAt int64
var httpAccessLogQueue = make(chan []*pb.HTTPAccessLog, 1024)

func init() {
	if !teaconst.IsMain {
		return
	}

	// 多进程读取
	var threads = runtime.NumCPU() / 2
	if threads <= 0 {
		threads = 1
	}
	for i := 0; i < threads; i++ {
		goman.New(func() {
			for pbAccessLogs := range httpAccessLogQueue {
				if lastHTTPAccessLogPolicyId <= 0 {
					continue
				}
				_, _, err := SharedStorageManager.WriteToPolicy(lastHTTPAccessLogPolicyId, pbAccessLogs)
				if err != nil {
					remotelogs.Error("HTTP_ACCESS_LOG_POLICY", "write failed: "+err.Error())
				}
			}
		})
	}
}

func (this *StorageManager) Write(pbAccessLogs []*pb.HTTPAccessLog) error {
	if len(pbAccessLogs) == 0 {
		return nil
	}

	var currentTime = time.Now().Unix()
	var cacheTime int64 = 30 // 策略缓存时间

	var tx *dbs.Tx
	var policyId int64
	if lastHTTPAccessLogPolicyUpdatedAt < currentTime-cacheTime {
		var err error
		policyId, err = models.SharedHTTPAccessLogPolicyDAO.FindCurrentPublicPolicyId(tx)
		if err != nil {
			return err
		}
		lastHTTPAccessLogPolicyId = policyId
		lastHTTPAccessLogPolicyUpdatedAt = currentTime
	} else {
		policyId = lastHTTPAccessLogPolicyId
	}

	if policyId <= 0 {
		return nil
	}

	select {
	case httpAccessLogQueue <- pbAccessLogs:
	default:

	}

	return nil
}

// WriteToPolicy 写入日志到策略
func (this *StorageManager) WriteToPolicy(policyId int64, accessLogs []*pb.HTTPAccessLog) (success bool, failMessage string, err error) {
	if !teaconst.IsPlus {
		return false, "only works in plus version", nil
	}

	this.locker.Lock()
	storage, ok := this.storageMap[policyId]
	this.locker.Unlock()

	if !ok {
		return false, "the policy has not been started yet", nil
	}

	if !storage.IsOk() {
		return false, "the policy failed to start", nil
	}

	err = storage.Write(accessLogs)
	if err != nil {
		return false, "", errors.New("write access log to policy '" + types.String(policyId) + "' failed: " + err.Error())
	}
	return true, "", nil
}
