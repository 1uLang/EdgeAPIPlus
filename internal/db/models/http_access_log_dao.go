package models

import (
	"encoding/json"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type HTTPAccessLogDAO dbs.DAO

var SharedHTTPAccessLogDAO *HTTPAccessLogDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPAccessLogDAO = NewHTTPAccessLogDAO()
	})
}

func NewHTTPAccessLogDAO() *HTTPAccessLogDAO {
	return dbs.NewDAO(&HTTPAccessLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPAccessLogs",
			Model:  new(HTTPAccessLog),
			PkName: "id",
		},
	}).(*HTTPAccessLogDAO)
}

// CreateHTTPAccessLogs 创建访问日志
func (this *HTTPAccessLogDAO) CreateHTTPAccessLogs(tx *dbs.Tx, accessLogs []*pb.HTTPAccessLog) error {
	dao := randomHTTPAccessLogDAO()
	if dao == nil {
		dao = &HTTPAccessLogDAOWrapper{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}
	}
	return this.CreateHTTPAccessLogsWithDAO(tx, dao, accessLogs)
}

// CreateHTTPAccessLogsWithDAO 使用特定的DAO创建访问日志
func (this *HTTPAccessLogDAO) CreateHTTPAccessLogsWithDAO(tx *dbs.Tx, daoWrapper *HTTPAccessLogDAOWrapper, accessLogs []*pb.HTTPAccessLog) error {
	if daoWrapper == nil {
		return errors.New("dao should not be nil")
	}
	if len(accessLogs) == 0 {
		return nil
	}

	dao := daoWrapper.DAO

	// TODO 改成事务批量提交，以加快速度

	for _, accessLog := range accessLogs {
		day := timeutil.Format("Ymd", time.Unix(accessLog.Timestamp, 0))
		tableDef, err := findHTTPAccessLogTable(dao.Instance, day, false)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{}
		fields["serverId"] = accessLog.ServerId
		fields["nodeId"] = accessLog.NodeId
		fields["status"] = accessLog.Status
		fields["createdAt"] = accessLog.Timestamp
		fields["requestId"] = accessLog.RequestId
		fields["firewallPolicyId"] = accessLog.FirewallPolicyId
		fields["firewallRuleGroupId"] = accessLog.FirewallRuleGroupId
		fields["firewallRuleSetId"] = accessLog.FirewallRuleSetId
		fields["firewallRuleId"] = accessLog.FirewallRuleId

		// TODO 根据集群、服务设置获取IP
		if tableDef.HasRemoteAddr {
			fields["remoteAddr"] = accessLog.RemoteAddr
		}
		if tableDef.HasDomain {
			fields["domain"] = accessLog.Host
		}

		content, err := json.Marshal(accessLog)
		if err != nil {
			return err
		}
		fields["content"] = content

		_, err = dao.Query(tx).
			Table(tableDef.Name).
			Sets(fields).
			Insert()
		if err != nil {
			// 是否为 Error 1146: Table 'xxx.xxx' doesn't exist  如果是，则创建表之后重试
			if strings.Contains(err.Error(), "1146") {
				tableDef, err = findHTTPAccessLogTable(dao.Instance, day, true)
				if err != nil {
					return err
				}
				_, err = dao.Query(tx).
					Table(tableDef.Name).
					Sets(fields).
					Insert()
				if err != nil {
					return err
				}
			} else {
				logs.Println("HTTP_ACCESS_LOG", err.Error())
			}
		}
	}

	return nil
}

// ListAccessLogs 读取往前的 单页访问日志
func (this *HTTPAccessLogDAO) ListAccessLogs(tx *dbs.Tx, lastRequestId string,
	size int64,
	day string,
	serverId int64,
	reverse bool,
	hasError bool,
	firewallPolicyId int64,
	firewallRuleGroupId int64,
	firewallRuleSetId int64,
	hasFirewallPolicy bool,
	userId int64,
	keyword string,
	ip string,
	domain string) (result []*HTTPAccessLog, nextLastRequestId string, hasMore bool, err error) {
	if len(day) != 8 {
		return
	}

	// 限制能查询的最大条数，防止占用内存过多
	if size > 1000 {
		size = 1000
	}

	result, nextLastRequestId, err = this.listAccessLogs(tx, lastRequestId, size, day, serverId, reverse, hasError, firewallPolicyId, firewallRuleGroupId, firewallRuleSetId, hasFirewallPolicy, userId, keyword, ip, domain)
	if err != nil || int64(len(result)) < size {
		return
	}

	moreResult, _, _ := this.listAccessLogs(tx, nextLastRequestId, 1, day, serverId, reverse, hasError, firewallPolicyId, firewallRuleGroupId, firewallRuleSetId, hasFirewallPolicy, userId, keyword, ip, domain)
	hasMore = len(moreResult) > 0
	return
}

// 读取往前的单页访问日志
func (this *HTTPAccessLogDAO) listAccessLogs(tx *dbs.Tx, lastRequestId string, size int64, day string, serverId int64, reverse bool, hasError bool, firewallPolicyId int64, firewallRuleGroupId int64, firewallRuleSetId int64, hasFirewallPolicy bool, userId int64, keyword string, ip string, domain string) (result []*HTTPAccessLog, nextLastRequestId string, err error) {
	if size <= 0 {
		return nil, lastRequestId, nil
	}

	serverIds := []int64{}
	if userId > 0 {
		serverIds, err = SharedServerDAO.FindAllEnabledServerIdsWithUserId(tx, userId)
		if err != nil {
			return
		}
		if len(serverIds) == 0 {
			return
		}
	}

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}

	locker := sync.Mutex{}

	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, daoWrapper := range daoList {
		go func(daoWrapper *HTTPAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO

			tableName, hasRemoteAddrField, hasDomainField, exists, err := findHTTPAccessLogTableName(dao.Instance, day)
			if !exists {
				// 表格不存在则跳过
				return
			}
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}

			query := dao.Query(tx)

			// 条件
			if serverId > 0 {
				query.Attr("serverId", serverId)
			} else if userId > 0 && len(serverIds) > 0 {
				query.Attr("serverId", serverIds).
					Reuse(false)
			}
			if hasError {
				query.Where("status>=400")
			}
			if firewallPolicyId > 0 {
				query.Attr("firewallPolicyId", firewallPolicyId)
			}
			if firewallRuleGroupId > 0 {
				query.Attr("firewallRuleGroupId", firewallRuleGroupId)
			}
			if firewallRuleSetId > 0 {
				query.Attr("firewallRuleSetId", firewallRuleSetId)
			}
			if hasFirewallPolicy {
				query.Where("firewallPolicyId>0")
				query.UseIndex("firewallPolicyId")
			}

			// keyword
			if len(ip) > 0 {
				// TODO 支持IP范围
				if hasRemoteAddrField {
					// IP格式
					if strings.Contains(ip, ",") || strings.Contains(ip, "-") {
						rangeConfig, err := shared.ParseIPRange(ip)
						if err == nil {
							if len(rangeConfig.IPFrom) > 0 && len(rangeConfig.IPTo) > 0 {
								query.Between("INET_ATON(remoteAddr)", utils.IP2Long(rangeConfig.IPFrom), utils.IP2Long(rangeConfig.IPTo))
							}
						}
					} else {
						query.Attr("remoteAddr", ip)
						query.UseIndex("remoteAddr")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.remoteAddr')=:ip1").
						Param("ip1", ip)
				}
			}
			if len(domain) > 0 {
				if hasDomainField {
					if strings.Contains(domain, "*") {
						domain = strings.ReplaceAll(domain, "*", "%")
						domain = regexp.MustCompile(`[^a-zA-Z0-9-.%]`).ReplaceAllString(domain, "")
						query.Where("domain LIKE :host2").
							Param("host2", domain)
					} else {
						query.Attr("domain", domain)
						query.UseIndex("domain")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.host')=:host1").
						Param("host1", domain)
				}
			}
			if len(keyword) > 0 {
				// remoteAddr
				if hasRemoteAddrField && net.ParseIP(keyword) != nil {
					query.Attr("remoteAddr", keyword)
				} else if hasRemoteAddrField && regexp.MustCompile(`^ip:.+`).MatchString(keyword) {
					keyword = keyword[3:]
					pieces := strings.SplitN(keyword, ",", 2)
					if len(pieces) == 1 || len(pieces[1]) == 0 {
						query.Attr("remoteAddr", pieces[0])
					} else {
						query.Between("INET_ATON(remoteAddr)", utils.IP2Long(pieces[0]), utils.IP2Long(pieces[1]))
					}
				} else {
					if regexp.MustCompile(`^ip:.+`).MatchString(keyword) {
						keyword = keyword[3:]
					}

					useOriginKeyword := false

					where := "JSON_EXTRACT(content, '$.remoteAddr') LIKE :keyword OR JSON_EXTRACT(content, '$.requestURI') LIKE :keyword OR JSON_EXTRACT(content, '$.host') LIKE :keyword OR JSON_EXTRACT(content, '$.userAgent') LIKE :keyword"

					jsonKeyword, err := json.Marshal(keyword)
					if err == nil {
						where += " OR JSON_CONTAINS(content, :jsonKeyword, '$.tags')"
						query.Param("jsonKeyword", jsonKeyword)
					}

					// 请求方法
					if keyword == http.MethodGet ||
						keyword == http.MethodPost ||
						keyword == http.MethodHead ||
						keyword == http.MethodConnect ||
						keyword == http.MethodPut ||
						keyword == http.MethodTrace ||
						keyword == http.MethodOptions ||
						keyword == http.MethodDelete ||
						keyword == http.MethodPatch {
						where += " OR JSON_EXTRACT(content, '$.requestMethod')=:originKeyword"
						useOriginKeyword = true
					}

					// 响应状态码
					if regexp.MustCompile(`^\d{3}$`).MatchString(keyword) {
						where += " OR JSON_EXTRACT(content, '$.status')=:intKeyword"
						query.Param("intKeyword", types.Int(keyword))
					}

					if regexp.MustCompile(`^\d{3}-\d{3}$`).MatchString(keyword) {
						pieces := strings.Split(keyword, "-")
						where += " OR JSON_EXTRACT(content, '$.status') BETWEEN :intKeyword1 AND :intKeyword2"
						query.Param("intKeyword1", types.Int(pieces[0]))
						query.Param("intKeyword2", types.Int(pieces[1]))
					}

					query.Where("("+where+")").
						Param("keyword", "%"+keyword+"%")
					if useOriginKeyword {
						query.Param("originKeyword", keyword)
					}
				}
			}

			// offset
			if len(lastRequestId) > 0 {
				if !reverse {
					query.Where("requestId<:requestId").
						Param("requestId", lastRequestId)
				} else {
					query.Where("requestId>:requestId").
						Param("requestId", lastRequestId)
				}
			}

			if !reverse {
				query.Desc("requestId")
			} else {
				query.Asc("requestId")
			}

			// 开始查询
			ones, err := query.
				Table(tableName).
				Limit(size).
				FindAll()
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			locker.Lock()
			for _, one := range ones {
				accessLog := one.(*HTTPAccessLog)
				result = append(result, accessLog)
			}
			locker.Unlock()
		}(daoWrapper)
	}
	wg.Wait()

	if len(result) == 0 {
		return nil, lastRequestId, nil
	}

	// 按照requestId排序
	sort.Slice(result, func(i, j int) bool {
		if !reverse {
			return result[i].RequestId > result[j].RequestId
		} else {
			return result[i].RequestId < result[j].RequestId
		}
	})

	if int64(len(result)) > size {
		result = result[:size]
	}

	requestId := result[len(result)-1].RequestId
	if reverse {
		lists.Reverse(result)
	}

	if !reverse {
		return result, requestId, nil
	} else {
		return result, requestId, nil
	}
}

// FindAccessLogWithRequestId 根据请求ID获取访问日志
func (this *HTTPAccessLogDAO) FindAccessLogWithRequestId(tx *dbs.Tx, requestId string) (*HTTPAccessLog, error) {
	if !regexp.MustCompile(`^\d{11,}`).MatchString(requestId) {
		return nil, errors.New("invalid requestId")
	}

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}

	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	var result *HTTPAccessLog = nil
	day := timeutil.FormatTime("Ymd", types.Int64(requestId[:10]))
	for _, daoWrapper := range daoList {
		go func(daoWrapper *HTTPAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO

			tableName, _, _, exists, err := findHTTPAccessLogTableName(dao.Instance, day)
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			if !exists {
				return
			}

			one, err := dao.Query(tx).
				Table(tableName).
				Attr("requestId", requestId).
				Find()
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			if one != nil {
				result = one.(*HTTPAccessLog)
			}
		}(daoWrapper)
	}
	wg.Wait()
	return result, nil
}

// SearchAccessLogs 根据请求ID获取访问日志
func (this *HTTPAccessLogDAO) SearchAccessLogs(tx *dbs.Tx, lastRequestId, day,
	ip, domain, code, method string, startAt, endAt uint64, userId int64, limit int64, allLog, errLog bool) (
	result []*HTTPAccessLog, nextRequestId string, err error) {

	if len(day) != 8 && limit < 0 {
		return
	}

	// 限制能查询的最大条数，防止占用内存过多
	if limit > 100 {
		limit = 100
	}

	serverIds := []int64{}
	if userId > 0 {
		serverIds, err = SharedServerDAO.FindAllEnabledServerIdsWithUserId(tx, userId)
		if err != nil {
			return
		}
		if len(serverIds) == 0 {
			return
		}
	}

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}
	var ids []int64
	wafLog := !allLog && !errLog
	if wafLog {
		ids, err = SharedHTTPFirewallRuleGroupDAO.FindRuleGroupIdWithCode(tx, code)
		if err != nil {
			return nil, "", err
		}
	}
	locker := sync.Mutex{}

	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, daoWrapper := range daoList {
		go func(daoWrapper *HTTPAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO

			tableName, hasRemoteAddrField, _, exists, err := findHTTPAccessLogTableName(dao.Instance, day)
			if !exists {
				// 表格不存在则跳过
				return
			}
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}

			query := dao.Query(tx)

			// 条件
			if userId > 0 && len(serverIds) > 0 {
				query.Attr("serverId", serverIds).
					Reuse(false)
			}
			// 时间条件限制
			if startAt > 0 && startAt < endAt {
				query.Where(fmt.Sprintf("createdAt>=%v", startAt))
				query.Where(fmt.Sprintf("createdAt<%v", endAt))
			}
			if wafLog {
				query.Where("firewallPolicyId>0")
				query.UseIndex("firewallPolicyId")
			}
			if errLog {
				query.Where("status>=400")
			}
			// keyword
			if len(ip) > 0 {
				// TODO 支持IP范围
				if hasRemoteAddrField {
					// IP格式
					if strings.Contains(ip, ",") || strings.Contains(ip, "-") {
						rangeConfig, err := shared.ParseIPRange(ip)
						if err == nil {
							if len(rangeConfig.IPFrom) > 0 && len(rangeConfig.IPTo) > 0 {
								query.Between("INET_ATON(remoteAddr)", utils.IP2Long(rangeConfig.IPFrom), utils.IP2Long(rangeConfig.IPTo))
							}
						}
					} else {
						query.Attr("remoteAddr", ip)
						query.UseIndex("remoteAddr")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.remoteAddr')=:ip1").
						Param("ip1", ip)
				}
			}
			if len(domain) > 0 {
				//if hasDomainField {
				//	if strings.Contains(domain, "*") {
				//		domain = strings.ReplaceAll(domain, "*", "%")
				//		domain = regexp.MustCompile(`[^a-zA-Z0-9-.%]`).ReplaceAllString(domain, "")
				//		query.Where("domain LIKE :host2").
				//			Param("host2", domain)
				//	} else {
				//		query.Attr("domain", domain)
				//		query.UseIndex("domain")
				//	}
				//} else {
				query.Where("JSON_EXTRACT(content, '$.host')=:host1").
					Param("host1", domain)
				//}
			}
			if len(method) > 0 {
				query.Where("JSON_EXTRACT(content, '$.requestMethod')=:method1").
					Param("method1", strings.ToUpper(method))
			}
			// offset
			if len(lastRequestId) > 0 {
				query.Where("requestId<:requestId").
					Param("requestId", lastRequestId)
			}
			query.Desc("requestId")
			if wafLog && len(ids) > 0 {
				query.Where(fmt.Sprintf("firewallRuleGroupId in (%s)", func(ids []int64) string {
					r := ""
					for _, id := range ids {
						r += fmt.Sprintf("%d,", id)
					}
					return r[:len(r)-1]
				}(ids)))
			}
			// 开始查询
			ones, err := query.
				Table(tableName).
				Limit(limit + 1).
				FindAll()
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			locker.Lock()
			for _, one := range ones {
				accessLog := one.(*HTTPAccessLog)
				result = append(result, accessLog)
			}
			locker.Unlock()
		}(daoWrapper)
	}
	wg.Wait()

	if len(result) == 0 {
		return nil, "", nil
	}

	// 按照requestId排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].RequestId > result[j].RequestId
	})

	if int64(len(result)) > limit && limit > 0 {
		nextRequestId = result[limit-1].RequestId
	}
	if int64(len(result)) >= limit {
		result = result[:limit]
	}

	return result, nextRequestId, nil

}

// StatisticsTop 统计指定域名的攻击ip排行
func (this *HTTPAccessLogDAO) StatisticsTop(tx *dbs.Tx,
	day string, userId int64, top int, ip2region func(string) (string, string)) (total int64, ips IpCount, region RegionCount, err error) {

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	var result []HTTPAccessLog
	serverIds := []int64{}

	if userId > 0 {
		serverIds, err = SharedServerDAO.FindAllEnabledServerIdsWithUserId(tx, userId)

		if err != nil {
			return
		}
		if len(serverIds) == 0 {
			return
		}
	}

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}

	locker := sync.Mutex{}

	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, daoWrapper := range daoList {
		go func(daoWrapper *HTTPAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO

			tableName, _, _, exists, err := findHTTPAccessLogTableName(dao.Instance, day)
			if !exists {
				// 表格不存在则跳过
				return
			}
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}

			query := dao.Query(tx)

			// 条件
			if userId > 0 && len(serverIds) > 0 {
				query.Attr("serverId", serverIds).
					Reuse(false)
			}
			query.Where("status>=400")

			var ones []*HTTPAccessLog
			// 开始查询
			_, err = query.
				Table(tableName).
				Group("remoteAddr").
				Result("remoteAddr, COUNT(1) AS count").
				Slice(&ones).
				FindAll()

			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			locker.Lock()
			for _, one := range ones {
				result = append(result, *one)
			}
			locker.Unlock()
		}(daoWrapper)
	}
	wg.Wait()
	ipCounts := make(map[string]int64, 0)
	regionCounts := make(map[string]int64, 0)

	for _, v := range result {

		total += v.Count
		province, city := ip2region(v.RemoteAddr)
		if province == "" {
			continue
		}
		regionCounts[province] += v.Count

		ipCounts[v.RemoteAddr+"("+province+city+")"] += v.Count
	}
	ipCountStu := IpCount{}
	regionCountStu := RegionCount{}
	for k, v := range ipCounts {
		ipCountStu.Count = append(ipCountStu.Count, v)
		ipCountStu.IP = append(ipCountStu.IP, k)
	}
	for k, v := range regionCounts {
		regionCountStu.Count = append(regionCountStu.Count, v)
		regionCountStu.Region = append(regionCountStu.Region, k)
	}
	sort.Sort(ipCountStu)
	sort.Sort(regionCountStu)
	min := len(ipCountStu.IP)
	if min > len(ipCountStu.Count) {
		min = len(ipCountStu.Count)
	}
	if min > top {
		min = top
	}
	ipCountStu.IP = ipCountStu.IP[:min]
	ipCountStu.Count = ipCountStu.Count[:min]

	min = len(regionCountStu.Region)
	if min > len(regionCountStu.Count) {
		min = len(regionCountStu.Count)
	}
	if min > top {
		min = top
	}
	regionCountStu.Region = regionCountStu.Region[:min]
	regionCountStu.Count = regionCountStu.Count[:min]

	return total, ipCountStu, regionCountStu, nil
}

type IpCount struct {
	IP    []string
	Count []int64
}

func (this IpCount) Len() int {
	min := len(this.Count)
	if min > len(this.IP) {
		return len(this.IP)
	}
	return min
}
func (this IpCount) Less(i, j int) bool {
	return this.Count[i] > this.Count[j]
}
func (this IpCount) Swap(i, j int) {
	this.IP[i], this.IP[j] = this.IP[j], this.IP[i]
	this.Count[i], this.Count[j] = this.Count[j], this.Count[i]
}

type RegionCount struct {
	Region []string
	Count  []int64
}

func (this RegionCount) Len() int {
	min := len(this.Count)
	if min > len(this.Region) {
		return len(this.Region)
	}
	return min
}
func (this RegionCount) Less(i, j int) bool {
	return this.Count[i] > this.Count[j]
}
func (this RegionCount) Swap(i, j int) {
	this.Region[i], this.Region[j] = this.Region[j], this.Region[i]
	this.Count[i], this.Count[j] = this.Count[j], this.Count[i]
}

// Statistics 统计指定提起下用户的攻击次数
func (this *HTTPAccessLogDAO) Statistics(tx *dbs.Tx, days []string, userId int64) (counts []int64, err error) {

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	serverIds := []int64{}

	if userId > 0 {
		serverIds, err = SharedServerDAO.FindAllEnabledServerIdsWithUserId(tx, userId)

		if err != nil {
			return
		}
		if len(serverIds) == 0 {
			return
		}
	}

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}
	if len(days) == 0 {
		return []int64{}, nil
	}
	counts = make([]int64, len(days))
	locker := sync.Mutex{}

	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, daoWrapper := range daoList {
		go func(daoWrapper *HTTPAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO
			dwg := &sync.WaitGroup{}
			dwg.Add(len(days))
			for k, day := range days {
				go func(k int, day string) {
					defer dwg.Done()
					tableName, _, _, exists, err := findHTTPAccessLogTableName(dao.Instance, day)
					if !exists {
						// 表格不存在则跳过
						return
					}
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}

					query := dao.Query(tx)

					// 条件
					if userId > 0 && len(serverIds) > 0 {
						query.Attr("serverId", serverIds).
							Reuse(false)
					}
					query.Where("status>=400")

					// 开始查询
					c, err := query.
						Table(tableName).
						Count()

					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					locker.Lock()
					counts[k] += c
					locker.Unlock()
				}(k, day)
			}
			dwg.Wait()
		}(daoWrapper)
	}
	wg.Wait()
	return counts, nil
}

// StatisticsType 统计各类型策略的条数
func (this *HTTPAccessLogDAO) StatisticsType(tx *dbs.Tx, day string, userId int64) (attacks []AttackType, err error) {

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	serverIds := []int64{}

	if userId > 0 {
		serverIds, err = SharedServerDAO.FindAllEnabledServerIdsWithUserId(tx, userId)

		if err != nil {
			return
		}
		if len(serverIds) == 0 {
			return
		}
	}

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}

	for _, group := range firewallconfigs.HTTPFirewallTemplate().Inbound.Groups {

		ids, err := SharedHTTPFirewallRuleGroupDAO.FindRuleGroupIdWithCode(tx, group.Code)
		if err != nil {
			return nil, err
		}
		attacks = append(attacks, AttackType{
			Code: group.Code,
			Name: group.Name,
			ids:  ids,
		})
	}

	locker := sync.Mutex{}
	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, daoWrapper := range daoList {
		go func(daoWrapper *HTTPAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO
			tableName, _, _, exists, err := findHTTPAccessLogTableName(dao.Instance, day)
			if !exists {
				// 表格不存在则跳过
				return
			}
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			gwg := &sync.WaitGroup{}
			gwg.Add(len(attacks))
			for k, group := range attacks {
				go func(k int, group AttackType) {
					defer gwg.Done()
					query := dao.Query(tx)
					// 条件
					if userId > 0 && len(serverIds) > 0 {
						query.Attr("serverId", serverIds).
							Reuse(false)
					}
					query.Where("status>=400")

					// 开始查询
					var c int64
					if len(group.ids) == 0 {
						c = 0
					} else {
						c, err = query.
							Table(tableName).
							Where(fmt.Sprintf("firewallRuleGroupId in (%s)", func(ids []int64) string {
								r := ""
								for _, id := range ids {
									r += fmt.Sprintf("%d,", id)
								}
								return r[:len(r)-1]
							}(group.ids))).
							Count()
						if err != nil {
							logs.Println("[DB_NODE]" + err.Error())
							return
						}
					}

					locker.Lock()
					attacks[k].Count += c
					locker.Unlock()
				}(k, group)
			}
			gwg.Wait()

		}(daoWrapper)
	}
	wg.Wait()

	return attacks, nil
}

type AttackType struct {
	Code  string  `json:"code"`  //攻击code
	Name  string  `json:"name"`  //攻击名称
	Count int64   `json:"count"` //条数
	ids   []int64 //对应策略分组id
}
