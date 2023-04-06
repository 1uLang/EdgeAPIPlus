// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package setup

import (
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
)

// 检查自建DNS全局设置
func (this *SQLExecutor) checkNS(db *dbs.DB) error {
	// 访问日志
	{
		one, err := db.FindOne("SELECT id FROM edgeSysSettings WHERE code=? LIMIT 1", systemconfigs.SettingCodeNSAccessLogSetting)
		if err != nil {
			return err
		}
		if len(one) == 0 {
			ref := &dnsconfigs.NSAccessLogRef{
				IsPrior:           false,
				IsOn:              true,
				LogMissingDomains: false,
			}
			refJSON, err := json.Marshal(ref)
			if err != nil {
				return err
			}
			_, err = db.Exec("INSERT edgeSysSettings (code, value) VALUES (?, ?)", systemconfigs.SettingCodeNSAccessLogSetting, refJSON)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
