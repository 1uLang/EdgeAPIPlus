// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package dnsconfigs

import (
	"bytes"
	"encoding/json"
)

func DefaultNSSOAConfig() *NSSOAConfig {
	return &NSSOAConfig{
		Hosts:          nil, // TODO 暂时不实现
		MName:          "",
		RName:          "", // email address, john.doe@example.com => john\.doe.example.com
		Serial:         0,
		RefreshSeconds: 7200,
		RetrySeconds:   3600,
		ExpireSeconds:  1209600, // 14 days
		MinimumTTL:     3600,
	}
}

// NSSOAConfig 参考：https://en.wikipedia.org/wiki/SOA_record
type NSSOAConfig struct {
	Hosts []string `yaml:"hosts" json:"hosts"`

	MName          string `yaml:"mName" json:"mName"`
	RName          string `yaml:"rName" json:"rName"`
	Serial         uint32 `yaml:"serial" json:"serial"`
	RefreshSeconds uint32 `yaml:"refreshSeconds" json:"refreshSeconds"`
	RetrySeconds   uint32 `yaml:"retrySeconds" json:"retrySeconds"`
	ExpireSeconds  uint32 `yaml:"expireSeconds" json:"expireSeconds"`
	MinimumTTL     uint32 `yaml:"minimumTTL" json:"minimumTTL"`
}

func (this *NSSOAConfig) Init() error {
	return nil
}

func (this *NSSOAConfig) IsSame(config2 *NSSOAConfig) bool {
	if config2 == nil {
		return false
	}

	hostsJSON, _ := json.Marshal(this.Hosts)
	host2JSON, _ := json.Marshal(config2.Hosts)

	return ((len(this.Hosts) == 0 && len(config2.Hosts) == 0) || bytes.Equal(hostsJSON, host2JSON)) &&
		this.MName == config2.MName &&
		this.RName == config2.RName &&
		this.RefreshSeconds == config2.RefreshSeconds &&
		this.RetrySeconds == config2.RetrySeconds &&
		this.ExpireSeconds == config2.ExpireSeconds &&
		this.MinimumTTL == config2.MinimumTTL
}
