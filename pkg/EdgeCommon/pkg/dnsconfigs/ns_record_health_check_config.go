// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsconfigs

// NSRecordHealthCheckConfig 单个记录健康检查配置
type NSRecordHealthCheckConfig struct {
	IsOn           bool  `json:"isOn"`           // 是否启用
	Port           int32 `json:"port"`           // 端口
	TimeoutSeconds int   `json:"timeoutSeconds"` // 超时秒数
	CountUp        int   `json:"countUp"`        // 连续上线次数
	CountDown      int   `json:"countDown"`      // 连续下线次数
}

func NewNSRecordHealthCheckConfig() *NSRecordHealthCheckConfig {
	return &NSRecordHealthCheckConfig{
		IsOn:           false,
		Port:           0,
		TimeoutSeconds: 0,
		CountUp:        0,
		CountDown:      0,
	}
}

func (this *NSRecordHealthCheckConfig) Init() error {
	return nil
}
