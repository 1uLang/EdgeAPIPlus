// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsla

import (
	"encoding/json"
	"testing"
)

func TestDomainListResponse(t *testing.T) {
	var bodyJSON = []byte(`{
  "status": {
    "code": 300,
    "name": "操作成功",
    "message": "",
    "request_time": "2022-08-07 07:43:34"
  },
  "total": {
    "all_total": 3,
    "data_total": 3,
    "skip_number": 0
  },
  "datas": [
    {
      "domainid": 6772732,
      "userid": 459662,
      "domainname": "hello3.com",
      "grade": "免费套餐",
      "domain_status": "正常",
      "domain_active": "yes",
      "groupid": 0,
      "nsstate": "未知",
      "createtime": "2022-08-07 07:42:27"
    },
    {
      "domainid": 6772731,
      "userid": 459662,
      "domainname": "hello2.com",
      "grade": "免费套餐",
      "domain_status": "正常",
      "domain_active": "yes",
      "groupid": 0,
      "nsstate": "未知",
      "createtime": "2022-08-07 07:42:19"
    },
    {
      "domainid": 6772358,
      "userid": 459662,
      "domainname": "hello1234.com",
      "grade": "免费套餐",
      "domain_status": "正常",
      "domain_active": "yes",
      "groupid": 0,
      "nsstate": "错误",
      "createtime": "2022-08-07 11:42:05"
    }
  ]
}`)
	var resp = &DomainListResponse{}
	err := json.Unmarshal(bodyJSON, resp)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", resp)
	t.Log(resp.Success(), resp.Error())
}
