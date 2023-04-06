// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package accounts

import (
	"github.com/1uLang/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

// CreateOrderLog 创建订单日志
func (this *UserOrderLogDAO) CreateOrderLog(tx *dbs.Tx, adminId int64, userId int64, orderId int64, status userconfigs.OrderStatus) error {
	var op = NewUserOrderLogOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.OrderId = orderId
	op.Status = status
	op.CreatedAt = time.Now().Unix()
	return this.Save(tx, op)
}
