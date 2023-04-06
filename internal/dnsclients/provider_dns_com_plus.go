// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnscom"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DNSComAPIEndpoint = "https://www.dns.com"
)

var goDNSComHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// DNSComProvider DNS.COM域名服务
// 参考文档：https://www.dns.com/document/api/4/81.html
type DNSComProvider struct {
	key    string
	secret string

	domainMap map[string]string // domainName => id
	locker    sync.Mutex
}

// Auth 认证
func (this *DNSComProvider) Auth(params maps.Map) error {
	this.domainMap = map[string]string{}

	this.key = params.GetString("key")
	if len(this.key) == 0 {
		return errors.New("require 'key' parameter")
	}

	this.secret = params.GetString("secret")
	if len(this.secret) == 0 {
		return errors.New("require 'secret' parameter")
	}
	return nil
}

// GetDomains 获取所有域名列表
func (this *DNSComProvider) GetDomains() (domains []string, err error) {
	var pageSize = 100
	var pageCount = 0

	var queryPage = func(page int) error {
		var resp = &dnscom.DomainListResponse{}
		err := this.doAPI(http.MethodGet, "/api/domain/list/", map[string]string{
			"page":     types.String(page),
			"pageSize": types.String(pageSize),
		}, &resp)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return this.composeError(resp.Code, resp.Message)
		}
		if page == 1 {
			pageCount = resp.Data.PageCount
		}
		for _, d := range resp.Data.Data {
			domains = append(domains, d.Domains)
		}
		return nil
	}

	err = queryPage(1)
	if err != nil {
		return nil, err
	}

	// 其他页
	if pageCount > 1 {
		for page := 2; page <= pageCount; page++ {
			err = queryPage(page)
			if err != nil {
				return nil, err
			}
		}
	}

	return
}

// GetRecords 获取域名解析记录列表
func (this *DNSComProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	// 获取域名ID
	domainId, err := this.queryDomainId(domain)
	if err != nil {
		return nil, err
	}
	if len(domainId) == 0 {
		return nil, errors.New("can not find domain '" + domain + "'")
	}

	// 列出记录
	var pageSize = 100
	var pageCount = 0
	var queryPage = func(page int) error {
		var resp = &dnscom.RecordListResponse{}
		err := this.doAPI(http.MethodGet, "/api/record/list/", map[string]string{
			"domainID": domainId,
			"page":     types.String(page),
			"pageSize": types.String(pageSize),
		}, &resp)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return this.composeError(resp.Code, resp.Message)
		}
		if page == 1 {
			pageCount = resp.Data.PageCount
		}

		for _, record := range resp.Data.Data {
			// 修正Record
			if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Value, ".") {
				record.Value += "."
			}

			records = append(records, &dnstypes.Record{
				Id:    types.String(record.RecordID),
				Name:  record.Record,
				Type:  record.Type,
				Value: record.Value,
				Route: types.String(record.ViewID),
				TTL:   types.Int32(record.TTL),
			})
		}

		return nil
	}

	err = queryPage(1)
	if err != nil {
		return nil, err
	}

	if pageCount > 1 {
		for page := 2; page <= pageCount; page++ {
			err = queryPage(page)
			if err != nil {
				return nil, err
			}
		}
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *DNSComProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	_ = domain

	// 区域
	{
		var resp = &dnscom.IPAreaViewListResponse{}
		err = this.doAPI(http.MethodGet, "/api/ip/areaviewlist/", map[string]string{}, resp)
		if err != nil {
			return
		}
		if resp.Code != 0 {
			return nil, this.composeError(resp.Code, resp.Message)
		}
		for _, route := range resp.Data {
			routes = append(routes, &dnstypes.Route{
				Name: "[地区]" + route.Name,
				Code: types.String(route.ViewID),
			})
		}
	}

	// ISP
	{
		var resp = &dnscom.IPISPViewListResponse{}
		err = this.doAPI(http.MethodGet, "/api/ip/ispviewlist/", map[string]string{}, resp)
		if err != nil {
			return
		}
		if resp.Code != 0 {
			return nil, this.composeError(resp.Code, resp.Message)
		}
		for _, route := range resp.Data {
			routes = append(routes, &dnstypes.Route{
				Name: "[ISP]" + route.Name,
				Code: types.String(route.ViewID),
			})
		}
	}

	return
}

// QueryRecord 查询单个记录
func (this *DNSComProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	// 获取域名ID
	domainId, err := this.queryDomainId(domain)
	if err != nil {
		return nil, err
	}
	if len(domainId) == 0 {
		return nil, errors.New("can not find domain '" + domain + "'")
	}

	// 列出记录
	var pageSize = 100
	var pageCount = 0
	var recordResult *dnstypes.Record
	var queryPage = func(page int) error {
		var resp = &dnscom.RecordListResponse{}
		err := this.doAPI(http.MethodGet, "/api/record/list/", map[string]string{
			"domainID": domainId,
			"host":     name,
			"page":     types.String(page),
			"pageSize": types.String(pageSize),
		}, &resp)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return this.composeError(resp.Code, resp.Message)
		}
		if page == 1 {
			pageCount = resp.Data.PageCount
		}

		for _, record := range resp.Data.Data {
			// 仍然比对name，因为搜索条件为空时，API仍然返回了全部的记录
			if record.Record == name && record.Type == recordType {
				// 修正Record
				if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Value, ".") {
					record.Value += "."
				}

				recordResult = &dnstypes.Record{
					Id:    types.String(record.RecordID),
					Name:  record.Record,
					Type:  record.Type,
					Value: record.Value,
					Route: types.String(record.ViewID),
					TTL:   types.Int32(record.TTL),
				}
				break
			}
		}

		return nil
	}

	err = queryPage(1)
	if err != nil {
		return nil, err
	}
	if recordResult != nil {
		return recordResult, nil
	}

	if pageCount > 1 {
		for page := 2; page <= pageCount; page++ {
			err = queryPage(page)
			if err != nil {
				return nil, err
			}
			if recordResult != nil {
				return recordResult, nil
			}
		}
	}

	return nil, nil
}

// AddRecord 设置记录
func (this *DNSComProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	// 查找域名ID
	domainId, err := this.queryDomainId(domain)
	if err != nil {
		return err
	}
	if len(domainId) == 0 {
		return errors.New("can not find domain '" + domain + "'")
	}

	// 创建记录
	var resp = &dnscom.CreateRecordResponse{}
	var viewId = "0"
	if len(newRecord.Route) > 0 {
		viewId = newRecord.Route
	}
	err = this.doAPI(http.MethodGet, "/api/record/create/", map[string]string{
		"domainID": domainId,
		"type":     newRecord.Type,
		"viewID":   viewId,
		"host":     newRecord.Name,
		"value":    newRecord.Value,
		"TTL":      types.String(newRecord.TTL),
	}, resp)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return this.composeError(resp.Code, resp.Message)
	}

	return nil
}

// UpdateRecord 修改记录
func (this *DNSComProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	domainId, err := this.queryDomainId(domain)
	if err != nil {
		return err
	}
	if len(domainId) == 0 {
		return errors.New("can not find domain '" + domain + "'")
	}

	var resp = &dnscom.RecordModifyResponse{}
	var newViewId = "0"
	if len(newRecord.Route) > 0 {
		newViewId = newRecord.Route
	}
	err = this.doAPI(http.MethodGet, "/api/record/modify/", map[string]string{
		"domainID":  domainId,
		"recordID":  record.Id,
		"newhost":   newRecord.Name,
		"newtype":   newRecord.Type,
		"newvalue":  newRecord.Value,
		"newttl":    types.String(newRecord.TTL),
		"newviewID": newViewId,
	}, resp)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return this.composeError(resp.Code, resp.Message)
	}

	return nil
}

// DeleteRecord 删除记录
func (this *DNSComProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	domainId, err := this.queryDomainId(domain)
	if err != nil {
		return err
	}
	if len(domainId) == 0 {
		return errors.New("can not find domain '" + domain + "'")
	}

	var resp = &dnscom.RecordRemoveResponse{}
	err = this.doAPI(http.MethodGet, "/api/record/remove", map[string]string{
		"domainID": domainId,
		"recordID": record.Id,
	}, resp)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return this.composeError(resp.Code, resp.Message)
	}

	return nil
}

// DefaultRoute 默认线路
func (this *DNSComProvider) DefaultRoute() string {
	return "0"
}

// 查找域名ID
func (this *DNSComProvider) queryDomainId(domain string) (string, error) {
	this.locker.Lock()
	domainId, ok := this.domainMap[domain]
	if ok {
		this.locker.Unlock()
		return domainId, nil
	}
	this.locker.Unlock()

	var pageSize = 100
	var pageCount = 0

	var queryPage = func(page int) error {
		var resp = &dnscom.DomainSearchResponse{}
		err := this.doAPI(http.MethodGet, "/api/domain/search/", map[string]string{
			"query":    domain,
			"page":     types.String(page),
			"pageSize": types.String(pageSize),
		}, &resp)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return this.composeError(resp.Code, resp.Message)
		}
		if page == 1 {
			pageCount = resp.Data.PageCount
		}
		for _, d := range resp.Data.Data {
			if d.Domains == domain {
				domainId = d.DomainsID
				return nil
			}
		}
		return nil
	}

	err := queryPage(1)
	if err != nil {
		return "", err
	}
	if len(domainId) > 0 {
		this.locker.Lock()
		this.domainMap[domain] = domainId
		this.locker.Unlock()
		return domainId, nil
	}

	// 其他页
	if pageCount > 1 {
		for page := 2; page <= pageCount; page++ {
			err = queryPage(page)
			if err != nil {
				return "", err
			}
			if len(domainId) > 0 {
				this.locker.Lock()
				this.domainMap[domain] = domainId
				this.locker.Unlock()
				return domainId, nil
			}
		}
	}

	return "", nil
}

// 发送请求
func (this *DNSComProvider) doAPI(method string, apiPath string, params map[string]string, respPtr interface{}) error {
	var apiURL = DNSComAPIEndpoint + apiPath
	method = strings.ToUpper(method)

	params["apiKey"] = this.key
	params["timestamp"] = types.String(time.Now().Unix())
	params["hash"] = this.hashParams(params)

	var query = url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	var reader io.Reader
	if method == http.MethodPost {
		reader = strings.NewReader(query.Encode())
	} else {
		apiURL += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, apiURL, reader)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := goDNSComHTTPClient.Do(req)
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

// 构造错误提示
func (this *DNSComProvider) composeError(code int, message string) error {
	return errors.New("error code:" + types.String(code) + ", message:" + message)
}

// 计算参数Hsh值
func (this *DNSComProvider) hashParams(params map[string]string) string {
	var keys = []string{}
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var source string
	for _, key := range keys {
		if source == "" {
			source += key + "=" + params[key]
		} else {
			source += "&" + key + "=" + params[key]
		}
	}

	var md = md5.Sum([]byte(source + this.secret))
	return hex.EncodeToString(md[:])
}
