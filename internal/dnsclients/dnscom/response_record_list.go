// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnscom

type RecordListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Data []struct {
			RecordID   int    `json:"recordID"`
			Record     string `json:"record"`
			Type       string `json:"type"`
			State      int    `json:"state"`
			ViewID     int64  `json:"viewID"`
			AreaViewID int64  `json:"areaViewID"`
			ISPViewID  int64  `json:"ISPViewID"`
			Value      string `json:"value"`
			TTL        int    `json:"TTL"`
		} `json:"data"`
		Page      int `json:"page"`
		PageSize  int `json:"pageSize"`
		PageCount int `json:"pageCount"`
	} `json:"data"`
}
