// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package dnsclients_test

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"testing"
)

func TestGoDaddyProvider_GetDomains(t *testing.T) {
	provider, err := testGoDaddyProvider()
	if err != nil {
		t.Fatal(err)
	}

	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(domains)
}

func TestGoDaddyProvider_GetRecords(t *testing.T) {
	provider, err := testGoDaddyProvider()
	if err != nil {
		t.Fatal(err)
	}

	records, err := provider.GetRecords("goedge.cloud")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestGoDaddyProvider_GetRoutes(t *testing.T) {
	provider, err := testGoDaddyProvider()
	if err != nil {
		t.Fatal(err)
	}

	routes, err := provider.GetRoutes("goedge.cloud")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestGoDaddyProvider_QueryRecord(t *testing.T) {
	provider, err := testGoDaddyProvider()
	if err != nil {
		t.Fatal(err)
	}

	for _, recordName := range []string{"www", "test", "@", ""} {
		t.Log("===", recordName, "===")
		record, err := provider.QueryRecord("goedge.cloud", recordName, dnstypes.RecordTypeA)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
}

func TestGoDaddyProvider_AddRecord(t *testing.T) {
	provider, err := testGoDaddyProvider()
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
		err := provider.AddRecord("goedge.cloud", record)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		var record = &dnstypes.Record{
			Id:    "",
			Name:  "test2",
			Type:  dnstypes.RecordTypeCNAME,
			Value: "goedge.cn.",
			Route: "",
			TTL:   0,
		}
		err := provider.AddRecord("goedge.cloud", record)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGoDaddyProvider_UpdateRecord(t *testing.T) {
	provider, err := testGoDaddyProvider()
	if err != nil {
		t.Fatal(err)
	}

	var record = &dnstypes.Record{
		Id: "test1" + dnsclients.GoDaddyIdDelim + "A" + dnsclients.GoDaddyIdDelim + stringutil.Md5("192.168.1.101"),
	}
	var newRecord = &dnstypes.Record{
		Id:    "",
		Name:  "test1",
		Type:  dnstypes.RecordTypeA,
		Value: "192.168.1.101",
		Route: "",
		TTL:   3600,
	}
	err = provider.UpdateRecord("goedge.cloud", record, newRecord)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoDaddyProvider_DeleteRecord(t *testing.T) {
	provider, err := testGoDaddyProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.DeleteRecord("goedge.cloud", &dnstypes.Record{
		Id:    "test" + dnsclients.GoDaddyIdDelim + "A" + dnsclients.GoDaddyIdDelim + stringutil.Md5("192.168.1.101"),
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

func testGoDaddyProvider() (dnsclients.ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='godaddy' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("can not find providers with type 'godaddy'")
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &dnsclients.GoDaddyProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
