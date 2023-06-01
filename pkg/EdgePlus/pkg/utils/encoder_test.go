// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import (
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
	"time"
)

func TestEncodeMap(t *testing.T) {
	{
		t.Log(Encode([]byte("123")))
	}
	{
		s, err := EncodeMap(maps.Map{"a": 1})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(s)

		t.Log(Decode([]byte(s)))
	}
}

func TestEncodeKey(t *testing.T) {
	var key = &Key{
		DayFrom:      "2023-06-01",
		DayTo:        "2023-06-01",
		MacAddresses: []string{"*"},
		Hostname:     "*",
		Company:      "CloudWAF",
		Nodes:        10,
		UpdatedAt:    time.Now().Unix(),
		Components:   []string{"*"},
	}
	encodedString, err := EncodeKey(key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("encoded:", encodedString)

	key, err = DecodeKey([]byte(encodedString))
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(key, t)
}
