//go:build plus

package nameservers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSDomainGroupStateEnabled  = 1 // 已启用
	NSDomainGroupStateDisabled = 0 // 已禁用
)

type NSDomainGroupDAO dbs.DAO

func NewNSDomainGroupDAO() *NSDomainGroupDAO {
	return dbs.NewDAO(&NSDomainGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSDomainGroups",
			Model:  new(NSDomainGroup),
			PkName: "id",
		},
	}).(*NSDomainGroupDAO)
}

var SharedNSDomainGroupDAO *NSDomainGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedNSDomainGroupDAO = NewNSDomainGroupDAO()
	})
}

// EnableNSDomainGroup 启用条目
func (this *NSDomainGroupDAO) EnableNSDomainGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSDomainGroupStateEnabled).
		Update()
	return err
}

// DisableNSDomainGroup 禁用条目
func (this *NSDomainGroupDAO) DisableNSDomainGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSDomainGroupStateDisabled).
		Update()
	return err
}

// FindEnabledNSDomainGroup 查找启用中的条目
func (this *NSDomainGroupDAO) FindEnabledNSDomainGroup(tx *dbs.Tx, id int64) (*NSDomainGroup, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSDomainGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSDomainGroup), err
}

// FindNSDomainGroupName 根据主键查找名称
func (this *NSDomainGroupDAO) FindNSDomainGroupName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateGroup 创建分组
func (this *NSDomainGroupDAO) CreateGroup(tx *dbs.Tx, userId int64, name string) (int64, error) {
	var op = NewNSDomainGroupOperator()
	op.UserId = userId
	op.Name = name
	op.State = NSDomainGroupStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateGroup 修改分组
func (this *NSDomainGroupDAO) UpdateGroup(tx *dbs.Tx, groupId int64, name string, isOn bool) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	var op = NewNSDomainGroupOperator()
	op.Id = groupId
	op.Name = name
	op.IsOn = isOn
	return this.Save(tx, op)
}

// FindAllGroups 查找所有分组
func (this *NSDomainGroupDAO) FindAllGroups(tx *dbs.Tx, userId int64) (result []*NSDomainGroup, err error) {
	_, err = this.Query(tx).
		Attr("userId", userId).
		State(NSDomainGroupStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllAvailableGroups 计算可用分组数量
func (this *NSDomainGroupDAO) CountAllAvailableGroups(tx *dbs.Tx, userId int64) (int64, error) {
	return this.Query(tx).
		Attr("userId", userId).
		Attr("isOn", true).
		State(NSDomainGroupStateEnabled).
		Count()
}

// FindAllAvailableGroups 查找所有分组
func (this *NSDomainGroupDAO) FindAllAvailableGroups(tx *dbs.Tx, userId int64) (result []*NSDomainGroup, err error) {
	_, err = this.Query(tx).
		Attr("userId", userId).
		Attr("isOn", true).
		State(NSDomainGroupStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CheckUserGroup 检查用户分组
func (this *NSDomainGroupDAO) CheckUserGroup(tx *dbs.Tx, userId int64, groupId int64) error {
	if groupId <= 0 || userId <= 0 {
		return models.ErrNotFound
	}
	b, err := this.Query(tx).
		Pk(groupId).
		State(NSDomainGroupStateEnabled).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if !b {
		return models.ErrNotFound
	}
	return nil
}
