// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnscom

type IPAreaViewListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Name   string `json:"Name"`
		ViewID int64  `json:"viewID"`
	} `json:"data"`
}
