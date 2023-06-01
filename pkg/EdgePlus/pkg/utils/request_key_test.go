// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils_test

import (
	"github.com/TeaOSLab/EdgePlus/pkg/utils"
	"github.com/iwind/TeaGo/types"
	"testing"
)

func TestGenerateRequestKey(t *testing.T) {
	requestKey, err := utils.GenerateRequestKey()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("request key: %+v", requestKey)
}

func TestGenerateRequestCode(t *testing.T) {
	requestCode, err := utils.GenerateRequestCode()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("request code:", "["+types.String(len(requestCode))+"]", requestCode)
	requestKey, err := utils.DecodeRequestCode(requestCode)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", requestKey)
	t.Log("mac addresses:", len(requestKey.MacAddresses))
}

func TestDecodeRequestCode(t *testing.T) {
	var requestCode = `F4BqUMBxDHPFsd4mIDUiSfiRor473+ctxycygBwxZUyqDZppJrlAjnT5E6qyH7Yb64icvlkCqiEPYbOkxh9TUhWHuoqsGAKcO+6vFaelBeojnlVXkg==`
	requestKey, err := utils.DecodeRequestCode(requestCode)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", requestKey)
}

func TestValidateRequestCode(t *testing.T) {
	{
		ok, errorCode := utils.ValidateRequestCode("123456")
		t.Log("ok:", ok, "errorCode:", errorCode)
	}

	{
		requestCode, err := utils.GenerateRequestCode()
		if err != nil {
			t.Fatal(err)
		}
		ok, errorCode := utils.ValidateRequestCode(requestCode)
		t.Log("ok:", ok, "errorCode:", errorCode)
	}

	{
		var requestCode = "F4BqUMBxDHPFsd4mIDUiSfiRpr471egtmnMyhx0xbxQeuEsuqXRiFtveCJGaELzffDATN5ULoP+Q/Y3NxsXNsvtxl9VkTA4VFq1s7b83BJVy6h3hKgwvhVw9H2upOf9aouD26JFZwr0ncM+cQGda3z64wOg3TFj8KhoM+ixaFY9SO0o3fg+0R8tKxA6rjGn/Do/CgKJTb4fF/tGGZ6QFY3UbO4KObaDmJrAQWag9IGKE5/GGOyBYWI9S45Auf6ee39X5JToDJHVJt3BV1fNNu3D9OrS+mg2SKLHhQdps7E5zor+K7Shhx8KV85qkdEImR+BA2rrxEDfcJz6+lQ=="
		ok, errorCode := utils.ValidateRequestCode(requestCode)
		t.Log("ok:", ok, "errorCode:", errorCode)
	}
}
