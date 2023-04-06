// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsla

type DomainListResponse struct {
	BaseResponse

	Total struct {
		AllTotal   int `json:"all_total"`
		DataTotal  int `json:"data_total"`
		SkipNumber int `json:"skip_number"`
	} `json:"total"`
	Datas []struct {
		DomainId   int64  `json:"domainid"`
		DomainName string `json:"domainname"`
	} `json:"datas"`
}
