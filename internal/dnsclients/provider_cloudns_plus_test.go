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

func TestClouDNSProvider_GetDomains(t *testing.T) {
	provider, err := testClouDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(domains)
}

func TestClouDNSProvider_GetRecords(t *testing.T) {
	provider, err := testClouDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	records, err := provider.GetRecords("goedge.org")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestClouDNSProvider_GetRoutes(t *testing.T) {
	provider, err := testClouDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	routes, err := provider.GetRoutes("goedge.org")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestClouDNSProvider_QueryRecord(t *testing.T) {
	provider, err := testClouDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	for _, recordName := range []string{"www", "test", "@", ""} {
		t.Log("===", recordName, "===")
		record, err := provider.QueryRecord("goedge.org", recordName, dnstypes.RecordTypeA)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
}

func TestClouDNSProvider_AddRecord(t *testing.T) {
	provider, err := testClouDNSProvider()
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
			TTL:   7200,
		}
		err := provider.AddRecord("goedge.org", record)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		var record = &dnstypes.Record{
			Id:    "",
			Name:  "test2",
			Type:  dnstypes.RecordTypeCNAME,
			Value: "goedge.org.",
			Route: "",
			TTL:   0,
		}
		err := provider.AddRecord("goedge.org", record)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestClouDNSProvider_UpdateRecord(t *testing.T) {
	provider, err := testClouDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	var record = &dnstypes.Record{
		Id: "262076455",
	}
	var newRecord = &dnstypes.Record{
		Id:    "",
		Name:  "test1",
		Type:  dnstypes.RecordTypeA,
		Value: "192.168.1.101",
		Route: "",
		TTL:   3600,
	}
	err = provider.UpdateRecord("goedge.org", record, newRecord)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClouDNSProvider_DeleteRecord(t *testing.T) {
	provider, err := testClouDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.DeleteRecord("goedge.org", &dnstypes.Record{
		Id:    "262075770",
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

func testClouDNSProvider() (dnsclients.ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='cloudns' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("can not find providers with type 'cloudns'")
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &dnsclients.ClouDNSProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
