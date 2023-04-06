// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package cloudns

type RecordResponse struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Host   string `json:"host"`
	Record string `json:"record"`
	TTL    string `json:"ttl"`
	Status int    `json:"status"`
}
