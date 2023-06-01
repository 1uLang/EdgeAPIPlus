// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsconfigs

// NSRecordsHealthCheckConfig 记录健康检查整体配置
type NSRecordsHealthCheckConfig struct {
	IsOn           bool  `json:"isOn"`           // 是否启用
	Port           int32 `json:"port"`           // 默认端口
	TimeoutSeconds int   `json:"timeoutSeconds"` // 默认超时秒数
	CountUp        int   `json:"countUp"`        // 默认连续上线次数
	CountDown      int   `json:"countDown"`      // 默认连续下线次数
}

func NewNSRecordsHealthCheckConfig() *NSRecordsHealthCheckConfig {
	return &NSRecordsHealthCheckConfig{
		IsOn:           false,
		Port:           80,
		TimeoutSeconds: 5,
		CountUp:        1,
		CountDown:      3,
	}
}

func (this *NSRecordsHealthCheckConfig) Init() error {
	return nil
}
