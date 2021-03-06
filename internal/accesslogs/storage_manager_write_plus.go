// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

//go:build plus
// +build plus

package accesslogs

import (
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
)

// 写入日志
func (this *StorageManager) Write(policyId int64, accessLogs []*pb.HTTPAccessLog) error {
	if !teaconst.IsPlus {
		return nil
	}

	this.locker.Lock()
	storage, ok := this.storageMap[policyId]
	this.locker.Unlock()

	if !ok {
		return nil
	}

	if !storage.IsOk() {
		return nil
	}

	return storage.Write(accessLogs)
}
