// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnscom

type CreateRecordResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DomainID int64  `json:"domainID"`
		RecordID int64  `json:"recordID"`
		Record   string `json:"record"`
		Type     string `json:"type"`
		TTL      int    `json:"TTL"`
		State    int    `json:"state"`
	} `json:"data"`
}
