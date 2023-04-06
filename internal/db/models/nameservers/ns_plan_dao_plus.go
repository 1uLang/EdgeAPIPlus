//go:build plus
// +build plus

package nameservers

import (
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSPlanStateEnabled  = 1 // 已启用
	NSPlanStateDisabled = 0 // 已禁用
)

type NSPlanDAO dbs.DAO

func NewNSPlanDAO() *NSPlanDAO {
	return dbs.NewDAO(&NSPlanDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSPlans",
			Model:  new(NSPlan),
			PkName: "id",
		},
	}).(*NSPlanDAO)
}

var SharedNSPlanDAO *NSPlanDAO

func init() {
	dbs.OnReady(func() {
		SharedNSPlanDAO = NewNSPlanDAO()
	})
}

// EnableNSPlan 启用条目
func (this *NSPlanDAO) EnableNSPlan(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSPlanStateEnabled).
		Update()
	return err
}

// DisableNSPlan 禁用条目
func (this *NSPlanDAO) DisableNSPlan(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSPlanStateDisabled).
		Update()
	return err
}

// FindEnabledNSPlan 查找启用中的条目
func (this *NSPlanDAO) FindEnabledNSPlan(tx *dbs.Tx, id int64) (*NSPlan, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(NSPlanStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSPlan), err
}

// FindNSPlanName 根据主键查找名称
func (this *NSPlanDAO) FindNSPlanName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreatePlan 创建套餐
func (this *NSPlanDAO) CreatePlan(tx *dbs.Tx, name string, monthlyPrice float32, yearlyPrice float32, config *dnsconfigs.NSPlanConfig) (int64, error) {
	var op = NewNSPlanOperator()
	op.Name = name
	op.MonthlyPrice = monthlyPrice
	op.YearlyPrice = yearlyPrice

	configJSON, err := json.Marshal(config)
	if err != nil {
		return 0, err
	}
	op.Config = configJSON
	op.IsOn = true
	op.State = NSPlanStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdatePlan 修改套餐
func (this *NSPlanDAO) UpdatePlan(tx *dbs.Tx, planId int64, name string, isOn bool, monthlyPrice float32, yearlyPrice float32, config *dnsconfigs.NSPlanConfig) error {
	if planId <= 0 {
		return errors.New("invalid planId")
	}
	var op = NewNSPlanOperator()
	op.Id = planId
	op.Name = name
	op.MonthlyPrice = monthlyPrice
	op.YearlyPrice = yearlyPrice

	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	op.Config = configJSON
	op.IsOn = isOn
	return this.Save(tx, op)
}

// UpdatePlanOrders 修改套餐排序
func (this *NSPlanDAO) UpdatePlanOrders(tx *dbs.Tx, planIds []int64) error {
	var total = len(planIds)
	if total == 0 {
		return nil
	}

	total++

	for index, planId := range planIds {
		err := this.Query(tx).
			Pk(planId).
			Set("order", total-index).
			UpdateQuickly()
		if err != nil {
			return err
		}
	}
	return nil
}

// FindAllPlans 查找所有套餐
func (this *NSPlanDAO) FindAllPlans(tx *dbs.Tx) (result []*NSPlan, err error) {
	_, err = this.Query(tx).
		State(NSPlanStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledPlans 查找所有可用套餐
func (this *NSPlanDAO) FindAllEnabledPlans(tx *dbs.Tx) (result []*NSPlan, err error) {
	_, err = this.Query(tx).
		State(NSPlanStateEnabled).
		Attr("isOn", true).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// ExistPlan 检查套餐是否存在
func (this *NSPlanDAO) ExistPlan(tx *dbs.Tx, planId int64) (bool, error) {
	if planId <= 0 {
		return false, nil
	}

	return this.Query(tx).
		Pk(planId).
		Exist()
}
