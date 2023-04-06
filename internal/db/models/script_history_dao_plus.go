//go:build plus
// +build plus

package models

import (
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"time"
)

type ScriptHistoryDAO dbs.DAO

func NewScriptHistoryDAO() *ScriptHistoryDAO {
	return dbs.NewDAO(&ScriptHistoryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeScriptHistories",
			Model:  new(ScriptHistory),
			PkName: "id",
		},
	}).(*ScriptHistoryDAO)
}

var SharedScriptHistoryDAO *ScriptHistoryDAO

func init() {
	dbs.OnReady(func() {
		SharedScriptHistoryDAO = NewScriptHistoryDAO()
	})
}

// PublishScripts 发布脚本
func (this *ScriptHistoryDAO) PublishScripts(tx *dbs.Tx, userId int64) error {
	scripts, err := SharedScriptDAO.FindAllEnabledAndOnScripts(tx, userId)
	if err != nil {
		return err
	}

	// 如果为空
	if len(scripts) == 0 {
		var op = NewScriptHistoryOperator()
		op.ScriptId = 0
		op.Code = ""
		op.Filename = ""
		op.UserId = userId
		op.Version = time.Now().Unix()
		err = this.Save(tx, op)
		if err != nil {
			return err
		}
		return this.NotifyUpdateAll(tx)
	}

	// 最大更新时间
	var version int64 = 0
	for _, script := range scripts {
		if int64(script.UpdatedAt) > version {
			version = int64(script.UpdatedAt)
		}
	}

	for _, script := range scripts {
		var op = NewScriptHistoryOperator()
		op.ScriptId = script.Id
		op.Code = script.Code
		op.Filename = script.Filename
		op.UserId = userId
		op.Version = version
		err := this.Save(tx, op)
		if err != nil {
			return err
		}
	}

	return this.NotifyUpdateAll(tx)
}

// FindAllScripts 查找所有脚本
func (this *ScriptHistoryDAO) FindAllScripts(tx *dbs.Tx, userId int64) (result []*ScriptHistory, err error) {
	// 最后一个版本
	version, err := this.Query(tx).
		Result("version").
		Attr("userId", userId).
		DescPk().
		FindInt64Col(0)
	if err != nil || version <= 0 {
		return nil, err
	}

	// 根据版本号对应
	_, err = this.Query(tx).
		Attr("userId", userId).
		Attr("version", version).
		Gt("scriptId", 0).
		Slice(&result).
		FindAll()

	return
}

// ComposeScriptConfigs 组合所有脚本配置
func (this *ScriptHistoryDAO) ComposeScriptConfigs(tx *dbs.Tx, userId int64, cacheMap *utils.CacheMap) ([]*serverconfigs.CommonScript, error) {
	var cacheKey = this.Table + ":ComposeScriptConfigs:" + types.String(userId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.([]*serverconfigs.CommonScript), nil
		}
	}

	var result = []*serverconfigs.CommonScript{}
	scripts, err := this.FindAllScripts(tx, userId)
	if err != nil {
		return nil, err
	}
	for _, script := range scripts {
		result = append(result, &serverconfigs.CommonScript{
			Id:       int64(script.Id),
			IsOn:     true,
			Filename: script.Filename,
			Code:     script.Code,
		})
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, result)
	}

	return result, nil
}

// CheckScriptsUpdates 检查脚本是否需要更新
func (this *ScriptHistoryDAO) CheckScriptsUpdates(tx *dbs.Tx, userId int64) (hasUpdates bool, version int64, err error) {
	lastUpdatedAt, err := SharedScriptDAO.FindScriptLastUpdatedTime(tx, userId)
	if err != nil {
		return false, 0, err
	}

	version, err = this.Query(tx).
		Result("version").
		Attr("userId", userId).
		Desc("version").
		FindInt64Col(0)
	if err != nil {
		return false, version, err
	}

	return version < lastUpdatedAt, version, nil
}

// NotifyUpdateAll 通知更新
func (this *ScriptHistoryDAO) NotifyUpdateAll(tx *dbs.Tx) error {
	clusterIds, err := SharedNodeClusterDAO.FindAllEnabledNodeClusterIds(tx)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, NodeTaskTypeScriptsChanged)
		if err != nil {
			return err
		}
	}
	return nil
}
