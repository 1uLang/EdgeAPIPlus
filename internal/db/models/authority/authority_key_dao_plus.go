// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus

package authority

import (
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	plusutils "github.com/TeaOSLab/EdgePlus/pkg/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// 更新Plus数值
func init() {
	dbs.OnReadyDone(func() {
		isPlus, _ := SharedAuthorityKeyDAO.IsPlus(nil)
		teaconst.IsPlus = isPlus

		var ticker = time.NewTicker(5 * time.Minute)
		goman.New(func() {
			for range ticker.C {
				isPlus, err := SharedAuthorityKeyDAO.IsPlus(nil)
				if err != nil {
					remotelogs.Error("AuthorityKeyDAO", "check isPlus failed: "+err.Error())
				}
				teaconst.IsPlus = isPlus
			}
		})
	})
}

// IsPlus 判断是否为商业版
func (this *AuthorityKeyDAO) IsPlus(tx *dbs.Tx) (bool, error) {
	key, err := this.ReadKey(tx)
	if err != nil {
		return false, err
	}
	if key == nil || len(key.Value) == 0 {
		return false, nil
	}

	m, err := plusutils.DecodeKey([]byte(key.Value))
	if err != nil {
		return false, err
	}

	teaconst.IsPlus = m.DayTo >= timeutil.Format("Y-m-d")
	teaconst.MaxNodes = types.Int32(m.Nodes)

	return teaconst.IsPlus, nil
}

// UpdateKey 设置Key
func (this *AuthorityKeyDAO) UpdateKey(tx *dbs.Tx, value string, dayFrom string, dayTo string, hostname string, macAddresses []string, company string) error {
	one, err := this.Query(tx).
		AscPk().
		Find()
	if err != nil {
		return err
	}
	var op = NewAuthorityKeyOperator()
	if one != nil {
		op.Id = one.(*AuthorityKey).Id
	}
	op.Value = value
	op.DayFrom = dayFrom
	op.DayTo = dayTo
	op.Hostname = hostname

	if len(macAddresses) == 0 {
		macAddresses = []string{}
	}
	macAddressesJSON, err := json.Marshal(macAddresses)
	if err != nil {
		return err
	}

	op.MacAddresses = macAddressesJSON
	op.Company = company
	op.UpdatedAt = time.Now().Unix()

	return this.Save(tx, op)
}

// ReadKey 读取Key
func (this *AuthorityKeyDAO) ReadKey(tx *dbs.Tx) (key *AuthorityKey, err error) {
	one, err := this.Query(tx).
		AscPk().
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	key = one.(*AuthorityKey)

	// 顺便更新相关变量
	if key.DayTo >= timeutil.Format("Y-m-d") {
		teaconst.IsPlus = true
	}

	return
}

// ResetKey 重置Key
func (this *AuthorityKeyDAO) ResetKey(tx *dbs.Tx) error {
	_, err := this.Query(tx).
		Delete()
	return err
}
