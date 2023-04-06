// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnscom

type DomainSearchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Data []struct {
			Domains   string `json:"domains"`
			DomainsID string `json:"domainsID"`
			State     int    `json:"state"`
		} `json:"data"`
		Page      int `json:"page"`
		PageCount int `json:"pageCount"`
	} `json:"data"`
}
