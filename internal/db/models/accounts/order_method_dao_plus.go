// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package accounts

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
)

// EnableOrderMethod 启用条目
func (this *OrderMethodDAO) EnableOrderMethod(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", OrderMethodStateEnabled).
		Update()
	return err
}

// DisableOrderMethod 禁用条目
func (this *OrderMethodDAO) DisableOrderMethod(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", OrderMethodStateDisabled).
		Update()
	return err
}

// FindEnabledOrderMethod 查找支付方式
func (this *OrderMethodDAO) FindEnabledOrderMethod(tx *dbs.Tx, id int64) (*OrderMethod, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", OrderMethodStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*OrderMethod), err
}

// FindEnabledBasicOrderMethod 查找支付方式基本信息
func (this *OrderMethodDAO) FindEnabledBasicOrderMethod(tx *dbs.Tx, id int64) (*OrderMethod, error) {
	result, err := this.Query(tx).
		Pk(id).
		Result("id", "code", "url", "isOn", "secret").
		Attr("state", OrderMethodStateEnabled).
		Find()
	if err != nil || result == nil {
		return nil, err
	}
	return result.(*OrderMethod), err
}

// FindEnabledOrderMethodWithCode 根据代号查找支付方式
func (this *OrderMethodDAO) FindEnabledOrderMethodWithCode(tx *dbs.Tx, code string) (*OrderMethod, error) {
	result, err := this.Query(tx).
		Attr("code", code).
		Attr("state", OrderMethodStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*OrderMethod), err
}

// FindOrderMethodName 根据主键查找名称
func (this *OrderMethodDAO) FindOrderMethodName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateMethod 创建支付方式
func (this *OrderMethodDAO) CreateMethod(tx *dbs.Tx, name string, code string, url string, description string) (int64, error) {
	var op = NewOrderMethodOperator()
	op.Name = name
	op.Code = code
	op.Url = url
	op.Description = description
	op.Secret = utils.Sha1RandomString()
	op.IsOn = true
	op.State = OrderMethodStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateMethod 修改支付方式
func (this *OrderMethodDAO) UpdateMethod(tx *dbs.Tx, methodId int64, name string, code string, url string, description string, isOn bool) error {
	if methodId <= 0 {
		return errors.New("invalid methodId")
	}

	var op = NewOrderMethodOperator()
	op.Id = methodId
	op.Name = name
	op.Code = code
	op.Url = url
	op.Description = description
	op.IsOn = isOn
	return this.Save(tx, op)
}

// UpdateMethodOrders 修改排序
func (this *OrderMethodDAO) UpdateMethodOrders(tx *dbs.Tx, methodIds []int64) error {
	var maxOrder = len(methodIds)
	for index, methodId := range methodIds {
		err := this.Query(tx).
			Pk(methodId).
			Set("order", maxOrder-index).
			UpdateQuickly()
		if err != nil {
			return err
		}
	}
	return nil
}

// FindAllEnabledMethodOrders 查找所有支付方式
func (this *OrderMethodDAO) FindAllEnabledMethodOrders(tx *dbs.Tx) (result []*OrderMethod, err error) {
	_, err = this.Query(tx).
		State(OrderMethodStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAndOnMethodOrders 查找所有已启用的支付方式
func (this *OrderMethodDAO) FindAllEnabledAndOnMethodOrders(tx *dbs.Tx) (result []*OrderMethod, err error) {
	_, err = this.Query(tx).
		State(OrderMethodStateEnabled).
		Attr("isOn", true).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}
