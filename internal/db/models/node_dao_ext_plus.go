// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus

package models

import (
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

func (this *NodeDAO) composeExtConfig(tx *dbs.Tx, config *nodeconfigs.NodeConfig, clusterIds []int64, cacheMap *utils.CacheMap) error {
	// 脚本
	scriptConfigs, err := SharedScriptHistoryDAO.ComposeScriptConfigs(tx, 0, cacheMap)
	if err != nil {
		return err
	}
	config.CommonScripts = scriptConfigs

	// 父节点
	if teaconst.IsPlus {
		if config.Level == 1 {
			parentNodes, err := SharedNodeDAO.FindParentNodeConfigs(tx, config.Id, config.GroupId, clusterIds, types.Int(config.Level))
			if err != nil {
				return err
			}
			config.ParentNodes = parentNodes
		}
	}

	return nil
}
