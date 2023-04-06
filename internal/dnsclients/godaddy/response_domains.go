// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package godaddy

type DomainsResponse []struct {
	DomainId int64  `json:"domainId"`
	Domain   string `json:"domain"`
	Status   string `json:"status"`
}
