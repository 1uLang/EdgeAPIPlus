// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients

import (
	"crypto/tls"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnsla"
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

const DNSLaAPIEndpoint = "https://api.dns.la"

var dnsLAHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type DNSLaProvider struct {
	BaseProvider

	apiId  string
	secret string
}

// Auth 认证
func (this *DNSLaProvider) Auth(params maps.Map) error {
	this.apiId = params.GetString("apiId")
	this.secret = params.GetString("secret")

	if len(this.apiId) == 0 {
		return errors.New("'apiId' should not be empty")
	}
	if len(this.secret) == 0 {
		return errors.New("'secret' should not be empty")
	}

	return nil
}

// GetDomains 获取所有域名列表
func (this *DNSLaProvider) GetDomains() (domains []string, err error) {
	for i := 1; i < 1000; i++ {
		var resp = &dnsla.DomainListResponse{}
		err = this.doAPI("/api/domain.ashx", map[string]string{
			"cmd":      "list",
			"pagesize": "100",
			"pageno":   types.String(i),
		}, resp)
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			return nil, resp.Error()
		}

		if len(resp.Datas) == 0 {
			return
		}

		for _, data := range resp.Datas {
			domains = append(domains, data.DomainName)
		}
	}
	return
}

// GetRecords 获取域名解析记录列表
func (this *DNSLaProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var resp = &dnsla.RecordListResponse{}
	err = this.doAPI("/api/record.ashx", map[string]string{
		"cmd":    "list",
		"domain": domain,
	}, resp)
	if err != nil {
		return
	}
	if !resp.Success() {
		return nil, resp.Error()
	}
	for _, data := range resp.Datas {
		// 修正Record
		if data.RecordType == dnstypes.RecordTypeCNAME && !strings.HasSuffix(data.RecordData, ".") {
			data.RecordData += "."
		}

		records = append(records, &dnstypes.Record{
			Id:    types.String(data.RecordId),
			Name:  data.Host,
			Type:  data.RecordType,
			Value: data.RecordData,
			Route: data.RecordLine,
			TTL:   types.Int32(data.TTL),
		})
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *DNSLaProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	var resp = &dnsla.RecordLineResponse{}
	err = this.doAPI("/api/dict.ashx", map[string]string{
		"cmd": "record_line",
	}, resp)
	if err != nil {
		return
	}
	if !resp.Success() {
		return nil, resp.Error()
	}

	for _, data := range resp.Datas {
		routes = append(routes, &dnstypes.Route{
			Name: data.Text,
			Code: data.Value,
		})
	}

	return
}

// QueryRecord 查询单个记录
func (this *DNSLaProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	records, err := this.GetRecords(domain)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.Name == name && record.Type == recordType {
			return record, nil
		}
	}
	return nil, nil
}

// AddRecord 设置记录
func (this *DNSLaProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	var resp = &dnsla.RecordCreateResponse{}

	var route = newRecord.Route
	if route == "default" {
		route = ""
	}

	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 600
	}

	if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
		newRecord.Value += "."
	}

	err := this.doAPI("/api/record.ashx", map[string]string{
		"cmd":        "create",
		"domain":     domain,
		"host":       newRecord.Name,
		"recordtype": newRecord.Type,
		"recorddata": newRecord.Value,
		"recordline": route,
		"ttl":        types.String(ttl),
	}, resp)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return resp.Error()
	}
	newRecord.Id = types.String(resp.ResultId)

	return nil
}

// UpdateRecord 修改记录
func (this *DNSLaProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	var resp = &dnsla.RecordCreateResponse{}

	var route = newRecord.Route
	if route == "default" {
		route = ""
	}

	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 600
	}

	if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
		newRecord.Value += "."
	}

	err := this.doAPI("/api/record.ashx", map[string]string{
		"cmd":        "edit",
		"domain":     domain,
		"recordid":   record.Id,
		"host":       newRecord.Name,
		"recordtype": newRecord.Type,
		"recorddata": newRecord.Value,
		"recordline": route,
		"ttl":        types.String(ttl),
	}, resp)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return resp.Error()
	}

	return nil
}

// DeleteRecord 删除记录
func (this *DNSLaProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	var resp = &dnsla.RecordRemoveResponse{}
	err := this.doAPI("/api/record.ashx", map[string]string{
		"cmd":      "remove",
		"domain":   domain,
		"recordid": record.Id,
	}, resp)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return resp.Error()
	}
	return nil
}

// DefaultRoute 默认线路
func (this *DNSLaProvider) DefaultRoute() string {
	return "default"
}

// 发送请求
func (this *DNSLaProvider) doAPI(path string, params map[string]string, respPtr interface{}) error {
	var apiURL = DNSLaAPIEndpoint + path
	var method = http.MethodPost
	var query = &url.Values{}
	query.Set("apiid", this.apiId)
	query.Set("apipass", this.secret)
	query.Set("rtype", "json")

	for k, v := range params {
		query.Set(k, v)
	}

	var reader = strings.NewReader(query.Encode())

	req, err := http.NewRequest(method, apiURL, reader)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := dnsLAHTTPClient.Do(req)
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
