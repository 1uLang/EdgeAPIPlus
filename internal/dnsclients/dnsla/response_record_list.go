// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsla

type RecordListResponse struct {
	BaseResponse

	Datas []struct {
		RecordId   int    `json:"recordid"`
		Host       string `json:"host"`
		RecordData string `json:"record_data"`
		TTL        int    `json:"ttl"`
		RecordType string `json:"record_type"`
		RecordLine string `json:"record_line"`
	} `json:"datas"`
}
