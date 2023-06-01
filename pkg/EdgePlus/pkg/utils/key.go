// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import timeutil "github.com/iwind/TeaGo/utils/time"

type Key struct {
	Id           string          `json:"id"`           // 用户ID
	DayFrom      string          `json:"dayFrom"`      // 开始日期
	DayTo        string          `json:"dayTo"`        // 结束日期
	MacAddresses []string        `json:"macAddresses"` // MAC地址，老的授权方式
	RequestCode  string          `json:"requestCode"`  // 授权请求码
	Hostname     string          `json:"hostname"`     // 主机名
	Company      string          `json:"company"`      // 公司名
	Nodes        int             `json:"nodes"`        // 节点数
	UpdatedAt    int64           `json:"updatedAt"`    // 更新时间
	Components   []ComponentCode `json:"components"`   // 组件
	Edition      Edition         `json:"edition"`      // 授权版本
	Email        string          `json:"email"`        // 联系人邮箱
}

func (this *Key) IsValid() bool {
	return this.DayTo >= timeutil.Format("Y-m-d")
}
