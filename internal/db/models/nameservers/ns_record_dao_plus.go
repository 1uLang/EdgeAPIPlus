package nameservers

import (
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"strconv"
	"strings"
)

const (
	NSRecordStateEnabled  = 1 // 已启用
	NSRecordStateDisabled = 0 // 已禁用
)

type NSRecordDAO dbs.DAO

func NewNSRecordDAO() *NSRecordDAO {
	return dbs.NewDAO(&NSRecordDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSRecords",
			Model:  new(NSRecord),
			PkName: "id",
		},
	}).(*NSRecordDAO)
}

var SharedNSRecordDAO *NSRecordDAO

func init() {
	dbs.OnReady(func() {
		SharedNSRecordDAO = NewNSRecordDAO()
	})
}

// EnableNSRecord 启用条目
func (this *NSRecordDAO) EnableNSRecord(tx *dbs.Tx, recordId int64) error {
	_, err := this.Query(tx).
		Pk(recordId).
		Set("state", NSRecordStateEnabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, recordId)
}

// DisableNSRecord 禁用条目
func (this *NSRecordDAO) DisableNSRecord(tx *dbs.Tx, recordId int64) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(recordId).
		Set("state", NSRecordStateDisabled).
		Set("version", version).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, recordId)
}

// FindEnabledNSRecord 查找启用中的条目
func (this *NSRecordDAO) FindEnabledNSRecord(tx *dbs.Tx, id int64) (*NSRecord, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSRecordStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSRecord), err
}

// FindNSRecordName 根据主键查找名称
func (this *NSRecordDAO) FindNSRecordName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateRecord 创建记录
func (this *NSRecordDAO) CreateRecord(tx *dbs.Tx, domainId int64, description string, name string, recordType dnsconfigs.RecordType, value string, ttl int32, routeIds []string) (int64, error) {
	// TTL默认值
	// TODO 根据用户级别设置不同的TTL
	if ttl <= 0 {
		ttl = 600
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return 0, err
	}

	// 处理各个类型的值
	recordType = strings.ToUpper(recordType)
	switch recordType {
	case dnsconfigs.RecordTypeCNAME:
		if !strings.HasSuffix(value, ".") {
			value += "."
		}
	}

	var op = NewNSRecordOperator()
	op.DomainId = domainId
	op.Description = description
	op.Name = name
	op.Type = recordType
	op.Value = value
	op.Ttl = ttl

	if len(routeIds) == 0 {
		op.RouteIds = `["default"]`
	} else {
		routeIds, err := json.Marshal(routeIds)
		if err != nil {
			return 0, err
		}
		op.RouteIds = routeIds
	}

	op.IsOn = true
	op.State = NSRecordStateEnabled
	op.Version = version
	recordId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = this.NotifyUpdate(tx, recordId)
	if err != nil {
		return 0, err
	}
	return recordId, nil
}

// UpdateRecord 修改记录
func (this *NSRecordDAO) UpdateRecord(tx *dbs.Tx, recordId int64, description string, name string, recordType dnsconfigs.RecordType, value string, ttl int32, routeIds []string, isOn bool) error {
	// TTL默认值
	// TODO 根据用户级别设置不同的TTL
	if ttl <= 0 {
		ttl = 600
	}

	if recordId <= 0 {
		return errors.New("invalid recordId")
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	// 处理各个类型的值
	recordType = strings.ToUpper(recordType)
	switch recordType {
	case dnsconfigs.RecordTypeCNAME:
		if !strings.HasSuffix(value, ".") {
			value += "."
		}
	}

	var op = NewNSRecordOperator()
	op.Id = recordId
	op.Description = description
	op.Name = name
	op.Type = recordType
	op.Value = value
	op.Ttl = ttl
	op.IsOn = isOn

	if len(routeIds) == 0 {
		op.RouteIds = `["default"]`
	} else {
		routeIds, err := json.Marshal(routeIds)
		if err != nil {
			return err
		}
		op.RouteIds = routeIds
	}

	op.Version = version

	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, recordId)
}

// CountAllEnabledDomainRecords 计算域名中记录数量
func (this *NSRecordDAO) CountAllEnabledDomainRecords(tx *dbs.Tx, domainId int64, dnsType dnsconfigs.RecordType, keyword string, routeCode string) (int64, error) {
	query := this.Query(tx).
		Attr("domainId", domainId).
		State(NSRecordStateEnabled)
	if len(dnsType) > 0 {
		query.Attr("type", dnsType)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR value LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(routeCode) > 0 {
		routeCodeJSON, err := json.Marshal(routeCode)
		if err != nil {
			return 0, err
		}
		query.JSONContains("routeIds", string(routeCodeJSON))
	}
	return query.Count()
}

// CountAllRecordsWithName 查询域名中相同记录名的数量
func (this *NSRecordDAO) CountAllRecordsWithName(tx *dbs.Tx, domainId int64, dnsType dnsconfigs.RecordType, name string) (int64, error) {
	return this.Query(tx).
		State(NSRecordStateEnabled).
		Attr("domainId", domainId).
		Attr("type", dnsType).
		Attr("name", name).
		Count()
}

// CountAllEnabledRecords 计算所有记录数量
func (this *NSRecordDAO) CountAllEnabledRecords(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		Where("domainId IN (SELECT id FROM " + SharedNSDomainDAO.Table + " WHERE state=1)").
		State(NSRecordStateEnabled).
		Count()
}

// ListEnabledRecords 列出单页记录
func (this *NSRecordDAO) ListEnabledRecords(tx *dbs.Tx,
	domainId int64,
	dnsType dnsconfigs.RecordType,
	keyword string,
	routeCode string,
	nameAsc bool,
	nameDesc bool,
	typeAsc bool,
	typeDesc bool,
	ttlAsc bool,
	ttlDesc bool,
	offset int64,
	size int64) (result []*NSRecord, err error) {
	var query = this.Query(tx).
		Attr("domainId", domainId).
		State(NSRecordStateEnabled)
	if len(dnsType) > 0 {
		query.Attr("type", dnsType)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR value LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(routeCode) > 0 {
		routeCodeJSON, err := json.Marshal(routeCode)
		if err != nil {
			return nil, err
		}
		query.JSONContains("routeIds", string(routeCodeJSON))
	}

	// 排序
	if nameAsc {
		query.Asc("name")
	} else if nameDesc {
		query.Desc("name")
	}

	if typeAsc {
		query.Asc("type")
	} else if typeDesc {
		query.Desc("type")
	}

	if ttlAsc {
		query.Asc("ttl")
	} else if ttlDesc {
		query.Desc("ttl")
	}

	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// IncreaseVersion 增加版本
func (this *NSRecordDAO) IncreaseVersion(tx *dbs.Tx) (int64, error) {
	return models.SharedSysLockerDAO.Increase(tx, "NS_RECORD_VERSION", 1)
}

// ListRecordsAfterVersion 列出某个版本后的记录
func (this *NSRecordDAO) ListRecordsAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*NSRecord, err error) {
	if size <= 0 {
		size = 10000
	}

	_, err = this.Query(tx).
		Gte("version", version).
		Limit(size).
		Asc("version").
		Slice(&result).
		FindAll()
	return
}

// FindEnabledRecordWithName 查询单条记录
func (this *NSRecordDAO) FindEnabledRecordWithName(tx *dbs.Tx, domainId int64, recordName string, recordType dnsconfigs.RecordType) (*NSRecord, error) {
	record, err := this.Query(tx).
		State(NSRecordStateEnabled).
		Attr("domainId", domainId).
		Attr("name", recordName).
		Attr("type", recordType).
		Find()
	if record == nil {
		return nil, err
	}
	return record.(*NSRecord), nil
}

// DisableRecordsInDomain 禁用某个域名中的所有记录
func (this *NSRecordDAO) DisableRecordsInDomain(tx *dbs.Tx, domainId int64) error {
	if domainId <= 0 {
		return nil
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Attr("domainId", domainId).
		Set("state", NSRecordStateDisabled).
		Set("version", version).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return SharedNSDomainDAO.NotifyUpdate(tx, domainId)
}

// DisableRecordsInDomainWithNameAndType 禁用某个域名中的某个名称的记录
func (this *NSRecordDAO) DisableRecordsInDomainWithNameAndType(tx *dbs.Tx, domainId int64, name string, recordType string) error {
	if domainId <= 0 {
		return nil
	}
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Attr("domainId", domainId).
		Attr("name", name).
		Attr("type", recordType).
		Set("version", version).
		Set("state", NSRecordStateDisabled).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return SharedNSDomainDAO.NotifyUpdate(tx, domainId)
}

// UpdateRecordsWithDomainId 批量修改记录
func (this *NSRecordDAO) UpdateRecordsWithDomainId(tx *dbs.Tx, domainId int64, searchName string, searchType string, searchValue string, searchRouteCodes []string, newName string, newType string, newValue string, newRouteCodes []string) error {
	searchType = strings.ToUpper(searchType)
	newType = strings.ToUpper(newType)

	var query = this.Query(tx)
	query.Attr("domainId", domainId)
	if len(searchName) > 0 {
		query.Attr("name", searchName)
	}
	if len(searchType) > 0 {
		query.Attr("type", searchType)
	}
	if len(searchValue) > 0 {
		if searchType == "CNAME" && !strings.HasSuffix(searchValue, ".") {
			searchValue += "."
		}
		query.Attr("value", searchValue)
	}
	if len(searchRouteCodes) > 0 {
		for _, routeCode := range searchRouteCodes {
			query.JSONContains("routeIds", strconv.Quote(routeCode))
		}
	}

	var shouldChange = false
	if len(newName) > 0 {
		query.Set("name", newName)
		shouldChange = true
	}
	if len(newType) > 0 {
		query.Set("type", newType)
		shouldChange = true
	}
	if len(newValue) > 0 {
		if newType == "CNAME" && !strings.HasSuffix(newValue, ".") {
			newValue += "."
		}
		query.Set("value", newValue)
		shouldChange = true
	}
	if len(newRouteCodes) > 0 {
		routeCodesJSON, err := json.Marshal(newRouteCodes)
		if err != nil {
			return err
		}
		query.Set("routeIds", routeCodesJSON)
		shouldChange = true
	}

	if !shouldChange {
		return nil
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}
	query.Set("version", version)

	err = query.UpdateQuickly()
	if err != nil {
		return err
	}

	return SharedNSDomainDAO.NotifyUpdate(tx, domainId)
}

// DisableRecordsWithDomainId 批量删除记录
func (this *NSRecordDAO) DisableRecordsWithDomainId(tx *dbs.Tx, domainId int64, searchName string, searchType string, searchValue string, searchRouteCodes []string) error {
	searchType = strings.ToUpper(searchType)

	var query = this.Query(tx)
	query.Attr("domainId", domainId)
	if len(searchName) > 0 {
		query.Attr("name", searchName)
	}
	if len(searchType) > 0 {
		query.Attr("type", searchType)
	}
	if len(searchValue) > 0 {
		if searchType == "CNAME" && !strings.HasSuffix(searchValue, ".") {
			searchValue += "."
		}
		query.Attr("value", searchValue)
	}
	if len(searchRouteCodes) > 0 {
		for _, routeCode := range searchRouteCodes {
			query.JSONContains("routeIds", strconv.Quote(routeCode))
		}
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	query.Set("version", version)
	query.Set("state", NSRecordStateDisabled)

	err = query.UpdateQuickly()
	if err != nil {
		return err
	}

	return SharedNSDomainDAO.NotifyUpdate(tx, domainId)
}

// UpdateRecordsIsOnWithDomainId 批量修改域名中的IsOn状态
func (this *NSRecordDAO) UpdateRecordsIsOnWithDomainId(tx *dbs.Tx, domainId int64, searchName string, searchType string, searchValue string, searchRouteCodes []string, isOn bool) error {
	searchType = strings.ToUpper(searchType)

	var query = this.Query(tx)
	query.Attr("domainId", domainId)
	if len(searchName) > 0 {
		query.Attr("name", searchName)
	}
	if len(searchType) > 0 {
		query.Attr("type", searchType)
	}
	if len(searchValue) > 0 {
		if searchType == "CNAME" && !strings.HasSuffix(searchValue, ".") {
			searchValue += "."
		}
		query.Attr("value", searchValue)
	}
	if len(searchRouteCodes) > 0 {
		for _, routeCode := range searchRouteCodes {
			query.JSONContains("routeIds", strconv.Quote(routeCode))
		}
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	query.Set("version", version)
	query.Set("isOn", isOn)

	err = query.UpdateQuickly()
	if err != nil {
		return err
	}

	return SharedNSDomainDAO.NotifyUpdate(tx, domainId)
}

// CheckUserRecord 检查用户记录
// 注意这里的userId可能为0
func (this *NSRecordDAO) CheckUserRecord(tx *dbs.Tx, userId int64, recordId int64) error {
	if recordId <= 0 {
		return nil
	}

	domainId, err := this.Query(tx).
		Result("domainId").
		Pk(recordId).
		FindInt64Col(0)
	if err != nil {
		return err
	}

	if domainId == 0 {
		return models.ErrNotFound
	}

	return SharedNSDomainDAO.CheckUserDomain(tx, userId, domainId)
}

// CountAllUserRecords 计算用户的记录数
func (this *NSRecordDAO) CountAllUserRecords(tx *dbs.Tx, userId int64) (int64, error) {
	if userId <= 0 {
		return 0, nil
	}

	return this.Query(tx).
		Where("domainId IN (SELECT id FROM "+SharedNSDomainDAO.Table+" WHERE state=1 AND userId=:userId)").
		Param("userId", userId).
		Count()
}

// NotifyUpdate 通知更新
func (this *NSRecordDAO) NotifyUpdate(tx *dbs.Tx, recordId int64) error {
	domainId, err := this.Query(tx).
		Pk(recordId).
		Result("domainId").
		FindInt64Col(0)
	if err != nil {
		return err
	}

	if domainId == 0 {
		return nil
	}

	clusterId, err := SharedNSDomainDAO.FindEnabledDomainClusterId(tx, domainId)
	if err != nil {
		return err
	}

	if clusterId > 0 {
		err = models.SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleDNS, clusterId, 0, models.NSNodeTaskTypeRecordChanged)
		if err != nil {
			return err
		}
	}
	return nil
}
