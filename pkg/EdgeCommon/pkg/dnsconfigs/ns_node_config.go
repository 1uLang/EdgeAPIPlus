// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus

package dnsconfigs

import (
	"context"
	"fmt"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
)

type NSNodeConfig struct {
	Id              int64                         `yaml:"id" json:"id"`
	IsPlus          bool                          `yaml:"isPlus" json:"isPlus"`
	NodeId          string                        `yaml:"nodeId" json:"nodeId"`
	Secret          string                        `yaml:"secret" json:"secret"`
	ClusterId       int64                         `yaml:"clusterId" json:"clusterId"`
	AccessLogRef    *NSAccessLogRef               `yaml:"accessLogRef" json:"accessLogRef"`
	RecursionConfig *NSRecursionConfig            `yaml:"recursionConfig" json:"recursionConfig"`
	DDoSProtection  *ddosconfigs.ProtectionConfig `yaml:"ddosProtection" json:"ddosProtection"`
	AllowedIPs      []string                      `yaml:"allowedIPs" json:"allowedIPs"`
	TimeZone        string                        `yaml:"timeZone" json:"timeZone"` // 自动设置时区
	Hosts           []string                      `yaml:"hosts" json:"hosts"`       // 主机名
	Email           string                        `yaml:"email" json:"email"`
	SOA             *NSSOAConfig                  `yaml:"soa" json:"soa"` // SOA配置
	SOASerial       uint32                        `yaml:"soaSerial" json:"soaSerial"`
	DetectAgents    bool                          `yaml:"detectAgents" json:"detectAgents"` // 是否实时监测Agents

	TCP *serverconfigs.TCPProtocolConfig `yaml:"tcp" json:"tcp"` // TCP配置
	TLS *serverconfigs.TLSProtocolConfig `yaml:"tls" json:"tls"` // TLS配置
	UDP *serverconfigs.UDPProtocolConfig `yaml:"udp" json:"udp"` // UDP配置

	Answer *NSAnswerConfig `yaml:"answer" json:"answer"` // 应答查询

	APINodeAddrs []*serverconfigs.NetworkAddressConfig `yaml:"apiNodeAddrs" json:"apiNodeAddrs"`

	paddedId string
}

func (this *NSNodeConfig) Init(ctx context.Context) error {
	this.paddedId = fmt.Sprintf("%08d", this.Id)

	// accessLog
	if this.AccessLogRef != nil {
		err := this.AccessLogRef.Init()
		if err != nil {
			return err
		}
	}

	// 递归DNS
	if this.RecursionConfig != nil {
		err := this.RecursionConfig.Init()
		if err != nil {
			return err
		}
	}

	// DDoS
	if this.DDoSProtection != nil {
		err := this.DDoSProtection.Init()
		if err != nil {
			return err
		}
	}

	// tcp
	if this.TCP != nil {
		err := this.TCP.Init()
		if err != nil {
			return err
		}
	}

	// tls
	if this.TLS != nil {
		err := this.TLS.Init(ctx)
		if err != nil {
			return err
		}
	}

	// udp
	if this.UDP != nil {
		err := this.UDP.Init()
		if err != nil {
			return err
		}
	}

	// api node addrs
	if len(this.APINodeAddrs) > 0 {
		for _, addr := range this.APINodeAddrs {
			err := addr.Init()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *NSNodeConfig) PaddedId() string {
	return this.paddedId
}
