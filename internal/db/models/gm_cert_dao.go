package models

import (
	"encoding/json"
	"errors"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/gmconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"strings"
	"time"
)

const (
	GMCertStateEnabled  = 1 // 已启用
	GMCertStateDisabled = 0 // 已禁用
)

type GMCertDAO dbs.DAO

func NewGMCertDAO() *GMCertDAO {
	return dbs.NewDAO(&GMCertDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeGMCerts",
			Model:  new(GMCert),
			PkName: "id",
		},
	}).(*GMCertDAO)
}

var SharedGMCertDAO *GMCertDAO

func init() {
	dbs.OnReady(func() {
		SharedGMCertDAO = NewGMCertDAO()
	})
}

// Init 初始化
func (this *GMCertDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableGMCert 启用条目
func (this *GMCertDAO) EnableGMCert(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", GMCertStateEnabled).
		Update()
	return err
}

// DisableGMCert 禁用条目
func (this *GMCertDAO) DisableGMCert(tx *dbs.Tx, certId int64) error {
	_, err := this.Query(tx).
		Pk(certId).
		Set("state", GMCertStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, certId)
}

// FindEnabledGMCert 查找启用中的条目
func (this *GMCertDAO) FindEnabledGMCert(tx *dbs.Tx, id int64) (*GMCert, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", GMCertStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*GMCert), err
}

// FindGMCertName 根据主键查找名称
func (this *GMCertDAO) FindGMCertName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateCert 创建证书
func (this *GMCertDAO) CreateCert(tx *dbs.Tx, adminId int64, userId int64, isOn bool, name string, description string, serverName string, signCertData []byte, signKeyData []byte, encCertData []byte, encKeyData []byte, timeBeginAt int64, timeEndAt int64, dnsNames []string, commonNames []string) (int64, error) {
	var op = NewGMCertOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.State = GMCertStateEnabled
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.ServerName = serverName

	op.SignCertData = signCertData
	op.SignKeyData = signKeyData
	op.EncCertData = encCertData
	op.EncKeyData = encKeyData
	op.TimeBeginAt = timeBeginAt
	op.TimeEndAt = timeEndAt

	dnsNamesJSON, err := json.Marshal(dnsNames)
	if err != nil {
		return 0, err
	}
	op.DnsNames = dnsNamesJSON

	commonNamesJSON, err := json.Marshal(commonNames)
	if err != nil {
		return 0, err
	}
	op.CommonNames = commonNamesJSON

	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateCert 修改证书
func (this *GMCertDAO) UpdateCert(tx *dbs.Tx,
	certId int64,
	isOn bool,
	name string,
	description string,
	serverName string,
	signCertData []byte, signKeyData []byte, encCertData []byte, encKeyData []byte,
	timeBeginAt int64,
	timeEndAt int64,
	dnsNames []string, commonNames []string) error {
	if certId <= 0 {
		return errors.New("invalid certId")
	}

	oldOne, err := this.Query(tx).Find()
	if err != nil {
		return err
	}
	if oldOne == nil {
		return nil
	}

	var op = NewGMCertOperator()
	op.Id = certId
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.ServerName = serverName

	// cert和key均为有重新上传才会修改
	if len(signCertData) > 0 {
		op.SignCertData = signCertData
	}
	if len(signKeyData) > 0 {
		op.SignKeyData = signKeyData
	}
	if len(encCertData) > 0 {
		op.EncCertData = encCertData
	}
	if len(encKeyData) > 0 {
		op.EncKeyData = encKeyData
	}

	op.TimeBeginAt = timeBeginAt
	op.TimeEndAt = timeEndAt

	dnsNamesJSON, err := json.Marshal(dnsNames)
	if err != nil {
		return err
	}
	op.DnsNames = dnsNamesJSON

	commonNamesJSON, err := json.Marshal(commonNames)
	if err != nil {
		return err
	}
	op.CommonNames = commonNamesJSON

	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, certId)
}

// ComposeCertConfig 组合配置
// ignoreData 是否忽略证书数据，避免因为数据过大影响传输
func (this *GMCertDAO) ComposeCertConfig(tx *dbs.Tx, certId int64, ignoreData bool, dataMap *shared.DataMap, cacheMap *utils.CacheMap) (*gmconfigs.GMCertConfig, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":ComposeCertConfig:" + types.String(certId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*gmconfigs.GMCertConfig), nil
	}

	cert, err := this.FindEnabledGMCert(tx, certId)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}

	var config = &gmconfigs.GMCertConfig{}
	config.Id = int64(cert.Id)
	config.IsOn = cert.IsOn
	config.Name = cert.Name
	config.Description = cert.Description
	if !ignoreData {
		if dataMap != nil {
			if len(cert.SignCertData) > 0 {
				config.SignCertData = dataMap.Put(cert.SignCertData)
			}
			if len(cert.SignKeyData) > 0 {
				config.SignKeyData = dataMap.Put(cert.SignKeyData)
			}
			if len(cert.EncCertData) > 0 {
				config.EncCertData = dataMap.Put(cert.EncCertData)
			}
			if len(cert.EncKeyData) > 0 {
				config.EncKeyData = dataMap.Put(cert.EncKeyData)
			}
		} else {
			config.SignCertData = cert.SignCertData
			config.SignKeyData = cert.SignKeyData
			config.EncCertData = cert.EncCertData
			config.EncKeyData = cert.EncKeyData
		}
	}
	config.ServerName = cert.ServerName
	config.TimeBeginAt = int64(cert.TimeBeginAt)
	config.TimeEndAt = int64(cert.TimeEndAt)

	if IsNotNull(cert.DnsNames) {
		var dnsNames = []string{}
		err := json.Unmarshal(cert.DnsNames, &dnsNames)
		if err != nil {
			return nil, err
		}
		config.DNSNames = dnsNames
	}

	if cert.CommonNames.IsNotNull() {
		var commonNames = []string{}
		err := json.Unmarshal(cert.CommonNames, &commonNames)
		if err != nil {
			return nil, err
		}
		config.CommonNames = commonNames
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// CountCerts 计算符合条件的证书数量
func (this *GMCertDAO) CountCerts(tx *dbs.Tx, isAvailable bool, isExpired bool, expiringDays int64, keyword string, userId int64, domains []string) (int64, error) {
	var query = this.Query(tx).
		State(GMCertStateEnabled)
	if isAvailable {
		query.Where("timeBeginAt<=UNIX_TIMESTAMP() AND timeEndAt>=UNIX_TIMESTAMP()")
	}
	if isExpired {
		query.Where("timeEndAt<UNIX_TIMESTAMP()")
	}
	if expiringDays > 0 {
		query.Where("timeEndAt>UNIX_TIMESTAMP() AND timeEndAt<:expiredAt").
			Param("expiredAt", time.Now().Unix()+expiringDays*86400)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword OR dnsNames LIKE :keyword OR commonNames LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		// 只查询管理员上传的
		query.Attr("userId", 0)
	}

	// 域名
	err := this.buildDomainSearchingQuery(query, domains)
	if err != nil {
		return 0, err
	}

	return query.Count()
}

// ListCertIds 列出符合条件的证书
func (this *GMCertDAO) ListCertIds(tx *dbs.Tx, isAvailable bool, isExpired bool, expiringDays int64, keyword string, userId int64, domains []string, offset int64, size int64) (certIds []int64, err error) {
	var query = this.Query(tx).
		State(GMCertStateEnabled)

	if isAvailable {
		query.Where("timeBeginAt<=UNIX_TIMESTAMP() AND timeEndAt>=UNIX_TIMESTAMP()")
	}
	if isExpired {
		query.Where("timeEndAt<UNIX_TIMESTAMP()")
	}
	if expiringDays > 0 {
		query.Where("timeEndAt>UNIX_TIMESTAMP() AND timeEndAt<:expiredAt").
			Param("expiredAt", time.Now().Unix()+expiringDays*86400)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword OR dnsNames LIKE :keyword OR commonNames LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		// 只查询管理员上传的
		query.Attr("userId", 0)
	}

	// 域名
	err = this.buildDomainSearchingQuery(query, domains)
	if err != nil {
		return nil, err
	}

	ones, err := query.
		ResultPk().
		DescPk().
		Offset(offset).
		Limit(size).
		FindAll()
	if err != nil {
		return nil, err
	}

	var result = []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*GMCert).Id))
	}
	return result, nil
}

// FindAllExpiringCerts 查找需要自动更新的任务
// 这里我们只返回有限的字段以节省内存
func (this *GMCertDAO) FindAllExpiringCerts(tx *dbs.Tx, days int) (result []*GMCert, err error) {
	if days < 0 {
		days = 0
	}

	var deltaSeconds = int64(days * 86400)
	_, err = this.Query(tx).
		State(GMCertStateEnabled).
		Attr("isOn", true).
		Where("FROM_UNIXTIME(timeEndAt, '%Y-%m-%d')=:day AND FROM_UNIXTIME(notifiedAt, '%Y-%m-%d')!=:today").
		Param("day", timeutil.FormatTime("Y-m-d", time.Now().Unix()+deltaSeconds)).
		Param("today", timeutil.Format("Y-m-d")).
		Result("id", "adminId", "userId", "timeEndAt", "name", "dnsNames", "notifiedAt", "acmeTaskId").
		Slice(&result).
		AscPk().
		FindAll()
	return
}

// UpdateCertNotifiedAt 设置当前证书事件通知时间
func (this *GMCertDAO) UpdateCertNotifiedAt(tx *dbs.Tx, certId int64) error {
	_, err := this.Query(tx).
		Pk(certId).
		Set("notifiedAt", time.Now().Unix()).
		Update()
	return err
}

// CheckUserCert 检查用户权限
func (this *GMCertDAO) CheckUserCert(tx *dbs.Tx, certId int64, userId int64) error {
	if certId <= 0 || userId <= 0 {
		return errors.New("not found")
	}
	ok, err := this.Query(tx).
		Pk(certId).
		Attr("userId", userId).
		State(GMCertStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not found")
	}
	return nil
}

// UpdateCertUser 修改证书所属用户
func (this *GMCertDAO) UpdateCertUser(tx *dbs.Tx, certId int64, userId int64) error {
	if certId <= 0 || userId <= 0 {
		return nil
	}
	return this.Query(tx).
		Pk(certId).
		Set("userId", userId).
		UpdateQuickly()
}

// NotifyUpdate 通知更新
func (this *GMCertDAO) NotifyUpdate(tx *dbs.Tx, certId int64) error {
	policyIds, err := SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithGmCertId(tx, certId)
	if err != nil {
		return err
	}
	if len(policyIds) == 0 {
		return nil
	}

	// 通知服务更新
	serverIds, err := SharedServerDAO.FindAllEnabledServerIdsWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return err
	}
	if len(serverIds) == 0 {
		return nil
	}
	for _, serverId := range serverIds {
		err := SharedServerDAO.NotifyUpdate(tx, serverId)
		if err != nil {
			return err
		}
	}

	// TODO 通知用户节点、API节点、管理系统（将来实现选择）更新

	return nil
}

// 构造通过域名搜索证书的查询对象
func (this *GMCertDAO) buildDomainSearchingQuery(query *dbs.Query, domains []string) error {
	if len(domains) == 0 {
		return nil
	}

	// 不要查询太多
	const maxDomains = 10_000
	if len(domains) > maxDomains {
		domains = domains[:maxDomains]
	}

	// 加入通配符
	var searchingDomains = []string{}
	var domainMap = map[string]bool{}
	for _, domain := range domains {
		domainMap[domain] = true
	}
	var reg = regexp.MustCompile(`^[\w*.-]+$`) // 为了下面的SQL语句安全先不支持其他字符
	for domain := range domainMap {
		if !reg.MatchString(domain) {
			continue
		}
		searchingDomains = append(searchingDomains, domain)

		if strings.Count(domain, ".") >= 2 && !strings.HasPrefix(domain, "*.") {
			var wildcardDomain = "*" + domain[strings.Index(domain, "."):]
			if !domainMap[wildcardDomain] {
				domainMap[wildcardDomain] = true
				searchingDomains = append(searchingDomains, wildcardDomain)
			}
		}
	}

	// 检测 JSON_OVERLAPS() 函数是否可用
	var canJSONOverlaps = false
	_, funcErr := this.Instance.FindCol(0, "SELECT JSON_OVERLAPS('[1]', '[1]')")
	canJSONOverlaps = funcErr == nil
	if canJSONOverlaps {
		domainsJSON, err := json.Marshal(searchingDomains)
		if err != nil {
			return err
		}

		query.
			Where("JSON_OVERLAPS(dnsNames, JSON_UNQUOTE(:domainsJSON))").
			Param("domainsJSON", string(domainsJSON))
		return nil
	}

	// 不支持JSON_OVERLAPS()的情形
	query.Reuse(false)

	// TODO 需要判断是否超出max_allowed_packet
	var sqlPieces = []string{}
	for _, domain := range searchingDomains {
		domainJSON, err := json.Marshal(domain)
		if err != nil {
			return err
		}

		sqlPieces = append(sqlPieces, "JSON_CONTAINS(dnsNames, '"+string(domainJSON)+"')")
	}
	if len(sqlPieces) > 0 {
		query.Where("(" + strings.Join(sqlPieces, " OR ") + ")")
	}

	return nil
}
