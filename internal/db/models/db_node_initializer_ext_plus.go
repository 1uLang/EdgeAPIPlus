// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus
// +build plus

package models

import (
	"errors"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"hash/crc32"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DNS服务访问
var nsAccessLogDAOMapping = map[int64]*NSAccessLogDAOWrapper{} // dbNodeId => DAO
var nsAccessLogTableMapping = map[string]bool{}                // tableName_crc(dsn) => true

// NSAccessLogDAOWrapper NS访问日志DAO
type NSAccessLogDAOWrapper struct {
	DAO    *NSAccessLogDAO
	NodeId int64
}

func randomNSAccessLogDAO() (dao *NSAccessLogDAOWrapper) {
	accessLogLocker.RLock()
	defer accessLogLocker.RUnlock()
	if len(nsAccessLogDAOMapping) == 0 {
		dao = nil
		return
	}

	var daoList = []*NSAccessLogDAOWrapper{}

	for _, d := range nsAccessLogDAOMapping {
		daoList = append(daoList, d)
	}

	var l = len(daoList)
	if l == 0 {
		return
	}

	if l == 1 {
		return daoList[0]
	}

	return daoList[rands.Int(0, l-1)]
}

func findNSAccessLogTableName(db *dbs.DB, day string) (tableName string, ok bool, err error) {
	if !regexp.MustCompile(`^\d{8}$`).MatchString(day) {
		err = errors.New("invalid day '" + day + "', should be YYYYMMDD")
		return
	}

	config, err := db.Config()
	if err != nil {
		return "", false, err
	}

	tableName = "edgeNSAccessLogs_" + day
	cacheKey := tableName + "_" + fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(config.Dsn)))

	accessLogLocker.RLock()
	_, ok = nsAccessLogTableMapping[cacheKey]
	accessLogLocker.RUnlock()
	if ok {
		return tableName, true, nil
	}

	tableNames, err := db.TableNames()
	if err != nil {
		return tableName, false, err
	}

	return tableName, utils.ContainsStringInsensitive(tableNames, tableName), nil
}

func findNSAccessLogTable(db *dbs.DB, day string, force bool) (string, error) {
	config, err := db.Config()
	if err != nil {
		return "", err
	}

	var tableName = "edgeNSAccessLogs_" + day
	var cacheKey = tableName + "_" + fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(config.Dsn)))

	if !force {
		accessLogLocker.RLock()
		_, ok := nsAccessLogTableMapping[cacheKey]
		accessLogLocker.RUnlock()
		if ok {
			return tableName, nil
		}
	}

	tableNames, err := db.TableNames()
	if err != nil {
		return tableName, err
	}

	if utils.ContainsStringInsensitive(tableNames, tableName) {
		accessLogLocker.Lock()
		nsAccessLogTableMapping[cacheKey] = true
		accessLogLocker.Unlock()
		return tableName, nil
	}

	// 创建表格
	_, err = db.Exec("CREATE TABLE `" + tableName + "` (\n  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n  `domainId` int(11) unsigned DEFAULT '0' COMMENT '域名ID',\n  `recordId` int(11) unsigned DEFAULT '0' COMMENT '记录ID',\n  `content` json DEFAULT NULL COMMENT '访问数据',\n  `requestId` varchar(128) DEFAULT NULL COMMENT '请求ID',\n  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n  `remoteAddr` varchar(128) DEFAULT NULL COMMENT 'IP',\n  PRIMARY KEY (`id`),\n  KEY `nodeId` (`nodeId`),\n  KEY `domainId` (`domainId`),\n  KEY `recordId` (`recordId`),\n  KEY `requestId` (`requestId`),\n  KEY `remoteAddr` (`remoteAddr`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='域名服务访问日志';")
	if err != nil {
		if CheckSQLErrCode(err, 1050) { // Error 1050: Table 'xxx' already exists
			accessLogLocker.Lock()
			nsAccessLogTableMapping[cacheKey] = true
			accessLogLocker.Unlock()

			return tableName, nil
		}

		return tableName, err
	}

	accessLogLocker.Lock()
	nsAccessLogTableMapping[cacheKey] = true
	accessLogLocker.Unlock()

	return tableName, nil
}

func initAccessLogDAO(db *dbs.DB, node *DBNode) {
	var nodeId = int64(node.Id)

	// nsAccessLog
	{
		tableName, err := findNSAccessLogTable(db, timeutil.Format("Ymd"), false)
		if err != nil {
			if !strings.Contains(err.Error(), "1050") { // 非表格已存在错误
				remotelogs.Error("DB_NODE", "create first table in database node failed: "+err.Error())

				// 创建节点日志
				createLogErr := SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleDatabase, nodeId, 0, 0, "error", "ACCESS_LOG", "can not create access log table: "+err.Error(), time.Now().Unix(), "", nil)
				if createLogErr != nil {
					remotelogs.Error("NODE_LOG", createLogErr.Error())
				}

				return
			} else {
				err = nil
			}
		}

		daoObject := dbs.DAOObject{
			Instance: db,
			DB:       node.Name + "(id:" + strconv.Itoa(int(node.Id)) + ")",
			Table:    tableName,
			PkName:   "id",
			Model:    new(NSAccessLog),
		}
		err = daoObject.Init()
		if err != nil {
			remotelogs.Error("DB_NODE", "initialize dao failed: "+err.Error())
			return
		}

		accessLogLocker.Lock()
		accessLogDBMapping[nodeId] = db
		var dao = &NSAccessLogDAO{
			DAOObject: daoObject,
		}
		nsAccessLogDAOMapping[nodeId] = &NSAccessLogDAOWrapper{
			DAO:    dao,
			NodeId: nodeId,
		}
		accessLogLocker.Unlock()
	}
}
