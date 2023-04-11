// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package models

import (
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
)

func (this *NodeDAO) composeExtConfig(tx *dbs.Tx, config *nodeconfigs.NodeConfig, clusterIds []int64, cacheMap *utils.CacheMap) error {
	return nil
}