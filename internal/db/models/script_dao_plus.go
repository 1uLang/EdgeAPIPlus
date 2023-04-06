//go:build plus
// +build plus

package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"time"
)

const (
	ScriptStateEnabled  = 1 // 已启用
	ScriptStateDisabled = 0 // 已禁用
)

type ScriptDAO dbs.DAO

func NewScriptDAO() *ScriptDAO {
	return dbs.NewDAO(&ScriptDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeScripts",
			Model:  new(Script),
			PkName: "id",
		},
	}).(*ScriptDAO)
}

var SharedScriptDAO *ScriptDAO

func init() {
	dbs.OnReady(func() {
		SharedScriptDAO = NewScriptDAO()
	})
}

// EnableScript 启用条目
func (this *ScriptDAO) EnableScript(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ScriptStateEnabled).
		Update()
	return err
}

// DisableScript 禁用条目
func (this *ScriptDAO) DisableScript(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ScriptStateDisabled).
		Update()
	return err
}

// FindEnabledScript 查找启用中的条目
func (this *ScriptDAO) FindEnabledScript(tx *dbs.Tx, id int64) (*Script, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ScriptStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Script), err
}

// FindScriptName 根据主键查找名称
func (this *ScriptDAO) FindScriptName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateScript 创建脚本
func (this *ScriptDAO) CreateScript(tx *dbs.Tx, userId int64, name string, filename string, code string) (int64, error) {
	var op = NewScriptOperator()
	op.UserId = userId
	op.Name = name
	op.Filename = filename
	op.Code = code
	op.IsOn = true
	op.UpdatedAt = time.Now().Unix()
	op.State = ScriptStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateScript 修改脚本
func (this *ScriptDAO) UpdateScript(tx *dbs.Tx, scriptId int64, name string, filename string, code string, isOn bool) error {
	if scriptId <= 0 {
		return errors.New("invalid scriptId")
	}

	var op = NewScriptOperator()
	op.Id = scriptId
	op.Name = name
	op.Filename = filename
	op.Code = code
	op.IsOn = isOn
	op.UpdatedAt = time.Now().Unix()
	return this.Save(tx, op)
}

// CountAllEnabledScripts 计算脚本数量
func (this *ScriptDAO) CountAllEnabledScripts(tx *dbs.Tx, userId int64) (int64, error) {
	return this.Query(tx).
		State(ScriptStateEnabled).
		Attr("userId", userId).
		Count()
}

// ListEnabledScripts 列出一页脚本
func (this *ScriptDAO) ListEnabledScripts(tx *dbs.Tx, userId int64, offset int64, size int64) (result []*Script, err error) {
	_, err = this.Query(tx).
		State(ScriptStateEnabled).
		Attr("userId", userId).
		Asc("filename").
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAndOnScripts 查找所有启用的脚本
func (this *ScriptDAO) FindAllEnabledAndOnScripts(tx *dbs.Tx, userId int64) (result []*Script, err error) {
	_, err = this.Query(tx).
		State(ScriptStateEnabled).
		Attr("userId", userId).
		Attr("isOn", true).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CheckUserScript 检查用户脚本
func (this *ScriptDAO) CheckUserScript(tx *dbs.Tx, userId int64, scriptId int64) error {
	exists, err := this.Query(tx).
		State(ScriptStateEnabled).
		Pk(scriptId).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("can not find user script with id '" + types.String(scriptId) + "'")
	}
	return nil
}

// FindScriptLastUpdatedTime 读取最新更改时间
func (this *ScriptDAO) FindScriptLastUpdatedTime(tx *dbs.Tx, userId int64) (int64, error) {
	return this.Query(tx).
		Result("updatedAt").
		State(ScriptStateEnabled).
		Attr("isOn", true).
		Attr("userId", userId).
		Desc("updatedAt").
		FindInt64Col(0)
}
