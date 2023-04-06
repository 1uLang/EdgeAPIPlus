package nameservers

import (
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

const (
	NSUserPlanStateEnabled  = 1 // 已启用
	NSUserPlanStateDisabled = 0 // 已禁用
)

type NSUserPlanPeriodUnit = string

const (
	NSUserPlanPeriodUnitMonthly NSUserPlanPeriodUnit = "monthly"
	NSUserPlanPeriodUnitYearly  NSUserPlanPeriodUnit = "yearly"
)

type NSUserPlanDAO dbs.DAO

func NewNSUserPlanDAO() *NSUserPlanDAO {
	return dbs.NewDAO(&NSUserPlanDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSUserPlans",
			Model:  new(NSUserPlan),
			PkName: "id",
		},
	}).(*NSUserPlanDAO)
}

var SharedNSUserPlanDAO *NSUserPlanDAO

func init() {
	dbs.OnReady(func() {
		SharedNSUserPlanDAO = NewNSUserPlanDAO()
	})
}

// EnableNSUserPlan 启用条目
func (this *NSUserPlanDAO) EnableNSUserPlan(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSUserPlanStateEnabled).
		Update()
	return err
}

// DisableNSUserPlan 禁用条目
func (this *NSUserPlanDAO) DisableNSUserPlan(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSUserPlanStateDisabled).
		Update()
	return err
}

// FindEnabledNSUserPlan 查找启用中的条目
func (this *NSUserPlanDAO) FindEnabledNSUserPlan(tx *dbs.Tx, id int64) (*NSUserPlan, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(NSUserPlanStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSUserPlan), err
}

// CreateUserPlan 创建用户套餐
func (this *NSUserPlanDAO) CreateUserPlan(tx *dbs.Tx, userId int64, planId int64, dayFrom string, dayTo string, periodUnit NSUserPlanPeriodUnit) (int64, error) {
	var op = NewNSUserPlanOperator()
	op.UserId = userId
	op.PlanId = planId
	op.DayFrom = dayFrom
	op.DayTo = dayTo
	op.PeriodUnit = periodUnit
	op.State = NSUserPlanStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateUserPlan 修改用户套餐
func (this *NSUserPlanDAO) UpdateUserPlan(tx *dbs.Tx, userPlanId int64, planId int64, dayFrom string, dayTo string, periodUnit NSUserPlanPeriodUnit) error {
	if userPlanId <= 0 {
		return errors.New("invalid userPlanId")
	}
	var op = NewNSUserPlanOperator()
	op.Id = userPlanId
	op.PlanId = planId
	op.DayFrom = dayFrom
	op.DayTo = dayTo
	op.PeriodUnit = periodUnit
	return this.Save(tx, op)
}

// CountUserPlans 计算用户套餐数量
func (this *NSUserPlanDAO) CountUserPlans(tx *dbs.Tx, userId int64, planId int64, periodUnit string, isExpired bool, expireDays int32) (int64, error) {
	var query = this.Query(tx).
		State(NSUserPlanStateEnabled)

	if userId > 0 {
		query.Attr("userId", userId)
	}
	if planId > 0 {
		query.Attr("planId", planId)
	}
	if len(periodUnit) > 0 {
		query.Attr("periodUnit", periodUnit)
	}
	if isExpired {
		query.Lt("dayTo", timeutil.Format("Ymd"))
	} else if expireDays == 0 {
		query.Gte("dayTo", timeutil.Format("Ymd"))
	} else if expireDays > 0 {
		query.Gte("dayTo", timeutil.Format("Ymd"))
		query.Lte("dayTo", timeutil.Format("Ymd", time.Now().AddDate(0, 0, int(expireDays))))
	}

	query.Where("planId IN (SELECT id FROM " + SharedNSPlanDAO.Table + " WHERE state=1)")

	return query.Count()
}

// ListUserPlans 列出单页用户套餐
func (this *NSUserPlanDAO) ListUserPlans(tx *dbs.Tx, userId int64, planId int64, periodUnit string, isExpired bool, expireDays int32, offset int64, size int64) (result []*NSUserPlan, err error) {
	var query = this.Query(tx)

	if userId > 0 {
		query.Attr("userId", userId)
	}
	if planId > 0 {
		query.Attr("planId", planId)
	}
	if len(periodUnit) > 0 {
		query.Attr("periodUnit", periodUnit)
	}
	if isExpired {
		query.Lt("dayTo", timeutil.Format("Ymd"))
	} else if expireDays == 0 {
		query.Gte("dayTo", timeutil.Format("Ymd"))
	} else if expireDays > 0 {
		query.Gte("dayTo", timeutil.Format("Ymd"))
		query.Lte("dayTo", timeutil.Format("Ymd", time.Now().AddDate(0, 0, int(expireDays))))
	}

	query.Where("planId IN (SELECT id FROM " + SharedNSPlanDAO.Table + " WHERE state=1)")

	_, err = query.
		State(NSUserPlanStateEnabled).
		Desc("dayFrom").
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindUserPlan 查找用户对应的套餐
func (this *NSUserPlanDAO) FindUserPlan(tx *dbs.Tx, userId int64) (*NSUserPlan, error) {
	if userId <= 0 {
		return nil, nil
	}

	one, err := this.Query(tx).
		State(NSUserPlanStateEnabled).
		Attr("userId", userId).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*NSUserPlan), nil
}

// FindUserPlanConfig 用户配置
// 需要保证非 error 的情况下，一定会返回一个不为空的 NSPlanConfig
func (this *NSUserPlanDAO) FindUserPlanConfig(tx *dbs.Tx, userId int64) (*dnsconfigs.NSPlanConfig, error) {
	// 查找用户套餐
	userPlan, err := this.FindUserPlan(tx, userId)
	if err != nil {
		return nil, err
	}

	if userPlan != nil &&
		len(userPlan.DayTo) > 0 &&
		userPlan.DayTo >= timeutil.Format("Ymd") /** 在有效期内 **/ {
		var planId = int64(userPlan.PlanId)
		if planId > 0 {
			plan, err := SharedNSPlanDAO.FindEnabledNSPlan(tx, planId)
			if err != nil {
				return nil, err
			}
			if plan != nil &&
				len(plan.Config) > 0 &&
				plan.IsOn {
				var config = dnsconfigs.DefaultNSUserPlanConfig()
				err = json.Unmarshal(plan.Config, config)
				if err != nil {
					return nil, errors.New("decode plan config failed '" + err.Error() + "'")
				}
				return config, nil
			}
		}
	}

	// 从用户全局设置中读取
	userConfig, err := models.SharedSysSettingDAO.ReadNSUserConfig(tx)
	if err != nil {
		return nil, err
	}
	if userConfig != nil && userConfig.DefaultPlanConfig != nil {
		return userConfig.DefaultPlanConfig, nil
	}

	return dnsconfigs.DefaultNSUserPlanConfig(), nil
}
