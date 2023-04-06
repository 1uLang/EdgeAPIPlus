// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/iwind/TeaGo/maps"
)

const (
	ProviderTypeGoDaddy ProviderType = "godaddy" // Godaddy DNS
	ProviderTypeClouDNS ProviderType = "cloudns" // ClouDNS
	ProviderTypeDNSCom  ProviderType = "dnscom"  // DNS.COM
	ProviderTypeDNSLA   ProviderType = "dnsla"   // DNSLA
)

// FindProvider 查找服务商实例
func FindProvider(providerType ProviderType) ProviderInterface {
	switch providerType {
	case ProviderTypeDNSPod:
		return &DNSPodProvider{}
	case ProviderTypeAliDNS:
		return &AliDNSProvider{}
	case ProviderTypeHuaweiDNS:
		return &HuaweiDNSProvider{}
	case ProviderTypeCloudFlare:
		return &CloudFlareProvider{}
	case ProviderTypeLocalEdgeDNS:
		return &LocalEdgeDNSProvider{}
	case ProviderTypeEdgeDNSAPI:
		return &EdgeDNSAPIProvider{}
	case ProviderTypeCustomHTTP:
		return &CustomHTTPProvider{}
	case ProviderTypeGoDaddy:
		return &GoDaddyProvider{}
	case ProviderTypeClouDNS:
		return &ClouDNSProvider{}
	case ProviderTypeDNSCom:
		return &DNSComProvider{}
	case ProviderTypeDNSLA:
		return &DNSLaProvider{}
	}

	return nil
}

func filterTypeMaps(typeMaps []maps.Map) []maps.Map {
	if !teaconst.IsPlus {
		return typeMaps
	}
	return []maps.Map{
		{
			"name":        "阿里云DNS",
			"code":        ProviderTypeAliDNS,
			"description": "阿里云提供的DNS服务。",
		},
		{
			"name":        "DNSPod",
			"code":        ProviderTypeDNSPod,
			"description": "DNSPod提供的DNS服务。",
		},
		{
			"name":        "华为云DNS",
			"code":        ProviderTypeHuaweiDNS,
			"description": "华为云解析DNS。",
		},
		{
			"name":        "CloudFlare DNS",
			"code":        ProviderTypeCloudFlare,
			"description": "CloudFlare提供的DNS服务。",
		},
		{
			"name":        "GoDaddy",
			"code":        ProviderTypeGoDaddy,
			"description": "Plus专享。GoDaddy提供的DNS服务",
		},
		{
			"name":        "ClouDNS",
			"code":        ProviderTypeClouDNS,
			"description": "Plus专享。ClouDNS提供的DNS服务。注意ClouDNS的API可能只对其付费用户开放。",
		},
		{
			"name":        "帝恩思(DNS.COM)",
			"code":        ProviderTypeDNSCom,
			"description": "Plus专享。DNS.COM提供的DNS服务。注意DNS.COM的API可能只对其付费用户开放。",
		},
		{
			"name":        "DNS.LA",
			"code":        ProviderTypeDNSLA,
			"description": "Plus专享。DNS.LA提供的DNS服务。",
		},
		{
			"name":        "EdgeDNS",
			"code":        ProviderTypeLocalEdgeDNS,
			"description": "当前GoEdge商业版系统提供的智能DNS服务。",
		},
		{
			"name":        "EdgeDNS API",
			"code":        ProviderTypeEdgeDNSAPI,
			"description": "通过API连接其他GoEdge商业版系统提供的DNS服务。",
		},
	}
}
