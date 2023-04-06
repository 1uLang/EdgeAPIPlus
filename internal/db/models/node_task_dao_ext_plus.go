// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package models

import (
	"github.com/1uLang/EdgeCommon/pkg/configutils"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"time"
)

// ExtractNSClusterTask 分解NS节点集群任务
func (this *NodeTaskDAO) ExtractNSClusterTask(tx *dbs.Tx, clusterId int64, taskType NodeTaskType) error {
	nodeIds, err := SharedNSNodeDAO.FindAllNodeIdsMatch(tx, clusterId, true, configutils.BoolStateYes)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Attr("role", nodeconfigs.NodeRoleDNS).
		Attr("clusterId", clusterId).
		Param("clusterIdString", types.String(clusterId)).
		Where("nodeId > 0").
		Attr("type", taskType).
		Delete()
	if err != nil {
		return err
	}

	var version = time.Now().UnixNano()
	for _, nodeId := range nodeIds {
		err = this.CreateNodeTask(tx, nodeconfigs.NodeRoleDNS, clusterId, nodeId, 0, taskType, version)
		if err != nil {
			return err
		}
	}

	_, err = this.Query(tx).
		Attr("role", nodeconfigs.NodeRoleDNS).
		Attr("clusterId", clusterId).
		Attr("nodeId", 0).
		Attr("type", taskType).
		Delete()
	if err != nil {
		return err
	}

	return nil
}
