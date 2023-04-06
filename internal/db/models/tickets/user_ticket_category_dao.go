package tickets

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	UserTicketCategoryStateEnabled  = 1 // 已启用
	UserTicketCategoryStateDisabled = 0 // 已禁用
)

type UserTicketCategoryDAO dbs.DAO

func NewUserTicketCategoryDAO() *UserTicketCategoryDAO {
	return dbs.NewDAO(&UserTicketCategoryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserTicketCategories",
			Model:  new(UserTicketCategory),
			PkName: "id",
		},
	}).(*UserTicketCategoryDAO)
}

var SharedUserTicketCategoryDAO *UserTicketCategoryDAO

func init() {
	dbs.OnReady(func() {
		SharedUserTicketCategoryDAO = NewUserTicketCategoryDAO()
	})
}

// EnableUserTicketCategory 启用条目
func (this *UserTicketCategoryDAO) EnableUserTicketCategory(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketCategoryStateEnabled).
		Update()
	return err
}

// DisableUserTicketCategory 禁用条目
func (this *UserTicketCategoryDAO) DisableUserTicketCategory(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketCategoryStateDisabled).
		Update()
	return err
}

// FindEnabledUserTicketCategory 查找启用中的条目
func (this *UserTicketCategoryDAO) FindEnabledUserTicketCategory(tx *dbs.Tx, id int64) (*UserTicketCategory, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserTicketCategoryStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserTicketCategory), err
}

// FindUserTicketCategoryName 根据主键查找名称
func (this *UserTicketCategoryDAO) FindUserTicketCategoryName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateCategory 创建分类
func (this *UserTicketCategoryDAO) CreateCategory(tx *dbs.Tx, name string) (int64, error) {
	var op = NewUserTicketCategoryOperator()
	op.Name = name
	op.IsOn = true
	op.State = UserTicketCategoryStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateCategory 修改分类
func (this *UserTicketCategoryDAO) UpdateCategory(tx *dbs.Tx, categoryId int64, name string, isOn bool) error {
	if categoryId <= 0 {
		return errors.New("invalid categoryId")
	}
	var op = NewUserTicketCategoryOperator()
	op.Id = categoryId
	op.Name = name
	op.IsOn = isOn
	return this.Save(tx, op)
}

// FindAllEnabledCategories 查找所有分类
func (this *UserTicketCategoryDAO) FindAllEnabledCategories(tx *dbs.Tx) (result []*UserTicketCategory, err error) {
	_, err = this.Query(tx).
		State(UserTicketCategoryStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAndOnCategories 查找所有分类
func (this *UserTicketCategoryDAO) FindAllEnabledAndOnCategories(tx *dbs.Tx) (result []*UserTicketCategory, err error) {
	_, err = this.Query(tx).
		State(UserTicketCategoryStateEnabled).
		Attr("isOn", true).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}
