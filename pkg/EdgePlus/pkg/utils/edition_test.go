// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils

import "testing"

func TestCompareEdition(t *testing.T) {
	t.Log(CompareEdition("", ""))
	t.Log(CompareEdition("", "pro"))
	t.Log(CompareEdition("basic", "pro"))
	t.Log(CompareEdition("pro", "pro"))
	t.Log(CompareEdition("ent", "pro"))
	t.Log(CompareEdition("pro", "basic"))
}
