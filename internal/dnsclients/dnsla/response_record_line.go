// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsla

type RecordLineResponse struct {
	BaseResponse

	Datas []struct {
		Value string `json:"value"`
		Text  string `json:"text"`
	} `json:"datas"`
}
