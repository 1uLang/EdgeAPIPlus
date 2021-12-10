// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package authority

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

// 更新Plus数值
func init() {
	dbs.OnReadyDone(func() {
		isPlus, _ := SharedAuthorityKeyDAO.IsPlus(nil)
		teaconst.IsPlus = isPlus

		var ticker = time.NewTicker(5 * time.Minute)
		go func() {
			for range ticker.C {
				isPlus, err := SharedAuthorityKeyDAO.IsPlus(nil)
				if err != nil {
					remotelogs.Error("AuthorityKeyDAO", "check isPlus failed: "+err.Error())
				}
				teaconst.IsPlus = isPlus
			}
		}()
	})
}

// IsPlus 判断是否为企业版
func (this *AuthorityKeyDAO) IsPlus(tx *dbs.Tx) (bool, error) {
	return true, nil
}
