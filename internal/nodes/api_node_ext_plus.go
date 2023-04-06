// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package nodes

import (
	"github.com/TeaOSLab/EdgeAPI/internal/accesslogs"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
)

func (this *APINode) startAccessLogStorages() {
	goman.New(func() {
		accesslogs.SharedStorageManager.Start()
	})
}
