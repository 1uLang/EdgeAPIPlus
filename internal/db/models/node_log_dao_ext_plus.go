// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package models

import (
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

func (this *NodeLogDAO) deleteNodeLogsWithCluster(tx *dbs.Tx, role nodeconfigs.NodeRole, clusterId int64) error {
	var query = this.Query(tx).
		Attr("role", role)

	switch role {
	case nodeconfigs.NodeRoleDNS:
		query.Where("nodeId IN (SELECT id FROM " + SharedNSNodeDAO.Table + " WHERE clusterId=:clusterId)")
		query.Param("clusterId", clusterId)
	default:
		return nil
	}

	_, err := query.Delete()
	return err
}
