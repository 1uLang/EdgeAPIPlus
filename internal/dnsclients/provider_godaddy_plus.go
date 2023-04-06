// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/godaddy"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/maps"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	GoDaddyAPIEndpoint  = "https://api.godaddy.com/v1"
	GoDaddyDefaultRoute = "default"
	GoDaddyIdDelim      = "$"
	GoDaddyDefaultTTL   = 600
)

var goDaddyHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// GoDaddyProvider
//
// 参考文档：https://developer.godaddy.com/doc/endpoint/domains
type GoDaddyProvider struct {
	BaseProvider

	key    string
	secret string
}

// Auth 认证
func (this *GoDaddyProvider) Auth(params maps.Map) error {
	this.key = params.GetString("key")
	if len(this.key) == 0 {
		return errors.New("'key' should not be empty")
	}

	this.secret = params.GetString("secret")
	if len(this.secret) == 0 {
		return errors.New("'secret' should not be empty")
	}

	return nil
}

// GetDomains 获取所有域名列表
func (this *GoDaddyProvider) GetDomains() (domains []string, err error) {
	var respDomains = godaddy.DomainsResponse{}
	err = this.doAPI(http.MethodGet, "/domains", nil, &respDomains)
	if err != nil {
		return
	}

	for _, domain := range respDomains {
		if domain.Status == "ACTIVE" {
			domains = append(domains, domain.Domain)
		}
	}

	return
}

// GetRecords 获取域名解析记录列表
func (this *GoDaddyProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var respRecords = godaddy.RecordsResponse{}
	err = this.doAPI(http.MethodGet, "/domains/"+domain+"/records", nil, &respRecords)
	if err != nil {
		return
	}

	for _, record := range respRecords {
		// 修正Record
		if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Data, ".") {
			record.Data += "."
		}

		var recordObj = &dnstypes.Record{
			Name:  record.Name,
			Type:  record.Type,
			Value: record.Data,
			Route: GoDaddyDefaultRoute,
			TTL:   record.TTL,
		}
		this.addRecordId(recordObj)
		records = append(records, recordObj)
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *GoDaddyProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	routes = []*dnstypes.Route{
		{Name: "默认", Code: GoDaddyDefaultRoute},
	}
	return
}

// QueryRecord 查询单个记录
func (this *GoDaddyProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	var respRecords = godaddy.RecordsResponse{}
	err := this.doAPI(http.MethodGet, "/domains/"+domain+"/records/"+recordType+"/"+name, nil, &respRecords)
	if err != nil {
		return nil, err
	}

	for _, record := range respRecords {
		// 再次检查名称
		if record.Name != name {
			continue
		}

		// 修正Record
		if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Data, ".") {
			record.Data += "."
		}

		return &dnstypes.Record{
			Id:    record.Name + GoDaddyIdDelim + record.Type + GoDaddyIdDelim + stringutil.Md5(record.Data),
			Name:  record.Name,
			Type:  record.Type,
			Value: record.Data,
			Route: GoDaddyDefaultRoute,
			TTL:   record.TTL,
		}, nil
	}
	return nil, nil
}

// AddRecord 设置记录
func (this *GoDaddyProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	if newRecord.TTL <= 0 {
		newRecord.TTL = GoDaddyDefaultTTL
	}
	if newRecord.Type == dnstypes.RecordTypeCNAME {
		if !strings.HasSuffix(newRecord.Value, ".") {
			newRecord.Value += "."
		}
	}
	var recordMaps = []maps.Map{
		{
			"data":     newRecord.Value,
			"name":     newRecord.Name,
			"ttl":      newRecord.TTL,
			"type":     newRecord.Type,
			"priority": 0,
			"weight":   0,
			"port":     65535,
		},
	}
	recordMapsJSON, err := json.Marshal(recordMaps)
	if err != nil {
		return errors.New("encode records failed: " + err.Error())
	}

	err = this.doAPI(http.MethodPatch, "/domains/"+domain+"/records", recordMapsJSON, nil)
	if err != nil {
		return err
	}

	this.addRecordId(newRecord)
	return nil
}

// UpdateRecord 修改记录
func (this *GoDaddyProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	var recordType = record.Type
	var recordName = record.Name
	var recordValueMd5 = stringutil.Md5(record.Value)

	if len(recordType) == 0 || len(recordName) == 0 {
		if len(record.Id) == 0 {
			return errors.New("invalid record to delete")
		}
		recordName, recordType, recordValueMd5 = this.splitRecordId(record.Id)
		if len(recordType) == 0 || len(recordName) == 0 {
			return errors.New("invalid record to delete")
		}
	}

	var respRecords = godaddy.RecordsResponse{}
	err := this.doAPI(http.MethodGet, "/domains/"+domain+"/records/"+recordType+"/"+recordName, nil, &respRecords)
	if err != nil {
		return err
	}

	this.addRecordId(newRecord)

	var found = false
	for index, gRecord := range respRecords {
		var gRecordValue = gRecord.Data
		if gRecord.Type == dnstypes.RecordTypeCNAME {
			if !strings.HasSuffix(gRecordValue, ".") {
				gRecordValue += "."
			}
		}
		if gRecord.Name == recordName && gRecord.Type == recordType && stringutil.Md5(gRecordValue) == recordValueMd5 {
			gRecord.Name = newRecord.Name

			if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
				newRecord.Value += "."
			}
			gRecord.Data = newRecord.Value

			gRecord.Type = newRecord.Type
			gRecord.TTL = newRecord.TTL

			if newRecord.TTL <= 0 {
				gRecord.TTL = GoDaddyDefaultTTL
			}

			respRecords[index] = gRecord

			found = true
			break
		}
	}

	if found {
		newRecordsJSON, err := json.Marshal(respRecords)
		if err != nil {
			return err
		}
		err = this.doAPI(http.MethodPut, "/domains/"+domain+"/records/"+recordType+"/"+recordName, newRecordsJSON, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteRecord 删除记录
func (this *GoDaddyProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	var recordType = record.Type
	var recordName = record.Name
	var recordValueMd5 = stringutil.Md5(record.Value)

	if len(recordType) == 0 || len(recordName) == 0 {
		if len(record.Id) == 0 {
			return errors.New("invalid record to delete")
		}
		recordName, recordType, recordValueMd5 = this.splitRecordId(record.Id)
		if len(recordType) == 0 || len(recordName) == 0 {
			return errors.New("invalid record to delete")
		}
	}

	var respRecords = godaddy.RecordsResponse{}
	err := this.doAPI(http.MethodGet, "/domains/"+domain+"/records/"+recordType+"/"+recordName, nil, &respRecords)
	if err != nil {
		return err
	}

	var newRecords = godaddy.RecordsResponse{}
	for _, gRecord := range respRecords {
		var gRecordValue = gRecord.Data
		if gRecord.Type == dnstypes.RecordTypeCNAME {
			if !strings.HasSuffix(gRecordValue, ".") {
				gRecordValue += "."
			}
		}
		if gRecord.Name == recordName && gRecord.Type == recordType && stringutil.Md5(gRecordValue) == recordValueMd5 {
			continue
		}
		newRecords = append(newRecords, gRecord)
	}

	if len(newRecords) > 0 {
		newRecordsJSON, err := json.Marshal(newRecords)
		if err != nil {
			return err
		}
		err = this.doAPI(http.MethodPut, "/domains/"+domain+"/records/"+recordType+"/"+recordName, newRecordsJSON, nil)
		if err != nil {
			return err
		}
	} else {
		err = this.doAPI(http.MethodDelete, "/domains/"+domain+"/records/"+recordType+"/"+recordName, nil, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// DefaultRoute 默认线路
func (this *GoDaddyProvider) DefaultRoute() string {
	return GoDaddyDefaultRoute
}

func (this *GoDaddyProvider) addRecordId(record *dnstypes.Record) {
	record.Id = record.Name + GoDaddyIdDelim + record.Type + GoDaddyIdDelim + stringutil.Md5(record.Value)
}

func (this *GoDaddyProvider) splitRecordId(recordId string) (recordName string, recordType string, valueMd5 string) {
	var pieces = strings.Split(recordId, GoDaddyIdDelim)
	if len(pieces) < 3 {
		return
	}
	return pieces[0], pieces[1], pieces[2]
}

// 发送请求
func (this *GoDaddyProvider) doAPI(method string, apiPath string, bodyJSON []byte, respPtr interface{}) error {
	apiURL := GoDaddyAPIEndpoint + apiPath
	method = strings.ToUpper(method)

	var bodyReader io.Reader = nil
	if len(bodyJSON) > 0 {
		bodyReader = bytes.NewReader(bodyJSON)
	}

	req, err := http.NewRequest(method, apiURL, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "sso-key "+this.key+":"+this.secret)
	resp, err := goDaddyHTTPClient.Do(req)
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
			return errors.New("decode json failed: " + err.Error() + ", response text: " + string(data))
		}
	}

	return nil
}
