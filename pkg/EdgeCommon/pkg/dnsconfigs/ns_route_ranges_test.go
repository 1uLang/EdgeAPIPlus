// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package dnsconfigs

import (
	"github.com/iwind/TeaGo/assert"
	"net"
	"testing"
)

func TestNSRouteRangeIPRange_Contains(t *testing.T) {
	var a = assert.NewAssertion(t)

	// ipv4
	{
		var r = &NSRouteRangeIPRange{
			IPFrom: "192.168.1.100",
			IPTo:   "192.168.3.200",
		}
		err := r.Init()
		if err != nil {
			t.Fatal(err)
		}
		a.IsFalse(r.Contains(net.ParseIP("aaa")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.1.200")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.3.200")))
		a.IsFalse(r.Contains(net.ParseIP("192.168.4.1")))
		a.IsFalse(r.Contains(net.ParseIP("::1")))
	}

	// ipv6
	{
		var prefix = "1:2:3:4:5:6"
		var r = &NSRouteRangeIPRange{
			IPFrom: prefix + ":1:8",
			IPTo:   prefix + ":5:10",
		}
		err := r.Init()
		if err != nil {
			t.Fatal(err)
		}
		a.IsFalse(r.Contains(net.ParseIP("aaa")))
		a.IsTrue(r.Contains(net.ParseIP(prefix + ":3:4")))
		a.IsTrue(r.Contains(net.ParseIP(prefix + ":5:9")))
		a.IsTrue(r.Contains(net.ParseIP(prefix + ":5:10")))
		a.IsTrue(r.Contains(net.ParseIP(prefix + ":4:8")))
		a.IsFalse(r.Contains(net.ParseIP(prefix + ":5:11")))
	}

	{
		var r = &NSRouteRangeCIDR{
			CIDR: "192.168.2.1/24",
		}
		err := r.Init()
		if err != nil {
			t.Fatal(err)
		}
		a.IsFalse(r.Contains(net.ParseIP("aaa")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.2.1")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.2.254")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.2.100")))
		a.IsFalse(r.Contains(net.ParseIP("192.168.3.1")))
		a.IsFalse(r.Contains(net.ParseIP("192.168.1.1")))
	}

	// reverse ipv4
	{
		var r = &NSRouteRangeIPRange{
			IPFrom: "192.168.1.100",
			IPTo:   "192.168.3.200",
		}
		err := r.Init()
		if err != nil {
			t.Fatal(err)
		}
		a.IsFalse(r.Contains(net.ParseIP("aaa")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.1.200")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.3.200")))
		a.IsFalse(r.Contains(net.ParseIP("192.168.4.1")))
	}

	// reverse cidr
	{
		var r = &NSRouteRangeCIDR{
			CIDR: "192.168.2.1/24",
		}
		err := r.Init()
		if err != nil {
			t.Fatal(err)
		}
		a.IsFalse(r.Contains(net.ParseIP("aaa")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.2.1")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.2.254")))
		a.IsTrue(r.Contains(net.ParseIP("192.168.2.100")))
		a.IsFalse(r.Contains(net.ParseIP("192.168.3.1")))
		a.IsFalse(r.Contains(net.ParseIP("192.168.1.1")))
	}
}

type testNSIPResolver struct {
}

func (this *testNSIPResolver) Resolve(ip net.IP) (countryId int64, provinceId int64, cityId int64, providerId int64) {
	return 1, 2, 3, 4
}

func TestNSRouteRangeRegion_Contains(t *testing.T) {
	{
		var r = &NSRouteRangeRegion{
			Regions:   nil,
			Connector: "",
		}
		r.Regions = append(r.Regions, &routeRegion{
			Type: "country",
			Id:   1,
			Name: "1",
		})
		r.Regions = append(r.Regions, &routeRegion{
			Type: "province",
			Id:   2,
			Name: "2",
		})
		r.SetRegionResolver(&testNSIPResolver{})
		err := r.Init()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(r.Contains(net.ParseIP("1.1.1.1")))

		r.Connector = "AND"
		t.Log(r.Contains(net.ParseIP("1.1.1.1")))
	}
}
