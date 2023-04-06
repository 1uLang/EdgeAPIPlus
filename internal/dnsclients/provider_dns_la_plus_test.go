// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package dnsclients_test

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestDNSLaProvider_GetDomains(t *testing.T) {
	provider, err := testDNSLaProvider()
	if err != nil {
		t.Fatal(err)
	}

	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(domains)
}

func TestDNSLAProvider_GetRecords(t *testing.T) {
	provider, err := testDNSLaProvider()
	if err != nil {
		t.Fatal(err)
	}

	records, err := provider.GetRecords("hello2.com")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestDNSLaProvider_GetRoutes(t *testing.T) {
	provider, err := testDNSLaProvider()
	if err != nil {
		t.Fatal(err)
	}

	routes, err := provider.GetRoutes("hello2.com")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestDNSLaProvider_QueryRecord(t *testing.T) {
	provider, err := testDNSLaProvider()
	if err != nil {
		t.Fatal(err)
	}

	for _, recordName := range []string{"www", "test", "@", ""} {
		t.Log("===", recordName, "===")
		record, err := provider.QueryRecord("hello2.com", recordName, dnstypes.RecordTypeA)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
}

func TestDNSLaProvider_AddRecord(t *testing.T) {
	provider, err := testDNSLaProvider()
	if err != nil {
		t.Fatal(err)
	}

	{
		var record = &dnstypes.Record{
			Id:    "",
			Name:  "test1",
			Type:  dnstypes.RecordTypeA,
			Value: "192.168.1.100",
			Route: "",
			TTL:   600,
		}
		err := provider.AddRecord("hello2.com", record)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("id:", record.Id)
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

func TestDNSLaProvider_UpdateRecord(t *testing.T) {
	provider, err := testDNSLaProvider()
	if err != nil {
		t.Fatal(err)
	}

	var record = &dnstypes.Record{
		Id: "20327509",
	}
	var newRecord = &dnstypes.Record{
		Id:    "",
		Name:  "test1",
		Type:  dnstypes.RecordTypeA,
		Value: "192.168.1.101",
		Route: "unic",
		TTL:   3600,
	}
	err = provider.UpdateRecord("hello2.com", record, newRecord)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDNSLaProvider_DeleteRecord(t *testing.T) {
	provider, err := testDNSLaProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.DeleteRecord("hello2.com", &dnstypes.Record{
		Id:    "20327472",
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

func testDNSLaProvider() (dnsclients.ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='dnsla' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("can not find providers with type 'dnsla'")
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &dnsclients.DNSLaProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
