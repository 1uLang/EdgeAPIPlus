// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package dnsclients_test

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestDNSComProvider_GetDomains(t *testing.T) {
	provider, err := testDNSComProvider()
	if err != nil {
		t.Fatal(err)
	}

	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(domains)
}

func TestDNSComProvider_GetRecords(t *testing.T) {
	provider, err := testDNSComProvider()
	if err != nil {
		t.Fatal(err)
	}

	records, err := provider.GetRecords("goedge.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestDNSComProvider_GetRoutes(t *testing.T) {
	provider, err := testDNSComProvider()
	if err != nil {
		t.Fatal(err)
	}

	routes, err := provider.GetRoutes("goedge.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestDNSComProvider_QueryRecord(t *testing.T) {
	provider, err := testDNSComProvider()
	if err != nil {
		t.Fatal(err)
	}

	for _, recordName := range []string{"www", "test", "@", ""} {
		t.Log("===", recordName, "===")
		record, err := provider.QueryRecord("goedge.cn", recordName, dnstypes.RecordTypeA)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
}

func TestDNSComProvider_AddRecord(t *testing.T) {
	provider, err := testDNSComProvider()
	if err != nil {
		t.Fatal(err)
	}

	{
		var record = &dnstypes.Record{
			Id:    "",
			Name:  "test1",
			Type:  dnstypes.RecordTypeA,
			Value: "192.168.1.100",
			Route: "285344768", // 285344768
			TTL:   7200,
		}
		err := provider.AddRecord("goedge.cn", record)
		if err != nil {
			t.Fatal(err)
		}
	}
	/**{
		var record = &dnstypes.Record{
			Id:    "",
			Name:  "test2",
			Type:  dnstypes.RecordTypeCNAME,
			Value: "goedge.cn.",
			Route: "",
			TTL:   0,
		}
		err := provider.AddRecord("goedge.cn", record)
		if err != nil {
			t.Fatal(err)
		}
	}**/
}

func TestDNSComProvider_UpdateRecord(t *testing.T) {
	provider, err := testDNSComProvider()
	if err != nil {
		t.Fatal(err)
	}

	var record = &dnstypes.Record{
		Id: "535669373",
	}
	var newRecord = &dnstypes.Record{
		Id:    "",
		Name:  "test1",
		Type:  dnstypes.RecordTypeA,
		Value: "192.168.1.101",
		Route: "285345792",
		TTL:   3600,
	}
	err = provider.UpdateRecord("goedge.cn", record, newRecord)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDNSComProvider_DeleteRecord(t *testing.T) {
	provider, err := testDNSComProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.DeleteRecord("goedge.cn", &dnstypes.Record{
		Id:    "535669356",
		Name:  "",
		Type:  "",
		Value: "",
		Route: "",
		TTL:   0,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func testDNSComProvider() (dnsclients.ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='dnscom' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("can not find providers with type 'dnscom'")
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &dnsclients.DNSComProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
