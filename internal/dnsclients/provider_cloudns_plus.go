// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients

import (
	"crypto/tls"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/cloudns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ClouDNSDefaultRoute = "default"
	ClouDNSAPIEndpoint  = "https://api.cloudns.net"
)

var clouDNSHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// ClouDNSProvider ClouDNS.net
// 参考文档：https://www.cloudns.net/wiki/article/41/
type ClouDNSProvider struct {
	authId       int64
	subAuthId    int64
	authPassword string
}

// Auth 认证
func (this *ClouDNSProvider) Auth(params maps.Map) error {
	this.authId = params.GetInt64("authId")
	this.subAuthId = params.GetInt64("subAuthId")
	this.authPassword = params.GetString("authPassword")
	return nil
}

// GetDomains 获取所有域名列表
func (this *ClouDNSProvider) GetDomains() (domains []string, err error) {
	var page = 1

	for {
		var zones = cloudns.ZonesResponse{}
		err = this.doAPI(http.MethodPost, "/dns/list-zones.json", map[string]string{
			"page":          types.String(page),
			"rows-per-page": "100",
		}, &zones)
		if err != nil {
			return
		}

		if len(zones) == 0 {
			break
		}

		for _, zone := range zones {
			if zone.Zone == "domain" {
				domains = append(domains, zone.Name)
			}
		}

		page++
	}

	return
}

// GetRecords 获取域名解析记录列表
func (this *ClouDNSProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var page = 1
	for {
		var respRecords = cloudns.RecordsResponse{}
		err = this.doAPI(http.MethodPost, "/dns/records.json", map[string]string{
			"domain-name":   domain,
			"page":          types.String(page),
			"rows-per-page": "100",
		}, &respRecords)
		if err != nil {
			return
		}

		if len(respRecords) == 0 {
			break
		}

		for _, record := range respRecords {
			// 修正Record
			if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Record, ".") {
				record.Record += "."
			}

			records = append(records, &dnstypes.Record{
				Id:    record.Id,
				Name:  record.Host,
				Type:  record.Type,
				Value: record.Record,
				Route: ClouDNSDefaultRoute,
				TTL:   types.Int32(record.TTL),
			})
		}

		page++
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *ClouDNSProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	routes = []*dnstypes.Route{
		{Name: "默认", Code: ClouDNSDefaultRoute},
	}

	// TODO 支持GeoDNS

	return
}

// QueryRecord 查询单个记录
func (this *ClouDNSProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	var respRecords = cloudns.RecordsResponse{}
	err := this.doAPI(http.MethodPost, "/dns/records.json", map[string]string{
		"domain-name": domain,
		"host":        name,
		"type":        recordType,
	}, &respRecords)
	if err != nil {
		return nil, err
	}

	if len(respRecords) == 0 {
		return nil, nil
	}

	for _, record := range respRecords {
		// 修正Record
		if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Record, ".") {
			record.Record += "."
		}

		return &dnstypes.Record{
			Id:    record.Id,
			Name:  record.Host,
			Type:  record.Type,
			Value: record.Record,
			Route: ClouDNSDefaultRoute,
			TTL:   types.Int32(record.TTL),
		}, nil
	}

	return nil, nil
}

// AddRecord 设置记录
func (this *ClouDNSProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 1800
	}

	var availableTTLs = []int32{60, 300, 900, 1800, 3600, 21600, 43200, 86400, 172800, 259200, 604800, 1209600, 2592000}
	var ttlFound = false
	for _, aTTL := range availableTTLs {
		if aTTL == ttl {
			ttlFound = true
			break
		}
	}
	if !ttlFound {
		ttl = 1800
	}

	var statusResp = &cloudns.StatusResponse{}
	err := this.doAPI(http.MethodPost, "/dns/add-record.json", map[string]string{
		"domain-name": domain,
		"record-type": newRecord.Type,
		"host":        newRecord.Name,
		"record":      newRecord.Value,
		"ttl":         types.String(ttl),
	}, statusResp)
	if err != nil {
		return err
	}

	if statusResp.Status != "Success" {
		return errors.New("Failed: " + statusResp.StatusDescription)
	}

	return nil
}

// UpdateRecord 修改记录
func (this *ClouDNSProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 1800
	}

	var availableTTLs = []int32{60, 300, 900, 1800, 3600, 21600, 43200, 86400, 172800, 259200, 604800, 1209600, 2592000}
	var ttlFound = false
	for _, aTTL := range availableTTLs {
		if aTTL == ttl {
			ttlFound = true
			break
		}
	}
	if !ttlFound {
		ttl = 1800
	}

	var statusResp = &cloudns.StatusResponse{}
	err := this.doAPI(http.MethodPost, "/dns/mod-record.json", map[string]string{
		"domain-name": domain,
		"record-id":   record.Id,
		"record-type": newRecord.Type,
		"host":        newRecord.Name,
		"record":      newRecord.Value,
		"ttl":         types.String(ttl),
	}, statusResp)
	if err != nil {
		return err
	}

	if statusResp.Status != "Success" {
		return errors.New("Failed: " + statusResp.StatusDescription)
	}

	return nil
}

// DeleteRecord 删除记录
func (this *ClouDNSProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	var statusResp = &cloudns.StatusResponse{}
	err := this.doAPI(http.MethodPost, "/dns/delete-record.json", map[string]string{
		"domain-name": domain,
		"record-id":   record.Id,
	}, statusResp)
	if err != nil {
		return err
	}

	if statusResp.Status != "Success" {
		return errors.New("Failed: " + statusResp.StatusDescription)
	}
	return nil
}

// DefaultRoute 默认线路
func (this *ClouDNSProvider) DefaultRoute() string {
	return ClouDNSDefaultRoute
}

// 发送请求
func (this *ClouDNSProvider) doAPI(method string, apiPath string, params map[string]string, respPtr interface{}) error {
	var apiURL = ClouDNSAPIEndpoint + apiPath
	method = strings.ToUpper(method)

	var query = url.Values{}
	if this.authId > 0 {
		query.Set("auth-id", types.String(this.authId))
	} else if this.subAuthId > 0 {
		query.Set("sub-auth-id", types.String(this.subAuthId))
	}
	query.Set("auth-password", this.authPassword)

	for k, v := range params {
		query.Set(k, v)
	}

	req, err := http.NewRequest(method, apiURL, strings.NewReader(query.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := clouDNSHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 0 {
		return errors.New("invalid response status '" + strconv.Itoa(resp.StatusCode) + "', response '" + string(data) + "'")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("response error: " + string(data))
	}

	if respPtr != nil {
		err = json.Unmarshal(data, respPtr)
		if err != nil {
			return errors.New("decode json failed: " + err.Error() + ": " + string(data))
		}
	}

	return nil
}
