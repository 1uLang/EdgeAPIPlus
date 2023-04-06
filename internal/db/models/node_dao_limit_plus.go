// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package models

import (
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

func (this *NodeDAO) CheckNodesLimit(tx *dbs.Tx) error {
	// 检查节点数量
	if teaconst.MaxNodes > 0 {
		count, err := this.Query(tx).
			State(NodeStateEnabled).
			Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)").
			Count()
		if err != nil {
			return err
		}
		if int64(teaconst.MaxNodes) <= count {
			return errors.New("[商业版]超出最大节点数限制：" + types.String(teaconst.MaxNodes) + "，当前已用：" + types.String(count) + "，请购买更多配额")
		}
	}

	return nil
}
