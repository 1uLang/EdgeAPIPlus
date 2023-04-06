// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
package models

import (
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
)

func (this *SysSettingDAO) ReadNSUserConfig(tx *dbs.Tx) (*dnsconfigs.NSUserConfig, error) {
	valueJSON, err := this.ReadSetting(tx, systemconfigs.SettingCodeNSUserConfig)
	if err != nil {
		return nil, err
	}

	if len(valueJSON) == 0 {
		return dnsconfigs.DefaultNSUserConfig(), nil
	}

	var config = dnsconfigs.DefaultNSUserConfig()
	err = json.Unmarshal(valueJSON, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
