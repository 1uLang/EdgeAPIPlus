package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/1uLang/EdgeCommon/pkg/systemconfigs"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/zero"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"golang.org/x/net/idna"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type HTTPAccessLogDAO dbs.DAO

var SharedHTTPAccessLogDAO *HTTPAccessLogDAO

// 队列
var (
	oldAccessLogQueue       = make(chan *pb.HTTPAccessLog)
	accessLogQueue          = make(chan *pb.HTTPAccessLog, 10_000)
	accessLogQueueMaxLength = 100_000 // 队列最大长度
	accessLogQueuePercent   = 100     // 0-100
	accessLogCountPerSecond = 10_000  // 每秒钟写入条数，0 表示不限制
	accessLogPerTx          = 100     // 单事务写入条数
	accessLogConfigJSON     = []byte{}
	accessLogQueueChanged   = make(chan zero.Zero, 1)

	accessLogEnableAutoPartial       = true    // 是否启用自动分表
	accessLogRowsPerTable      int64 = 500_000 // 自动分表的单表最大值
)

type accessLogTableQuery struct {
	daoWrapper         *HTTPAccessLogDAOWrapper
	name               string
	hasRemoteAddrField bool
	hasDomainField     bool
}

func init() {
	dbs.OnReady(func() {
		SharedHTTPAccessLogDAO = NewHTTPAccessLogDAO()
	})

	// 队列相关
	dbs.OnReadyDone(func() {
		// 检查队列变化
		goman.New(func() {
			var ticker = time.NewTicker(60 * time.Second)

			// 先执行一次初始化
			SharedHTTPAccessLogDAO.SetupQueue()

			// 循环执行
			for {
				select {
				case <-ticker.C:
					SharedHTTPAccessLogDAO.SetupQueue()
				case <-accessLogQueueChanged:
					SharedHTTPAccessLogDAO.SetupQueue()
				}
			}
		})

		// 导出队列内容
		goman.New(func() {
			var ticker = time.NewTicker(1 * time.Second)
			var accessLogPerLoop = accessLogPerTx

			for range ticker.C {
				var countTxs = accessLogCountPerSecond / accessLogPerLoop
				if countTxs <= 0 {
					countTxs = 1
				}
				for i := 0; i < countTxs; i++ {
					var before = time.Now()
					hasMore, err := SharedHTTPAccessLogDAO.DumpAccessLogsFromQueue(accessLogPerLoop)

					// 如果用时过长，则调整每次写入次数
					var costMs = time.Since(before).Milliseconds()
					if costMs > 150 {
						accessLogPerLoop = accessLogPerTx / 4
					} else if costMs > 100 {
						accessLogPerLoop = accessLogPerTx / 2
					} // 这里不需要恢复成默认值，因为可能是写入数量比较小
					if err != nil {
						remotelogs.Error("HTTP_ACCESS_LOG_QUEUE", "dump access logs failed: "+err.Error())
					} else if !hasMore {
						break
					}
				}
			}
		})
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
	// 写入队列
	var queue = accessLogQueue // 这样写非常重要，防止在写入过程中队列有切换
	for _, accessLog := range accessLogs {
		if accessLog.FirewallPolicyId == 0 { // 如果是WAF记录，则采取采样率
			// 采样率
			if accessLogQueuePercent <= 0 {
				return nil
			}
			if accessLogQueuePercent < 100 && rands.Int(1, 100) > accessLogQueuePercent {
				return nil
			}
		}

		select {
		case queue <- accessLog:
		default:
			// 超出的丢弃
		}
	}

	return nil
}

// DumpAccessLogsFromQueue 从队列导入访问日志
func (this *HTTPAccessLogDAO) DumpAccessLogsFromQueue(size int) (hasMore bool, err error) {
	if size <= 0 {
		size = 100
	}

	var dao = randomHTTPAccessLogDAO()
	if dao == nil {
		dao = &HTTPAccessLogDAOWrapper{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}
	}

	if len(oldAccessLogQueue) == 0 && len(accessLogQueue) == 0 {
		return false, nil
	}

	// 开始事务
	tx, err := dao.DAO.Instance.Begin()
	if err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Commit()
	}()

	// 复制变量，防止中途改变
	var oldQueue = oldAccessLogQueue
	var newQueue = accessLogQueue

	hasMore = true

Loop:
	for i := 0; i < size; i++ {
		// old
		select {
		case accessLog := <-oldQueue:
			err := this.CreateHTTPAccessLog(tx, dao.DAO, accessLog)
			if err != nil {
				return false, err
			}
			continue Loop
		default:

		}

		// new
		select {
		case accessLog := <-newQueue:
			err := this.CreateHTTPAccessLog(tx, dao.DAO, accessLog)
			if err != nil {
				return false, err
			}
			continue Loop
		default:
			hasMore = false
			break Loop
		}
	}

	return hasMore, nil
}

// CreateHTTPAccessLog 写入单条访问日志
func (this *HTTPAccessLogDAO) CreateHTTPAccessLog(tx *dbs.Tx, dao *HTTPAccessLogDAO, accessLog *pb.HTTPAccessLog) error {
	var day = ""
	// 注意：如果你修改了 TimeISO8601 的逻辑，这里也需要同步修改
	if len(accessLog.TimeISO8601) > 10 {
		day = strings.ReplaceAll(accessLog.TimeISO8601[:10], "-", "")
	} else {
		timeutil.FormatTime("Ymd", accessLog.Timestamp)
	}

	tableDef, err := SharedHTTPAccessLogManager.FindLastTable(dao.Instance, day, true)
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

	if len(accessLog.RequestBody) > 0 {
		fields["requestBody"] = accessLog.RequestBody
		accessLog.RequestBody = nil
	}

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

	var lastId int64
	lastId, err = dao.Query(tx).
		Table(tableDef.Name).
		Sets(fields).
		Insert()
	if err != nil {
		// 错误重试
		if CheckSQLErrCode(err, 1146) { // Error 1146: Table 'xxx' doesn't exist
			err = SharedHTTPAccessLogManager.CreateTable(dao.Instance, tableDef.Name)
			if err != nil {
				return err
			}

			// 重新尝试
			lastId, err = dao.Query(tx).
				Table(tableDef.Name).
				Sets(fields).
				Insert()
		}

		if err != nil {
			return err
		}
	}

	if accessLogEnableAutoPartial && accessLogRowsPerTable > 0 && lastId >= accessLogRowsPerTable {
		SharedHTTPAccessLogManager.ResetTable(dao.Instance, day)
	}

	return nil
}

// ListAccessLogs 读取往前的 单页访问日志
func (this *HTTPAccessLogDAO) ListAccessLogs(tx *dbs.Tx,
	partition int32,
	lastRequestId string,
	size int64,
	day string,
	hourFrom string,
	hourTo string,
	clusterId int64,
	nodeId int64,
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

	result, nextLastRequestId, err = this.listAccessLogs(tx, partition, lastRequestId, size, day, hourFrom, hourTo, clusterId, nodeId, serverId, reverse, hasError, firewallPolicyId, firewallRuleGroupId, firewallRuleSetId, hasFirewallPolicy, userId, keyword, ip, domain)
	if err != nil || int64(len(result)) < size {
		return
	}

	moreResult, _, _ := this.listAccessLogs(tx, partition, nextLastRequestId, 1, day, hourFrom, hourTo, clusterId, nodeId, serverId, reverse, hasError, firewallPolicyId, firewallRuleGroupId, firewallRuleSetId, hasFirewallPolicy, userId, keyword, ip, domain)
	hasMore = len(moreResult) > 0
	return
}

// 读取往前的单页访问日志
func (this *HTTPAccessLogDAO) listAccessLogs(tx *dbs.Tx,
	partition int32,
	lastRequestId string,
	size int64,
	day string,
	hourFrom string,
	hourTo string,
	clusterId int64,
	nodeId int64,
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
	domain string) (result []*HTTPAccessLog, nextLastRequestId string, err error) {
	if size <= 0 {
		return nil, lastRequestId, nil
	}

	var serverIds = []int64{}
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
	var daoList = []*HTTPAccessLogDAOWrapper{}
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

	// 查询某个集群下的节点
	var nodeIds = []int64{}
	if clusterId > 0 {
		nodeIds, err = SharedNodeDAO.FindAllEnabledNodeIdsWithClusterId(tx, clusterId)
		if err != nil {
			remotelogs.Error("DB_NODE", err.Error())
			return
		}
		sort.Slice(nodeIds, func(i, j int) bool {
			return nodeIds[i] < nodeIds[j]
		})
	}

	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, partition)
	//	if err != nil {
	//		return nil, "", err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}
	//
	//// 检查各个分表是否一致
	//if partition < 0 {
	//	var newTableQueries = []*accessLogTableQuery{}
	//	for _, tableQuery := range tableQueries {
	//		if tableQuery.name != maxTableName {
	//			continue
	//		}
	//		newTableQueries = append(newTableQueries, tableQuery)
	//	}
	//	tableQueries = newTableQueries
	//}

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, "", err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return nil, "", nil
	}

	var locker = sync.Mutex{}

	// 这里正则表达式中的括号不能轻易变更，因为后面有引用
	// TODO 支持多个查询条件的组合，比如 status:200 proto:HTTP/1.1
	var statusPrefixReg = regexp.MustCompile(`status:\s*(\d{3})\b`)
	var statusRangeReg = regexp.MustCompile(`status:\s*(\d{3})-(\d{3})\b`)
	var urlReg = regexp.MustCompile(`^(http|https)://`)
	var requestPathReg = regexp.MustCompile(`requestPath:(\S+)`)
	var protoReg = regexp.MustCompile(`proto:(\S+)`)
	var schemeReg = regexp.MustCompile(`scheme:(\S+)`)

	var count = len(tableQueries)
	var wg = &sync.WaitGroup{}
	wg.Add(count)
	for _, tableQuery := range tableQueries {
		go func(tableQuery *accessLogTableQuery, keyword string) {
			defer wg.Done()

			var dao = tableQuery.daoWrapper.DAO
			var query = dao.Query(tx)
			query.Result("id", "serverId", "nodeId", "status", "createdAt", "content", "requestId", "firewallPolicyId", "firewallRuleGroupId", "firewallRuleSetId", "firewallRuleId", "remoteAddr", "domain")

			// 条件
			if nodeId > 0 {
				query.Attr("nodeId", nodeId)
			} else if clusterId > 0 {
				if len(nodeIds) > 0 {
					var nodeIdStrings = []string{}
					for _, subNodeId := range nodeIds {
						nodeIdStrings = append(nodeIdStrings, types.String(subNodeId))
					}

					query.Where("nodeId IN (" + strings.Join(nodeIdStrings, ",") + ")")
					query.Reuse(false)
				} else {
					// 如果没有节点，则直接返回空
					return
				}
			}
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
				if tableQuery.hasRemoteAddrField {
					// IP格式
					if strings.Contains(ip, ",") || strings.Contains(ip, "-") {
						rangeConfig, err := shared.ParseIPRange(ip)
						if err == nil {
							if len(rangeConfig.IPFrom) > 0 && len(rangeConfig.IPTo) > 0 {
								query.Between("INET_ATON(remoteAddr)", utils.IP2Long(rangeConfig.IPFrom), utils.IP2Long(rangeConfig.IPTo))
							}
						}
					} else {
						// 去掉IPv6的[]
						ip = strings.Trim(ip, "[]")

						query.Attr("remoteAddr", ip)
						query.UseIndex("remoteAddr")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.remoteAddr')=:ip1").
						Param("ip1", ip)
				}
			}
			if len(domain) > 0 {
				if tableQuery.hasDomainField {
					if strings.Contains(domain, "*") {
						domain = strings.ReplaceAll(domain, "*", "%")
						domain = regexp.MustCompile(`[^a-zA-Z0-9-.%]`).ReplaceAllString(domain, "")
						query.Where("domain LIKE :host2").
							Param("host2", domain)
					} else {
						// 中文域名
						if !regexp.MustCompile(`^[a-zA-Z0-9-.]+$`).MatchString(domain) {
							unicodeDomain, err := idna.ToASCII(domain)
							if err == nil && len(unicodeDomain) > 0 {
								domain = unicodeDomain
							}
						}

						query.Attr("domain", domain)
						query.UseIndex("domain")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.host')=:host1").
						Param("host1", domain)
				}
			}

			if len(keyword) > 0 {
				var isSpecialKeyword = false

				if tableQuery.hasRemoteAddrField && net.ParseIP(keyword) != nil { // ip
					isSpecialKeyword = true
					query.Attr("remoteAddr", keyword)
				} else if tableQuery.hasRemoteAddrField && regexp.MustCompile(`^ip:.+`).MatchString(keyword) { // ip:x.x.x.x
					isSpecialKeyword = true
					keyword = keyword[3:]
					pieces := strings.SplitN(keyword, ",", 2)
					if len(pieces) == 1 || len(pieces[1]) == 0 || pieces[0] == pieces[1] {
						query.Attr("remoteAddr", pieces[0])
					} else {
						query.Between("INET_ATON(remoteAddr)", utils.IP2Long(pieces[0]), utils.IP2Long(pieces[1]))
					}
				} else if statusRangeReg.MatchString(keyword) { // status:200-400
					isSpecialKeyword = true
					var matches = statusRangeReg.FindStringSubmatch(keyword)
					query.Between("status", types.Int(matches[1]), types.Int(matches[2]))
				} else if statusPrefixReg.MatchString(keyword) { // status:200
					isSpecialKeyword = true
					var matches = statusPrefixReg.FindStringSubmatch(keyword)
					query.Attr("status", matches[1])
				} else if requestPathReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = requestPathReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.requestPath')=:keyword").
						Param("keyword", matches[1])
				} else if protoReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = protoReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.proto')=:keyword").
						Param("keyword", strings.ToUpper(matches[1]))
				} else if schemeReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = schemeReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.scheme')=:keyword").
						Param("keyword", strings.ToLower(matches[1]))
				} else if urlReg.MatchString(keyword) { // https://xxx/yyy
					u, err := url.Parse(keyword)
					if err == nil {
						isSpecialKeyword = true
						query.Attr("domain", u.Host)
						query.Where("JSON_EXTRACT(content, '$.requestURI') LIKE :keyword").
							Param("keyword", dbutils.QuoteLikePrefix("\""+u.RequestURI()))
					}
				}
				if !isSpecialKeyword {
					if regexp.MustCompile(`^ip:.+`).MatchString(keyword) {
						keyword = keyword[3:]
					}

					var useOriginKeyword = false

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
						where += " OR status=:intKeyword"
						query.Param("intKeyword", types.Int(keyword))
					}

					if regexp.MustCompile(`^\d{3}-\d{3}$`).MatchString(keyword) {
						pieces := strings.Split(keyword, "-")
						where += " OR status BETWEEN :intKeyword1 AND :intKeyword2"
						query.Param("intKeyword1", types.Int(pieces[0]))
						query.Param("intKeyword2", types.Int(pieces[1]))
					}

					if regexp.MustCompile(`^\d{20,}\s*\.?$`).MatchString(keyword) {
						where += " OR requestId=:requestId"
						query.Param("requestId", strings.TrimRight(keyword, ". "))
					}

					query.Where("("+where+")").
						Param("keyword", dbutils.QuoteLike(keyword))
					if useOriginKeyword {
						query.Param("originKeyword", keyword)
					}
				}
			}

			// hourFrom - hourTo
			if len(hourFrom) > 0 && len(hourTo) > 0 {
				var hourFromInt = types.Int(hourFrom)
				var hourToInt = types.Int(hourTo)
				if hourFromInt >= 0 && hourFromInt <= 23 && hourToInt >= hourFromInt && hourToInt <= 23 {
					var y = types.Int(day[:4])
					var m = types.Int(day[4:6])
					var d = types.Int(day[6:])
					var timeFrom = time.Date(y, time.Month(m), d, hourFromInt, 0, 0, 0, time.Local)
					var timeTo = time.Date(y, time.Month(m), d, hourToInt, 59, 59, 0, time.Local)
					query.Between("createdAt", timeFrom.Unix(), timeTo.Unix())
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
				Table(tableQuery.name).
				Limit(size).
				FindAll()
			if err != nil {
				remotelogs.Println("DB_NODE", err.Error())
				return
			}

			locker.Lock()
			for _, one := range ones {
				var accessLog = one.(*HTTPAccessLog)
				result = append(result, accessLog)
			}
			locker.Unlock()
		}(tableQuery, keyword)
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

	var requestId = result[len(result)-1].RequestId
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

	// 准备查询
	var day = timeutil.FormatTime("Ymd", types.Int64(requestId[:10]))
	var tableQueries = []*accessLogTableQuery{}
	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		tableDefs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}
		for _, def := range tableDefs {
			tableQueries = append(tableQueries, &accessLogTableQuery{
				daoWrapper:         daoWrapper,
				name:               def.Name,
				hasRemoteAddrField: def.HasRemoteAddr,
				hasDomainField:     def.HasDomain,
			})
		}
	}

	var count = len(tableQueries)
	var wg = &sync.WaitGroup{}
	wg.Add(count)
	var result *HTTPAccessLog = nil
	for _, tableQuery := range tableQueries {
		go func(tableQuery *accessLogTableQuery) {
			defer wg.Done()

			var dao = tableQuery.daoWrapper.DAO
			one, err := dao.Query(tx).
				Table(tableQuery.name).
				Attr("requestId", requestId).
				Find()
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			if one != nil {
				result = one.(*HTTPAccessLog)
			}
		}(tableQuery)
	}
	wg.Wait()
	return result, nil
}

// SetupQueue 建立队列
func (this *HTTPAccessLogDAO) SetupQueue() {
	configJSON, err := SharedSysSettingDAO.ReadSetting(nil, systemconfigs.SettingCodeAccessLogQueue)
	if err != nil {
		remotelogs.Error("HTTP_ACCESS_LOG_QUEUE", "read settings failed: "+err.Error())
		return
	}

	if len(configJSON) == 0 {
		return
	}

	if bytes.Equal(accessLogConfigJSON, configJSON) {
		return
	}
	accessLogConfigJSON = configJSON

	var config = &serverconfigs.AccessLogQueueConfig{}
	err = json.Unmarshal(configJSON, config)
	if err != nil {
		remotelogs.Error("HTTP_ACCESS_LOG_QUEUE", "decode settings failed: "+err.Error())
		return
	}

	accessLogQueuePercent = config.Percent
	accessLogCountPerSecond = config.CountPerSecond
	if accessLogCountPerSecond <= 0 {
		accessLogCountPerSecond = 10_000
	}
	if config.MaxLength <= 0 {
		config.MaxLength = 100_000
	}

	accessLogEnableAutoPartial = config.EnableAutoPartial
	if config.RowsPerTable > 0 {
		accessLogRowsPerTable = config.RowsPerTable
	}

	if accessLogQueueMaxLength != config.MaxLength {
		accessLogQueueMaxLength = config.MaxLength
		oldAccessLogQueue = accessLogQueue
		accessLogQueue = make(chan *pb.HTTPAccessLog, config.MaxLength)
	}

	if Tea.IsTesting() {
		remotelogs.Println("HTTP_ACCESS_LOG_QUEUE", "change queue "+string(configJSON))
	}
}

// SearchAccessLogs 根据请求ID获取访问日志
func (this *HTTPAccessLogDAO) SearchAccessLogs(tx *dbs.Tx, lastRequestId, day,
	ip, domain, code, method, keyword string, startAt, endAt uint64, userId int64, limit int64, allLog, errLog bool) (
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
	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
	//	if err != nil {
	//		return nil, "", err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}
	//
	//// 检查各个分表是否一致
	//var newTableQueries = []*accessLogTableQuery{}
	//for _, tableQuery := range tableQueries {
	//	if tableQuery.name != maxTableName {
	//		continue
	//	}
	//	newTableQueries = append(newTableQueries, tableQuery)
	//}
	//tableQueries = newTableQueries

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, "", err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return nil, "", nil
	}
	locker := sync.Mutex{}

	// 这里正则表达式中的括号不能轻易变更，因为后面有引用
	// TODO 支持多个查询条件的组合，比如 status:200 proto:HTTP/1.1
	var statusPrefixReg = regexp.MustCompile(`status:\s*(\d{3})\b`)
	var statusRangeReg = regexp.MustCompile(`status:\s*(\d{3})-(\d{3})\b`)
	var urlReg = regexp.MustCompile(`^(http|https)://`)
	var requestPathReg = regexp.MustCompile(`requestPath:(\S+)`)
	var protoReg = regexp.MustCompile(`proto:(\S+)`)
	var schemeReg = regexp.MustCompile(`scheme:(\S+)`)

	var count = len(tableQueries)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, tableQuery := range tableQueries {
		go func(tableQuery *accessLogTableQuery, keyword string) {
			defer wg.Done()

			var dao = tableQuery.daoWrapper.DAO
			var query = dao.Query(tx)
			query.Result("id", "serverId", "nodeId", "status", "createdAt", "content", "requestId", "firewallPolicyId", "firewallRuleGroupId", "firewallRuleSetId", "firewallRuleId", "remoteAddr", "domain")

			// 条件
			if userId > 0 && len(serverIds) > 0 {
				query.Attr("serverId", serverIds).
					Reuse(false)
			}
			// 时间条件限制
			if startAt > 0 && startAt < endAt {
				query.Between("createdAt", startAt, endAt)
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
				if tableQuery.hasRemoteAddrField {
					// IP格式
					if strings.Contains(ip, ",") || strings.Contains(ip, "-") {
						rangeConfig, err := shared.ParseIPRange(ip)
						if err == nil {
							if len(rangeConfig.IPFrom) > 0 && len(rangeConfig.IPTo) > 0 {
								query.Between("INET_ATON(remoteAddr)", utils.IP2Long(rangeConfig.IPFrom), utils.IP2Long(rangeConfig.IPTo))
							}
						}
					} else {
						// 去掉IPv6的[]
						ip = strings.Trim(ip, "[]")

						query.Attr("remoteAddr", ip)
						query.UseIndex("remoteAddr")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.remoteAddr')=:ip1").
						Param("ip1", ip)
				}
			}
			if len(domain) > 0 {
				if tableQuery.hasDomainField {
					if strings.Contains(domain, "*") {
						domain = strings.ReplaceAll(domain, "*", "%")
						domain = regexp.MustCompile(`[^a-zA-Z0-9-.%]`).ReplaceAllString(domain, "")
						query.Where("domain LIKE :host2").
							Param("host2", domain)
					} else {
						// 中文域名
						if !regexp.MustCompile(`^[a-zA-Z0-9-.]+$`).MatchString(domain) {
							unicodeDomain, err := idna.ToASCII(domain)
							if err == nil && len(unicodeDomain) > 0 {
								domain = unicodeDomain
							}
						}

						query.Attr("domain", domain)
						query.UseIndex("domain")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.host')=:host1").
						Param("host1", domain)
				}
			}
			if len(method) > 0 {
				query.Where("JSON_EXTRACT(content, '$.requestMethod')=:method1").
					Param("method1", strings.ToUpper(method))
			}
			if len(keyword) > 0 {
				var isSpecialKeyword = false
				if tableQuery.hasRemoteAddrField && net.ParseIP(keyword) != nil { // ip
					isSpecialKeyword = true
					query.Attr("remoteAddr", keyword)
				} else if tableQuery.hasRemoteAddrField && regexp.MustCompile(`^ip:.+`).MatchString(keyword) { // ip:x.x.x.x
					isSpecialKeyword = true
					keyword = keyword[3:]
					pieces := strings.SplitN(keyword, ",", 2)
					if len(pieces) == 1 || len(pieces[1]) == 0 || pieces[0] == pieces[1] {
						query.Attr("remoteAddr", pieces[0])
					} else {
						query.Between("INET_ATON(remoteAddr)", utils.IP2Long(pieces[0]), utils.IP2Long(pieces[1]))
					}
				} else if statusRangeReg.MatchString(keyword) { // status:200-400
					isSpecialKeyword = true
					var matches = statusRangeReg.FindStringSubmatch(keyword)
					query.Between("status", types.Int(matches[1]), types.Int(matches[2]))
				} else if statusPrefixReg.MatchString(keyword) { // status:200
					isSpecialKeyword = true
					var matches = statusPrefixReg.FindStringSubmatch(keyword)
					query.Attr("status", matches[1])
				} else if requestPathReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = requestPathReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.requestPath')=:keyword").
						Param("keyword", matches[1])
				} else if protoReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = protoReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.proto')=:keyword").
						Param("keyword", strings.ToUpper(matches[1]))
				} else if schemeReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = schemeReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.scheme')=:keyword").
						Param("keyword", strings.ToLower(matches[1]))
				} else if urlReg.MatchString(keyword) { // https://xxx/yyy
					u, err := url.Parse(keyword)
					if err == nil {
						isSpecialKeyword = true
						query.Attr("domain", u.Host)
						query.Where("JSON_EXTRACT(content, '$.requestURI') LIKE :keyword").
							Param("keyword", dbutils.QuoteLikePrefix("\""+u.RequestURI()))
					}
				}
				if !isSpecialKeyword {
					if regexp.MustCompile(`^ip:.+`).MatchString(keyword) {
						keyword = keyword[3:]
					}

					var useOriginKeyword = false

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
						where += " OR status=:intKeyword"
						query.Param("intKeyword", types.Int(keyword))
					}

					if regexp.MustCompile(`^\d{3}-\d{3}$`).MatchString(keyword) {
						pieces := strings.Split(keyword, "-")
						where += " OR status BETWEEN :intKeyword1 AND :intKeyword2"
						query.Param("intKeyword1", types.Int(pieces[0]))
						query.Param("intKeyword2", types.Int(pieces[1]))
					}

					if regexp.MustCompile(`^\d{20,}\s*\.?$`).MatchString(keyword) {
						where += " OR requestId=:requestId"
						query.Param("requestId", strings.TrimRight(keyword, ". "))
					}

					query.Where("("+where+")").
						Param("keyword", dbutils.QuoteLike(keyword))
					if useOriginKeyword {
						query.Param("originKeyword", keyword)
					}
				}
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
				Table(tableQuery.name).
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
		}(tableQuery, keyword)
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
	day string, userId int64, ip2region func(string) (string, string, string)) (resp *StatisticsTop, err error) {

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	serverIds := []int64{}
	resp = &StatisticsTop{Tops: make([]*StatisticsTopItem, 0)}
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
	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}
	//
	//// 检查各个分表是否一致
	//
	//var newTableQueries = []*accessLogTableQuery{}
	//for _, tableQuery := range tableQueries {
	//	if tableQuery.name != maxTableName {
	//		continue
	//	}
	//	newTableQueries = append(newTableQueries, tableQuery)
	//}
	//tableQueries = newTableQueries

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return
	}
	locker := sync.Mutex{}
	var result []*HTTPAccessLog
	wg := &sync.WaitGroup{}
	wg.Add(len(serverIds))
	for _, serverId := range serverIds {

		go func(serverId int64) {
			defer wg.Done()

			dwg := &sync.WaitGroup{}
			dwg.Add(len(tableQueries))
			l := sync.Mutex{}
			serverLogs := make([]*HTTPAccessLog, 0)
			for _, tableQuery := range tableQueries {
				go func(tableQuery *accessLogTableQuery) {
					defer dwg.Done()

					var dao = tableQuery.daoWrapper.DAO
					var query = dao.Query(tx)
					// 条件
					query.Attr("serverId", serverId).
						Reuse(false)

					query.Where("firewallPolicyId>0")
					query.UseIndex("firewallPolicyId")

					var ones []*HTTPAccessLog
					// 开始查询
					_, err = query.
						Table(tableQuery.name).
						Group("remoteAddr").
						Result("remoteAddr, COUNT(1) AS count").
						Slice(&ones).
						FindAll()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					l.Lock()
					serverLogs = append(serverLogs, ones...)
					l.Unlock()
				}(tableQuery)
			}
			dwg.Wait()
			// 统计

			mergeLogs, total, ips, region := statisticsTop(serverLogs, ip2region)

			locker.Lock()
			result = append(result, mergeLogs...)
			resp.Tops = append(resp.Tops, &StatisticsTopItem{ServerId: serverId, Total: total, Ips: ips, Region: region})
			locker.Unlock()
		}(serverId)
	}
	wg.Wait()
	_, total, ips, region := statisticsTop(result, ip2region)
	resp.Tops = append(resp.Tops, &StatisticsTopItem{ServerId: 0, Total: total, Ips: ips, Region: region})
	return resp, nil
}
func statisticsTop(result []*HTTPAccessLog, ip2region func(string) (string, string, string)) (mergeLogs []*HTTPAccessLog, total int64, ips IpCount, region RegionCount) {

	// ip+city
	ipCounts := make(map[string]int64, 0)
	// ip+city - ip
	ipCity := make(map[string]string, 0)
	regionCounts := make(map[string]int64, 0)

	for _, v := range result {
		total += v.Count

		country, province, city := ip2region(v.RemoteAddr)
		if province == "" {
			province = country
		}
		regionCounts[province] += v.Count

		ipCounts[v.RemoteAddr+"("+country+province+city+")"] += v.Count
		ipCity[v.RemoteAddr+"("+country+province+city+")"] = v.RemoteAddr
	}
	ipCountStu := IpCount{}
	regionCountStu := RegionCount{}
	for k, v := range ipCounts {
		ipCountStu.Count = append(ipCountStu.Count, v)
		ipCountStu.IP = append(ipCountStu.IP, k)
		mergeLogs = append(mergeLogs, &HTTPAccessLog{
			RemoteAddr: ipCity[k],
			Count:      v,
		})
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
	ipCountStu.IP = ipCountStu.IP[:min]
	ipCountStu.Count = ipCountStu.Count[:min]

	min = len(regionCountStu.Region)
	if min > len(regionCountStu.Count) {
		min = len(regionCountStu.Count)
	}
	regionCountStu.Region = regionCountStu.Region[:min]
	regionCountStu.Count = regionCountStu.Count[:min]
	return mergeLogs, total, ipCountStu, regionCountStu
}

type StatisticsTop struct {
	Tops []*StatisticsTopItem `json:"tops"`
}
type StatisticsTopItem struct {
	ServerId int64 `json:"server_id"`
	Total    int64
	Ips      IpCount
	Region   RegionCount
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

type CodeCount struct {
	Code  []uint32
	Count []int64
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
	wg := &sync.WaitGroup{}
	wg.Add(len(days))
	for k, day := range days {
		go func(k int, day string) {
			defer wg.Done()
			// 准备查询
			var tableQueries = []*accessLogTableQuery{}
			//var maxTableName = ""
			//for _, daoWrapper := range daoList {
			//	var instance = daoWrapper.DAO.Instance
			//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
			//	if err != nil {
			//		return
			//	}
			//	if !def.Exists {
			//		continue
			//	}
			//
			//	if len(maxTableName) == 0 || def.Name > maxTableName {
			//		maxTableName = def.Name
			//	}
			//
			//	tableQueries = append(tableQueries, &accessLogTableQuery{
			//		daoWrapper:         daoWrapper,
			//		name:               def.Name,
			//		hasRemoteAddrField: def.HasRemoteAddr,
			//		hasDomainField:     def.HasDomain,
			//	})
			//}
			//
			//// 检查各个分表是否一致
			//
			//var newTableQueries = []*accessLogTableQuery{}
			//for _, tableQuery := range tableQueries {
			//	if tableQuery.name != maxTableName {
			//		continue
			//	}
			//	newTableQueries = append(newTableQueries, tableQuery)
			//}
			//tableQueries = newTableQueries

			for _, daoWrapper := range daoList {
				var instance = daoWrapper.DAO.Instance
				defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
				if err != nil {
					return
				}

				if len(defs) > 0 {
					for _, def := range defs {
						tableQueries = append(tableQueries, &accessLogTableQuery{
							daoWrapper:         daoWrapper,
							name:               def.Name,
							hasRemoteAddrField: def.HasRemoteAddr,
							hasDomainField:     def.HasDomain,
						})
					}
				}

			}

			if len(tableQueries) == 0 {
				return
			}
			dwg := &sync.WaitGroup{}
			dwg.Add(len(tableQueries))
			for _, tableQuery := range tableQueries {
				go func(tableQuery *accessLogTableQuery) {
					defer dwg.Done()
					var dao = tableQuery.daoWrapper.DAO
					var query = dao.Query(tx)

					// 条件
					if userId > 0 && len(serverIds) > 0 {
						query.Attr("serverId", serverIds).
							Reuse(false)
					}
					query.Where("firewallPolicyId>0")
					query.UseIndex("firewallPolicyId")
					// 开始查询
					c, err := query.
						Table(tableQuery.name).
						Count()

					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					locker.Lock()
					counts[k] += c
					locker.Unlock()
				}(tableQuery)
			}
			dwg.Wait()
		}(k, day)
	}
	wg.Wait()
	return counts, nil
}

// StatisticsType 统计各类型策略的条数
func (this *HTTPAccessLogDAO) StatisticsType(tx *dbs.Tx, day string, userId int64) (attacks []*AttackType, err error) {

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
	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}

	// 检查各个分表是否一致
	//var newTableQueries = []*accessLogTableQuery{}
	//for _, tableQuery := range tableQueries {
	//	if tableQuery.name != maxTableName {
	//		continue
	//	}
	//	newTableQueries = append(newTableQueries, tableQuery)
	//}
	//tableQueries = newTableQueries

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return
	}
	var attacks1 = make([]AttackType, 0)
	attacks = make([]*AttackType, 0)

	// 统计所有waf策略
	for _, group := range firewallconfigs.HTTPFirewallTemplate().Inbound.Groups {

		ids, err := SharedHTTPFirewallRuleGroupDAO.FindRuleGroupIdWithCode(tx, group.Code)
		if err != nil {
			return nil, err
		}
		attacks1 = append(attacks1, AttackType{
			Code: group.Code,
			Name: group.Name,
			ids:  ids,
		})
	}

	locker := sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(serverIds))

	for _, serverId := range serverIds {
		go func(serverId int64) {
			defer wg.Done()

			gwg := &sync.WaitGroup{}
			gwg.Add(len(attacks1))
			for k, group := range attacks1 {
				go func(k int, group AttackType) {
					defer gwg.Done()
					dwg := &sync.WaitGroup{}
					dwg.Add(len(tableQueries))
					var count int64
					for _, tableQuery := range tableQueries {
						go func(tableQuery *accessLogTableQuery, count *int64) {
							defer dwg.Done()
							var dao = tableQuery.daoWrapper.DAO
							var query = dao.Query(tx)
							// 条件
							query.Attr("serverId", serverId).
								Reuse(false)
							query.Where("firewallPolicyId>0")
							query.UseIndex("firewallPolicyId")
							// 开始查询
							var c int64
							if len(group.ids) == 0 {
								c = 0
							} else {
								c, err = query.
									Table(tableQuery.name).
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
							*count += c
							locker.Unlock()
						}(tableQuery, &count)
					}
					dwg.Wait()
					locker.Lock()
					attacks = append(attacks, &AttackType{
						Code:     attacks1[k].Code,
						Name:     attacks1[k].Name,
						ServerId: serverId,
						Count:    count,
					})
					locker.Unlock()
				}(k, group)
			}
			gwg.Wait()
		}(serverId)
	}
	wg.Wait()

	return attacks, nil
}

// AccessStatistics 统计访问条数 访问总次数  防护总次数 访问IP总数 拦截IP总数
func (this *HTTPAccessLogDAO) AccessStatistics(tx *dbs.Tx, day string, userId int64) (stats []*AccessStatistics, err error) {

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
	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}
	//
	//// 检查各个分表是否一致
	//var newTableQueries = []*accessLogTableQuery{}
	//for _, tableQuery := range tableQueries {
	//	if tableQuery.name != maxTableName {
	//		continue
	//	}
	//	newTableQueries = append(newTableQueries, tableQuery)
	//}
	//tableQueries = newTableQueries

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return
	}
	locker := sync.Mutex{}
	stats = make([]*AccessStatistics, 0)
	count := len(serverIds)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, serverId := range serverIds {
		go func(serverId int64) {
			defer wg.Done()

			dwg := &sync.WaitGroup{}
			dwg.Add(len(tableQueries))
			stat := &AccessStatistics{ServerId: serverId}
			//访问IP地址列表
			svrAddrStats1 := []*HTTPAccessLog{}
			// 攻击访问IP地址列表
			svrAddrStats2 := []*HTTPAccessLog{}
			for _, tableQuery := range tableQueries {
				go func(tableQuery *accessLogTableQuery, stat *AccessStatistics) {
					defer dwg.Done()

					var dao = tableQuery.daoWrapper.DAO
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					// 1. 统计该服务的服务总数
					accessTotal, err := dao.Query(tx).Attr("serverId", serverId).Debug(false).
						Reuse(false).
						Table(tableQuery.name).
						Count()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					// 2. 统计防护的总数
					attackTotal, err := dao.Query(tx).Attr("serverId", serverId).Debug(false).
						Reuse(false).Where("firewallPolicyId>0").UseIndex("firewallPolicyId").
						Table(tableQuery.name).
						Count()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					// 3. 访问IP列表
					var logs1 []*HTTPAccessLog
					_, err = dao.Query(tx).Debug(false).SQL(fmt.Sprintf("SELECT  `remoteAddr` FROM `%s`"+
						" WHERE `serverId`=%d GROUP BY `remoteAddr`", tableQuery.name, serverId)).Slice(&logs1).
						FindAll()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					// 4. 拦截IP列表
					var logs2 []*HTTPAccessLog
					_, err = dao.Query(tx).Debug(false).SQL(fmt.Sprintf("SELECT  `remoteAddr` FROM `%s` WHERE `serverId`=%d and `firewallPolicyId`>0 GROUP BY `remoteAddr`", tableQuery.name, serverId)).Slice(&logs2).
						FindAll()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					locker.Lock()
					stat.AccessTotal += accessTotal
					stat.AttackTotal += attackTotal
					svrAddrStats1 = append(svrAddrStats1, logs1...)
					svrAddrStats2 = append(svrAddrStats2, logs2...)
					locker.Unlock()
				}(tableQuery, stat)
			}
			dwg.Wait()
			// 访问IP地址列表 差重map
			fraddr1 := map[string]bool{}
			// 攻击IP地址列表 差重map
			fraddr2 := map[string]bool{}
			locker.Lock()
			for _, v := range svrAddrStats1 {
				if _, ok := fraddr1[v.RemoteAddr]; !ok {
					fraddr1[v.RemoteAddr] = true
					stat.AccessIpTotal++
				}
			}
			for _, v := range svrAddrStats2 {
				if _, ok := fraddr2[v.RemoteAddr]; !ok {
					fraddr2[v.RemoteAddr] = true
					stat.AttackIpTotal++
				}
			}
			stats = append(stats, stat)
			locker.Unlock()
		}(serverId)
	}
	wg.Wait()

	return stats, nil
}

// AttackURLTop 统计最受攻击的域名排行
func (this *HTTPAccessLogDAO) AttackURLTop(tx *dbs.Tx, day string, userId int64) (resp *AttackURLTop, err error) {
	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	serverIds := []int64{}
	resp = &AttackURLTop{Tops: make([]*AttackUrlTopItem1, 0)}
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
	hostStats := make([]*HTTPAccessLog, 0)
	uriStats := make([]*HTTPAccessLog, 0)
	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}
	//
	//// 检查各个分表是否一致
	//
	//var newTableQueries = []*accessLogTableQuery{}
	//for _, tableQuery := range tableQueries {
	//	if tableQuery.name != maxTableName {
	//		continue
	//	}
	//	newTableQueries = append(newTableQueries, tableQuery)
	//}
	//tableQueries = newTableQueries

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(serverIds))
	for _, serverId := range serverIds {

		l := sync.Mutex{}
		serverHostStats := make([]*HTTPAccessLog, 0)
		serverUriStats := make([]*HTTPAccessLog, 0)
		go func(serverId int64) {
			defer wg.Done()

			dwg := &sync.WaitGroup{}
			dwg.Add(len(tableQueries))
			for _, tableQuery := range tableQueries {
				go func(tableQuery *accessLogTableQuery) {
					defer dwg.Done()
					dao := tableQuery.daoWrapper.DAO
					// 开始查询 - 统计域名
					query := dao.Query(tx)

					// 条件
					query.Attr("serverId", serverId).
						Reuse(false)

					query.Where("firewallPolicyId>0")
					query.UseIndex("firewallPolicyId")

					var ones []*HTTPAccessLog

					// 开始查询
					_, err = query.
						Table(tableQuery.name).
						Group("host").
						Result("TRIM(BOTH '\"' FROM json_extract(content,'$.host')) AS host, COUNT(*) AS count").
						Slice(&ones).
						FindAll()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					l.Lock()
					serverHostStats = append(serverHostStats, ones...)
					l.Unlock()

					ones = make([]*HTTPAccessLog, 0)
					// 开始查询 - 统计URI
					query = dao.Query(tx)

					// 条件
					query.Attr("serverId", serverId).
						Reuse(false)

					query.Where("firewallPolicyId>0")
					query.UseIndex("firewallPolicyId")

					// 开始查询
					_, err = query.
						Table(tableQuery.name).
						Group("uri").
						Result("TRIM(BOTH '\"' FROM json_extract(content,'$.requestURI')) AS uri, COUNT(*) AS count").
						Slice(&ones).
						FindAll()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					l.Lock()
					serverUriStats = append(serverUriStats, ones...)
					l.Unlock()
				}(tableQuery)
			}
			dwg.Wait()
			mergeHostStats, hostCountStu := statisticsHostTop(serverHostStats)
			serverUriStats, uriCountStu := statisticsUriTop(serverUriStats)
			locker.Lock()
			hostStats = append(hostStats, mergeHostStats...)
			uriStats = append(uriStats, serverUriStats...)
			resp.Tops = append(resp.Tops, &AttackUrlTopItem1{ServerId: serverId, Hosts: hostCountStu, Uris: uriCountStu})
			locker.Unlock()
		}(serverId)
	}
	wg.Wait()
	_, hostCountStu := statisticsHostTop(hostStats)
	_, uriCountStu := statisticsUriTop(uriStats)
	resp.Tops = append(resp.Tops, &AttackUrlTopItem1{ServerId: 0, Hosts: hostCountStu, Uris: uriCountStu})
	return resp, nil
}
func statisticsUriTop(serverStats []*HTTPAccessLog) ([]*HTTPAccessLog, *ValueCount) {
	// 同服务下排序
	// 统计同域名下访护次数
	uriCountMaps := make(map[string]int64, 0)
	// 合并同域名的访护次数
	mergeStats := make([]*HTTPAccessLog, 0)
	// 排序
	uriCountStu := ValueCount{}
	for _, stat := range serverStats {
		uriCountMaps[stat.Uri] += stat.Count
	}
	for uri, count := range uriCountMaps {
		mergeStats = append(mergeStats, &HTTPAccessLog{Uri: uri, Count: count})
		uriCountStu.Value = append(uriCountStu.Value, uri)
		uriCountStu.Count = append(uriCountStu.Count, count)
	}
	// 排序
	sort.Sort(uriCountStu)

	min := len(uriCountStu.Value)
	if min > len(uriCountStu.Count) {
		min = len(uriCountStu.Count)
	}

	uriCountStu.Value = uriCountStu.Value[:min]
	uriCountStu.Count = uriCountStu.Count[:min]
	return mergeStats, &uriCountStu
}

func statisticsHostTop(serverStats []*HTTPAccessLog) ([]*HTTPAccessLog, *ValueCount) {
	// 同服务下排序
	// 统计同域名下访护次数
	hostCountMaps := make(map[string]int64, 0)
	// 合并同域名的访护次数
	mergeStats := make([]*HTTPAccessLog, 0)
	// 排序
	hostCountStu := ValueCount{}
	for _, stat := range serverStats {
		hostCountMaps[stat.Host] += stat.Count
	}
	for host, count := range hostCountMaps {
		mergeStats = append(mergeStats, &HTTPAccessLog{Host: host, Count: count})
		hostCountStu.Value = append(hostCountStu.Value, host)
		hostCountStu.Count = append(hostCountStu.Count, count)
	}
	// 排序
	sort.Sort(hostCountStu)

	min := len(hostCountStu.Value)
	if min > len(hostCountStu.Count) {
		min = len(hostCountStu.Count)
	}
	hostCountStu.Value = hostCountStu.Value[:min]
	hostCountStu.Count = hostCountStu.Count[:min]
	return mergeStats, &hostCountStu
}

// AccessIPTop 客户端访问IP排行
func (this *HTTPAccessLogDAO) AccessIPTop(tx *dbs.Tx, day string, userId int64, top int) (resp *AccessIPTop, err error) {
	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	serverIds := []int64{}
	resp = &AccessIPTop{Tops: make([]*AccessIPTopItem, 0)}
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
	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}
	//
	//// 检查各个分表是否一致
	//
	//var newTableQueries = []*accessLogTableQuery{}
	//for _, tableQuery := range tableQueries {
	//	if tableQuery.name != maxTableName {
	//		continue
	//	}
	//	newTableQueries = append(newTableQueries, tableQuery)
	//}
	//tableQueries = newTableQueries

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return
	}
	locker := sync.Mutex{}
	stats := make([]*HTTPAccessLog, 0)
	wg := &sync.WaitGroup{}
	wg.Add(len(serverIds))
	for _, serverId := range serverIds {

		go func(serverId int64) {
			defer wg.Done()

			l := sync.Mutex{}
			serverLogs := make([]*HTTPAccessLog, 0)

			dwg := &sync.WaitGroup{}
			dwg.Add(len(tableQueries))
			for _, tableQuery := range tableQueries {
				go func(tableQuery *accessLogTableQuery) {
					defer dwg.Done()
					dao := tableQuery.daoWrapper.DAO
					// 开始查询
					query := dao.Query(tx)

					// 条件
					query.Attr("serverId", serverId).
						Reuse(false)

					var ones []*HTTPAccessLog
					// 开始查询
					_, err = query.
						Table(tableQuery.name).
						Group("remoteAddr").
						Result("remoteAddr, COUNT(1) AS count").
						Slice(&ones).
						FindAll()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					l.Lock()
					serverLogs = append(serverLogs, ones...)
					l.Unlock()
				}(tableQuery)
			}
			dwg.Wait()
			mergeStats, ipCountStu := statisticsAccessIpTop(serverLogs, top)
			locker.Lock()
			stats = append(stats, mergeStats...)
			resp.Tops = append(resp.Tops, &AccessIPTopItem{ServerId: serverId, Counts: ipCountStu.Count, IPs: ipCountStu.IP})
			locker.Unlock()
		}(serverId)
	}
	wg.Wait()

	_, ipCountStu := statisticsAccessIpTop(stats, top)
	resp.Tops = append(resp.Tops, &AccessIPTopItem{ServerId: 0, Counts: ipCountStu.Count, IPs: ipCountStu.IP})
	return resp, nil
}
func statisticsAccessIpTop(serverStats []*HTTPAccessLog, top int) ([]*HTTPAccessLog, *IpCount) {
	// 同服务下排序
	// 统计同域名下访护次数
	urlCountMaps := make(map[string]int64, 0)
	// 合并同域名的访护次数
	mergeStats := make([]*HTTPAccessLog, 0)
	// 排序
	ipCountStu := IpCount{}
	for _, stat := range serverStats {
		urlCountMaps[stat.RemoteAddr] += stat.Count
	}
	for url, count := range urlCountMaps {
		mergeStats = append(mergeStats, &HTTPAccessLog{RemoteAddr: url, Count: count})
		ipCountStu.IP = append(ipCountStu.IP, url)
		ipCountStu.Count = append(ipCountStu.Count, count)
	}
	// 排序
	sort.Sort(ipCountStu)

	min := len(ipCountStu.IP)
	if min > len(ipCountStu.Count) {
		min = len(ipCountStu.Count)
	}
	if min > top {
		min = top
	}
	ipCountStu.IP = ipCountStu.IP[:min]
	ipCountStu.Count = ipCountStu.Count[:min]
	return mergeStats, &ipCountStu
}
func statisticsStatusCodeTop(serverStats []*HTTPAccessLog) ([]*HTTPAccessLog, *CodeCount) {
	// 同服务下排序
	// 统计同域名下访护次数
	codeCountMaps := make(map[uint32]int64, 0)
	// 合并同域名的访护次数
	mergeStats := make([]*HTTPAccessLog, 0)
	// 排序
	codeCountStu := CodeCount{}
	for _, stat := range serverStats {
		codeCountMaps[stat.Status] += stat.Count
	}
	for code, count := range codeCountMaps {
		mergeStats = append(mergeStats, &HTTPAccessLog{Status: code, Count: count})
		codeCountStu.Code = append(codeCountStu.Code, code)
		codeCountStu.Count = append(codeCountStu.Count, count)
	}

	min := len(codeCountStu.Code)
	if min > len(codeCountStu.Count) {
		min = len(codeCountStu.Count)
	}

	codeCountStu.Code = codeCountStu.Code[:min]
	codeCountStu.Count = codeCountStu.Count[:min]
	return mergeStats, &codeCountStu
}

// StatusCodeStatistics 客户端访问IP排行
func (this *HTTPAccessLogDAO) StatusCodeStatistics(tx *dbs.Tx, day string, userId int64) (resp *StatusCodeTop, err error) {
	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	serverIds := []int64{}
	resp = &StatusCodeTop{Tops: make([]*StatusCodeTopItem, 0)}
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
	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	//var maxTableName = ""
	//for _, daoWrapper := range daoList {
	//	var instance = daoWrapper.DAO.Instance
	//	def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, 0)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if !def.Exists {
	//		continue
	//	}
	//
	//	if len(maxTableName) == 0 || def.Name > maxTableName {
	//		maxTableName = def.Name
	//	}
	//
	//	tableQueries = append(tableQueries, &accessLogTableQuery{
	//		daoWrapper:         daoWrapper,
	//		name:               def.Name,
	//		hasRemoteAddrField: def.HasRemoteAddr,
	//		hasDomainField:     def.HasDomain,
	//	})
	//}
	//
	//// 检查各个分表是否一致
	//
	//var newTableQueries = []*accessLogTableQuery{}
	//for _, tableQuery := range tableQueries {
	//	if tableQuery.name != maxTableName {
	//		continue
	//	}
	//	newTableQueries = append(newTableQueries, tableQuery)
	//}
	//tableQueries = newTableQueries

	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		defs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}

		if len(defs) > 0 {
			for _, def := range defs {
				tableQueries = append(tableQueries, &accessLogTableQuery{
					daoWrapper:         daoWrapper,
					name:               def.Name,
					hasRemoteAddrField: def.HasRemoteAddr,
					hasDomainField:     def.HasDomain,
				})
			}
		}

	}

	if len(tableQueries) == 0 {
		return
	}
	locker := sync.Mutex{}
	stats := make([]*HTTPAccessLog, 0)
	wg := &sync.WaitGroup{}
	wg.Add(len(serverIds))
	for _, serverId := range serverIds {

		go func(serverId int64) {
			defer wg.Done()

			l := sync.Mutex{}
			serverLogs := make([]*HTTPAccessLog, 0)

			dwg := &sync.WaitGroup{}
			dwg.Add(len(tableQueries))
			for _, tableQuery := range tableQueries {
				go func(tableQuery *accessLogTableQuery) {
					defer dwg.Done()
					// 开始查询
					query := tableQuery.daoWrapper.DAO.Query(tx)

					// 条件
					query.Attr("serverId", serverId).
						Reuse(false)

					var ones []*HTTPAccessLog
					// 开始查询
					_, err = query.
						Table(tableQuery.name).
						Group("status").
						Result("status, COUNT(1) AS count").
						Slice(&ones).
						FindAll()
					if err != nil {
						logs.Println("[DB_NODE]" + err.Error())
						return
					}
					l.Lock()
					serverLogs = append(serverLogs, ones...)
					l.Unlock()
				}(tableQuery)
			}
			dwg.Wait()
			mergeStats, codeCountStu := statisticsStatusCodeTop(serverLogs)
			locker.Lock()
			stats = append(stats, mergeStats...)
			resp.Tops = append(resp.Tops, &StatusCodeTopItem{ServerId: serverId, Counts: codeCountStu.Count, Codes: codeCountStu.Code})
			locker.Unlock()
		}(serverId)
	}
	wg.Wait()

	_, ipCountStu := statisticsStatusCodeTop(stats)
	resp.Tops = append(resp.Tops, &StatusCodeTopItem{ServerId: 0, Counts: ipCountStu.Count, Codes: ipCountStu.Code})
	return resp, nil
}

type AttackType struct {
	ServerId int64   `json:"serverId"` //所属服务ID
	Code     string  `json:"code"`     //攻击code
	Name     string  `json:"name"`     //攻击名称
	Count    int64   `json:"count"`    //条数
	ids      []int64 //对应策略分组id
}
type AccessStatistics struct {
	ServerId      int64 `json:"serverId"`      // 服务ID
	AccessTotal   int64 `json:"accessTotal"`   // 访问总数
	AttackTotal   int64 `json:"attackTotal"`   // 防护总数
	AccessIpTotal int64 `json:"accessIpTotal"` // 访问IP总数
	AttackIpTotal int64 `json:"attackIpTotal"` // 拦截IP总数
}

type AttackURLTop struct {
	Tops []*AttackUrlTopItem1 `json:"tops"`
}
type AttackUrlTopItem struct {
	ServerId int64  `json:"server_id"`
	Url      string `json:"url"`
	Count    int64  `json:"count"`
}
type AttackUrlTopItem1 struct {
	ServerId int64       `json:"server_id"`
	Uris     *ValueCount `json:"uris"`
	Hosts    *ValueCount `json:"hosts"`
}

type AccessIPTop struct {
	Tops []*AccessIPTopItem `json:"tops"`
}
type StatusCodeTop struct {
	Tops []*StatusCodeTopItem `json:"tops"`
}

type AccessIPTopItem struct {
	ServerId int64    `json:"server_id"`
	IPs      []string `json:"url"`
	Counts   []int64  `json:"count"`
}
type StatusCodeTopItem struct {
	ServerId int64    `json:"server_id"`
	Codes    []uint32 `json:"codes"`
	Counts   []int64  `json:"count"`
}
type ValueCount struct {
	Value []string
	Count []int64
}

func (this ValueCount) Len() int {
	min := len(this.Count)
	if min > len(this.Value) {
		return len(this.Value)
	}
	return min
}
func (this ValueCount) Less(i, j int) bool {
	return this.Count[i] > this.Count[j]
}
func (this ValueCount) Swap(i, j int) {
	this.Value[i], this.Value[j] = this.Value[j], this.Value[i]
	this.Count[i], this.Count[j] = this.Count[j], this.Count[i]
}
