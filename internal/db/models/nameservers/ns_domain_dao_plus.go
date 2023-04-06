//go:build plus
// +build plus

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
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strings"
	"time"
)

const (
	NSDomainStateEnabled  = 1 // 已启用
	NSDomainStateDisabled = 0 // 已禁用
)

type NSDomainDAO dbs.DAO

func NewNSDomainDAO() *NSDomainDAO {
	return dbs.NewDAO(&NSDomainDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSDomains",
			Model:  new(NSDomain),
			PkName: "id",
		},
	}).(*NSDomainDAO)
}

var SharedNSDomainDAO *NSDomainDAO

func init() {
	dbs.OnReady(func() {
		SharedNSDomainDAO = NewNSDomainDAO()
	})
}

// EnableNSDomain 启用条目
func (this *NSDomainDAO) EnableNSDomain(tx *dbs.Tx, domainId int64) error {
	_, err := this.Query(tx).
		Pk(domainId).
		Set("state", NSDomainStateEnabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, domainId)
}

// DisableNSDomain 禁用条目
func (this *NSDomainDAO) DisableNSDomain(tx *dbs.Tx, domainId int64) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(domainId).
		Set("state", NSDomainStateDisabled).
		Set("version", version).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, domainId)
}

// FindEnabledNSDomain 查找启用中的条目
func (this *NSDomainDAO) FindEnabledNSDomain(tx *dbs.Tx, id int64) (*NSDomain, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSDomainStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSDomain), err
}

// FindNSDomainName 根据主键查找名称
func (this *NSDomainDAO) FindNSDomainName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateDomain 创建域名
func (this *NSDomainDAO) CreateDomain(tx *dbs.Tx, clusterId int64, userId int64, groupIds []int64, name string, status dnsconfigs.NSDomainStatus) (int64, error) {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return 0, err
	}

	var op = NewNSDomainOperator()
	op.ClusterId = clusterId
	op.UserId = userId

	// group ids
	if groupIds == nil {
		groupIds = []int64{}
	}
	groupIdsJSON, err := json.Marshal(groupIds)
	if err != nil {
		return 0, err
	}
	op.GroupIds = groupIdsJSON

	op.Name = strings.ToLower(name)
	op.Version = version
	op.IsOn = true
	if len(status) == 0 {
		op.Status = dnsconfigs.NSDomainStatusNone
	} else {
		op.Status = status
	}
	op.State = NSDomainStateEnabled
	domainId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = this.NotifyUpdate(tx, domainId)
	if err != nil {
		return domainId, err
	}
	return domainId, nil
}

// UpdateDomain 修改域名
// 不能允许用户修改域名名称，因为要重新验证
func (this *NSDomainDAO) UpdateDomain(tx *dbs.Tx, domainId int64, clusterId int64, userId int64, groupIds []int64, isOn bool) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}

	oldClusterId, err := this.Query(tx).
		Pk(domainId).
		Result("clusterId").
		FindInt64Col(0)
	if err != nil {
		return err
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	var op = NewNSDomainOperator()
	op.Id = domainId

	// 如果集群为0，表示不修改
	if clusterId > 0 {
		op.ClusterId = clusterId
	}

	op.UserId = userId

	// group ids
	if groupIds == nil {
		groupIds = []int64{}
	}
	groupIdsJSON, err := json.Marshal(groupIds)
	if err != nil {
		return err
	}
	op.GroupIds = groupIdsJSON

	op.IsOn = isOn
	op.Version = version
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	// 通知更新
	if clusterId > 0 && oldClusterId > 0 && oldClusterId != clusterId {
		err = models.SharedNSClusterDAO.NotifyUpdate(tx, oldClusterId)
		if err != nil {
			return err
		}
	}

	return this.NotifyUpdate(tx, domainId)
}

// UpdateDomainStatus 修改域名状态
func (this *NSDomainDAO) UpdateDomainStatus(tx *dbs.Tx, domainId int64, status dnsconfigs.NSDomainStatus) error {
	if !dnsconfigs.NSDomainStatusIsValid(status) {
		return errors.New("invalid status '" + status + "'")
	}

	if domainId <= 0 {
		return errors.New("invalid 'domainId'")
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Pk(domainId).
		Set("version", version).
		Set("status", status).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, domainId)
}

// CountAllEnabledDomains 计算域名数量
func (this *NSDomainDAO) CountAllEnabledDomains(tx *dbs.Tx, clusterId int64, userId int64, groupId int64, status dnsconfigs.NSDomainStatus, keyword string) (int64, error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	} else {
		query.Where("clusterId IN (SELECT id FROM " + models.SharedNSClusterDAO.Table + " WHERE state=1)")
	}

	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		query.Where("(userId=0 OR userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1))")
	}

	if groupId > 0 {
		query.JSONContains("groupIds", types.String(groupId))
	}

	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	if len(status) > 0 {
		query.Attr("status", status)
	}

	return query.
		State(NSDomainStateEnabled).
		Count()
}

// ListEnabledDomains 列出单页域名
func (this *NSDomainDAO) ListEnabledDomains(tx *dbs.Tx, clusterId int64, userId int64, groupId int64, keyword string, offset int64, size int64) (result []*NSDomain, err error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	} else {
		query.Where("clusterId IN (SELECT id FROM " + models.SharedNSClusterDAO.Table + " WHERE state=1)")
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		query.Where("(userId=0 OR userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1))")
	}

	if groupId > 0 {
		query.JSONContains("groupIds", types.String(groupId))
	}

	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	_, err = query.
		State(NSDomainStateEnabled).
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// IncreaseVersion 增加版本
func (this *NSDomainDAO) IncreaseVersion(tx *dbs.Tx) (int64, error) {
	return models.SharedSysLockerDAO.Increase(tx, "NS_DOMAIN_VERSION", 1)
}

// ListDomainsAfterVersion 列出某个版本后的域名
func (this *NSDomainDAO) ListDomainsAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*NSDomain, err error) {
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

// FindDomainIdWithName 根据名称查找域名ID
func (this *NSDomainDAO) FindDomainIdWithName(tx *dbs.Tx, clusterId int64, userId int64, name string) (int64, error) {
	if len(name) == 0 {
		return 0, nil
	}

	name = strings.ToLower(name)

	var query = this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		// 很重要，防止影响其他用户
		query.Attr("userId", 0)
	}
	return query.
		Attr("name", name).
		State(NSDomainStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// FindDomainWithName 根据名称查找域名
func (this *NSDomainDAO) FindDomainWithName(tx *dbs.Tx, userId int64, name string) (*NSDomain, error) {
	if len(name) == 0 {
		return nil, nil
	}

	var query = this.Query(tx)
	one, err := query.
		Attr("userId", userId).
		Attr("name", name).
		State(NSDomainStateEnabled).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*NSDomain), nil
}

// FindEnabledDomainTSIG 获取TSIG配置
func (this *NSDomainDAO) FindEnabledDomainTSIG(tx *dbs.Tx, domainId int64) ([]byte, error) {
	tsig, err := this.Query(tx).
		Pk(domainId).
		Result("tsig").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	return []byte(tsig), nil
}

// UpdateDomainTSIG 修改TSIG配置
func (this *NSDomainDAO) UpdateDomainTSIG(tx *dbs.Tx, domainId int64, tsigJSON []byte) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Pk(domainId).
		Set("tsig", tsigJSON).
		Set("version", version).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, domainId)
}

// FindEnabledDomainClusterId 获取域名的集群ID
func (this *NSDomainDAO) FindEnabledDomainClusterId(tx *dbs.Tx, domainId int64) (int64, error) {
	return this.Query(tx).
		Pk(domainId).
		State(NSDomainStateEnabled).
		Result("clusterId").
		FindInt64Col(0)
}

// ExistUserDomain 检查域名是否存在
func (this *NSDomainDAO) ExistUserDomain(tx *dbs.Tx, userId int64, name string) (bool, error) {
	name = strings.ToLower(name)
	return this.Query(tx).
		Attr("userId", userId).
		Attr("name", name).
		State(NSDomainStateEnabled).
		Exist()
}

// ExistVerifiedDomain 检查是否有验证通过的域名存在
func (this *NSDomainDAO) ExistVerifiedDomain(tx *dbs.Tx, name string) (bool, error) {
	name = strings.ToLower(name)
	return this.Query(tx).
		Attr("name", name).
		State(NSDomainStateEnabled).
		Attr("status", dnsconfigs.NSDomainStatusVerified).
		Exist()
}

// CheckUserDomain 检查用户域名
func (this *NSDomainDAO) CheckUserDomain(tx *dbs.Tx, userId int64, domainId int64) error {
	if userId <= 0 || domainId <= 0 {
		return models.ErrNotFound
	}

	b, err := this.Query(tx).
		Pk(domainId).
		Attr("userId", userId).
		State(NSDomainStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !b {
		return models.ErrNotFound
	}
	return nil
}

// DisableDomainWithName 使用名称删除域名
func (this *NSDomainDAO) DisableDomainWithName(tx *dbs.Tx, name string) error {
	if len(name) == 0 {
		return nil
	}
	domainId, err := this.Query(tx).
		ResultPk().
		State(NSDomainStateEnabled).
		Attr("userId", 0). // 不要删除用户的域名
		Attr("name", name).
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if domainId == 0 {
		return nil
	}
	return this.DisableNSDomain(tx, domainId)
}

// DisableUserDomainWithName 使用名称删除某个用户域名
func (this *NSDomainDAO) DisableUserDomainWithName(tx *dbs.Tx, userId int64, name string) error {
	if userId <= 0 || len(name) == 0 {
		return nil
	}
	domainId, err := this.Query(tx).
		ResultPk().
		State(NSDomainStateEnabled).
		Attr("userId", userId).
		Attr("name", name).
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if domainId == 0 {
		return nil
	}
	return this.DisableNSDomain(tx, domainId)
}

// FindDomainVerifyingInfo 查找域名验证信息
func (this *NSDomainDAO) FindDomainVerifyingInfo(tx *dbs.Tx, domainId int64, autoCreate bool) (*NSDomain, error) {
	one, err := this.Query(tx).
		Pk(domainId).
		State(NSDomainStateEnabled).
		Result("verifyTXT", "verifyExpiresAt", "status").
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	var domain = one.(*NSDomain)
	if autoCreate {
		if domain.Status == dnsconfigs.NSDomainStatusNone && (len(domain.VerifyTXT) == 0 || int64(domain.VerifyExpiresAt) < time.Now().Unix()) {
			// 生成一个
			var txt = rands.HexString(32)
			var expiresAt = time.Now().Unix() + 7200
			err = this.Query(tx).
				Pk(domainId).
				Set("verifyTXT", txt).
				Set("verifyExpiresAt", expiresAt).
				UpdateQuickly()
			if err != nil {
				return nil, err
			}
			domain.VerifyTXT = txt
			domain.VerifyExpiresAt = uint64(expiresAt)
		}
	}

	return domain, nil
}

// FindVerifiedDomainWithName 查询验证过的域名
func (this *NSDomainDAO) FindVerifiedDomainWithName(tx *dbs.Tx, name string) (*NSDomain, error) {
	one, err := this.Query(tx).
		State(NSDomainStateEnabled).
		Attr("name", name).
		Attr("status", dnsconfigs.NSDomainStatusVerified).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*NSDomain), nil
}

// NotifyUpdate 通知更改
func (this *NSDomainDAO) NotifyUpdate(tx *dbs.Tx, domainId int64) error {
	clusterId, err := this.Query(tx).
		Result("clusterId").
		Pk(domainId).
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if clusterId > 0 {
		return models.SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleDNS, clusterId, 0, models.NSNodeTaskTypeDomainChanged)
	}

	return nil
}
