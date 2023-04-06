// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package godaddy

type RecordsResponse []struct {
	Data string `json:"data"`
	Name string `json:"name"`
	TTL  int32  `json:"ttl"`
	Type string `json:"type"`
}
