// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/systemconfigs"
	"github.com/1uLang/EdgeCommon/pkg/userconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// CreateOrder 创建订单
func (this *UserOrderDAO) CreateOrder(tx *dbs.Tx, adminId int64, userId int64, orderType userconfigs.OrderType, methodId int64, amount float32) (orderId int64, code string, err error) {
	// 查询过期时间
	configJSON, err := models.SharedSysSettingDAO.ReadSetting(tx, systemconfigs.SettingCodeUserOrderConfig)
	if err != nil {
		return 0, "", err
	}
	var config = userconfigs.DefaultUserOrderConfig()
	if len(configJSON) == 0 {
		err = json.Unmarshal(configJSON, config)
		if err != nil {
			return 0, "", errors.New("decode order config failed: " + err.Error())
		}
	}

	// 保存订单
	var op = NewUserOrderOperator()
	op.UserId = userId
	op.Type = orderType
	op.MethodId = methodId
	op.Amount = amount
	op.Status = userconfigs.OrderStatusNone

	if config.OrderLife != nil && config.OrderLife.Count > 0 {
		op.ExpiredAt = time.Now().Unix() + int64(config.OrderLife.Duration().Seconds())
	} else {
		op.ExpiredAt = time.Now().Unix() + 3600 /** 默认一个小时 **/
	}

	op.State = UserOrderStateEnabled
	orderId, err = this.SaveInt64(tx, op)
	if err != nil {
		return 0, "", err
	}

	var orderCode = timeutil.Format("Ymd") + fmt.Sprintf("%08d", orderId) // 16 bytes
	err = this.Query(tx).
		Pk(orderId).
		Set("code", orderCode).
		UpdateQuickly()
	if err != nil {
		return 0, "", err
	}

	// 生成订单日志
	err = SharedUserOrderLogDAO.CreateOrderLog(tx, adminId, userId, orderId, userconfigs.OrderStatusNone)
	if err != nil {
		return 0, "", err
	}

	return orderId, orderCode, nil
}

// FindEnabledOrder 查找某个订单
func (this *UserOrderDAO) FindEnabledOrder(tx *dbs.Tx, orderId int64) (*UserOrder, error) {
	one, err := this.Query(tx).
		Pk(orderId).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*UserOrder), nil
}

// CancelOrder 取消订单
func (this *UserOrderDAO) CancelOrder(tx *dbs.Tx, adminId int64, userId int64, orderId int64) error {
	status, err := this.Query(tx).
		Pk(orderId).
		Result("status").
		FindStringCol("")
	if err != nil {
		return err
	}
	if status != userconfigs.OrderStatusNone {
		return errors.New("can not cancel the order with status '" + status + "'")
	}

	err = this.Query(tx).
		Pk(orderId).
		Set("status", userconfigs.OrderStatusCancelled).
		Set("cancelledAt", time.Now().Unix()).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return SharedUserOrderLogDAO.CreateOrderLog(tx, adminId, userId, orderId, userconfigs.OrderStatusCancelled)
}

// FinishOrder 完成订单
// 不需要检查过期时间，因为用户可能在支付页面停留非常久后才完成支付
func (this *UserOrderDAO) FinishOrder(tx *dbs.Tx, adminId int64, userId int64, orderId int64) error {
	// 检查订单状态
	order, err := SharedUserOrderDAO.FindEnabledOrder(tx, orderId)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("can not find order")
	}

	if order.Status != userconfigs.OrderStatusNone {
		return errors.New("you can not finish the order, cause order status is '" + order.Status + "'")
	}

	// 用户账户
	account, err := SharedUserAccountDAO.FindUserAccountWithUserId(tx, int64(order.UserId))
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("user account not found")
	}

	switch order.Type {
	case userconfigs.OrderTypeCharge:
		err = SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), float32(order.Amount), userconfigs.AccountEventTypeCharge, "充值，订单号："+order.Code, maps.Map{
			"orderCode": order.Code,
		})
		if err != nil {
			return err
		}
	}

	err = this.Query(tx).
		Pk(orderId).
		Set("status", userconfigs.OrderStatusFinished).
		Set("finishedAt", time.Now().Unix()).
		UpdateQuickly()

	if err != nil {
		return err
	}

	return SharedUserOrderLogDAO.CreateOrderLog(tx, adminId, userId, orderId, userconfigs.OrderStatusFinished)
}

// CountEnabledUserOrders 计算订单数量
func (this *UserOrderDAO) CountEnabledUserOrders(tx *dbs.Tx, userId int64, status userconfigs.OrderStatus, keyword string) (int64, error) {
	var query = this.Query(tx).
		State(UserOrderStateEnabled)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if len(status) > 0 {
		query.Attr("status", status)
	}
	if len(keyword) > 0 {
		query.Where("(code LIKE :keyword)")
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	return query.Count()
}

// ListEnabledUserOrders 列出单页订单
func (this *UserOrderDAO) ListEnabledUserOrders(tx *dbs.Tx, userId int64, status userconfigs.OrderStatus, keyword string, offset int64, size int64) (result []*UserOrder, err error) {
	var query = this.Query(tx).
		State(UserOrderStateEnabled)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if len(status) > 0 {
		query.Attr("status", status)
	}
	if len(keyword) > 0 {
		query.Where("(code LIKE :keyword)")
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindUserOrderIdWithCode 根据订单号查找订单ID
func (this *UserOrderDAO) FindUserOrderIdWithCode(tx *dbs.Tx, code string) (int64, error) {
	if len(code) == 0 {
		return 0, nil
	}
	return this.Query(tx).
		ResultPk().
		State(UserOrderStateEnabled).
		Attr("code", code).
		FindInt64Col(0)
}

// FindUserOrderWithCode 根据订单号查找订单
func (this *UserOrderDAO) FindUserOrderWithCode(tx *dbs.Tx, code string) (*UserOrder, error) {
	if len(code) == 0 {
		return nil, nil
	}
	one, err := this.Query(tx).
		State(UserOrderStateEnabled).
		Attr("code", code).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserOrder), nil
}
