package services

import (
	"context"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/configutils"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/lists"
	"net"
	"strings"
	"time"
)

// IPListService IP名单相关服务
type IPListService struct {
	BaseService
}

// CreateIPList 创建IP列表
func (this *IPListService) CreateIPList(ctx context.Context, req *pb.CreateIPListRequest) (*pb.CreateIPListResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户相关信息
	if userId > 0 {
		// 检查服务ID
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	}

	listId, err := models.SharedIPListDAO.CreateIPList(tx, userId, req.ServerId, req.Type, req.Name, req.Code, req.TimeoutJSON, req.Description, req.IsPublic, req.IsGlobal)
	if err != nil {
		return nil, err
	}
	return &pb.CreateIPListResponse{IpListId: listId}, nil
}

// UpdateIPList 修改IP列表
func (this *IPListService) UpdateIPList(ctx context.Context, req *pb.UpdateIPListRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedIPListDAO.UpdateIPList(tx, req.IpListId, req.Name, req.Code, req.TimeoutJSON, req.Description)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledIPList 查找IP列表
func (this *IPListService) FindEnabledIPList(ctx context.Context, req *pb.FindEnabledIPListRequest) (*pb.FindEnabledIPListResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedIPListDAO.CheckUserIPList(tx, userId, req.IpListId)
		if err != nil {
			return nil, err
		}
	}

	list, err := models.SharedIPListDAO.FindEnabledIPList(tx, req.IpListId, nil)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return &pb.FindEnabledIPListResponse{IpList: nil}, nil
	}
	return &pb.FindEnabledIPListResponse{IpList: &pb.IPList{
		Id:          int64(list.Id),
		IsOn:        list.IsOn,
		Type:        list.Type,
		Name:        list.Name,
		Code:        list.Code,
		TimeoutJSON: list.Timeout,
		Description: list.Description,
		IsGlobal:    list.IsGlobal,
	}}, nil
}

// CountAllEnabledIPLists 计算名单数量
func (this *IPListService) CountAllEnabledIPLists(ctx context.Context, req *pb.CountAllEnabledIPListsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedIPListDAO.CountAllEnabledIPLists(tx, req.Type, req.IsPublic, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledIPLists 列出单页名单
func (this *IPListService) ListEnabledIPLists(ctx context.Context, req *pb.ListEnabledIPListsRequest) (*pb.ListEnabledIPListsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	ipLists, err := models.SharedIPListDAO.ListEnabledIPLists(tx, req.Type, req.IsPublic, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbLists []*pb.IPList
	for _, list := range ipLists {
		pbLists = append(pbLists, &pb.IPList{
			Id:          int64(list.Id),
			IsOn:        list.IsOn,
			Type:        list.Type,
			Name:        list.Name,
			Code:        list.Code,
			TimeoutJSON: list.Timeout,
			IsPublic:    list.IsPublic,
			Description: list.Description,
			IsGlobal:    list.IsGlobal,
		})
	}
	return &pb.ListEnabledIPListsResponse{IpLists: pbLists}, nil
}

// DeleteIPList 删除IP名单
func (this *IPListService) DeleteIPList(ctx context.Context, req *pb.DeleteIPListRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedIPListDAO.DisableIPList(tx, req.IpListId)
	if err != nil {
		return nil, err
	}

	// 删除所有IP
	err = models.SharedIPItemDAO.DisableIPItemsWithListId(tx, req.IpListId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ExistsEnabledIPList 检查IPList是否存在
func (this *IPListService) ExistsEnabledIPList(ctx context.Context, req *pb.ExistsEnabledIPListRequest) (*pb.ExistsEnabledIPListResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	b, err := models.SharedIPListDAO.ExistsEnabledIPList(tx, req.IpListId)
	if err != nil {
		return nil, err
	}
	return &pb.ExistsEnabledIPListResponse{Exists: b}, nil
}

// FindEnabledIPListContainsIP 根据IP来搜索IP名单
func (this *IPListService) FindEnabledIPListContainsIP(ctx context.Context, req *pb.FindEnabledIPListContainsIPRequest) (*pb.FindEnabledIPListContainsIPResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	items, err := models.SharedIPItemDAO.FindEnabledItemsWithIP(tx, req.Ip)
	if err != nil {
		return nil, err
	}

	var pbLists = []*pb.IPList{}
	var listIds = []int64{}
	var cacheMap = utils.NewCacheMap()
	for _, item := range items {
		if lists.ContainsInt64(listIds, int64(item.ListId)) {
			continue
		}

		list, err := models.SharedIPListDAO.FindEnabledIPList(tx, int64(item.ListId), cacheMap)
		if err != nil {
			return nil, err
		}
		if list == nil {
			continue
		}
		if !list.IsPublic {
			continue
		}
		pbLists = append(pbLists, &pb.IPList{
			Id:          int64(list.Id),
			IsOn:        list.IsOn,
			Type:        list.Type,
			Name:        list.Name,
			Code:        list.Code,
			IsPublic:    list.IsPublic,
			IsGlobal:    list.IsGlobal,
			Description: "",
		})

		listIds = append(listIds, int64(item.ListId))
	}
	return &pb.FindEnabledIPListContainsIPResponse{IpLists: pbLists}, nil
}

// ------- api 客户定制化接口

var eventLevels = []string{"debug", "notice", "warning", "error", "critical", "fatal"}

type DeleteBlackIpRequest struct {
	GroupId int64    `json:"groupId"` //所属黑白名单分组id 非必填， 默认将ip写入到全局的黑名单分组中
	Ips     []string `json:"ips"`     //ipv6添加单个ip ：1406:3c00:0:2409:13:58:103:15 ipv4添加单个ip：192.168.1.2
}

type CreateBlackIPRequest struct {
	GroupId int64  `protobuf:"varint,1,opt,name=groupId,proto3" json:"groupId,omitempty"` //所属黑白名单分组id 非必填， 默认将ip写入到全局的黑名单分组中
	Type    string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`        //ip类型 ipv4 ipv6 all
	Level   int32  `protobuf:"varint,3,opt,name=level,proto3" json:"level,omitempty"`     //级别 0/1/2/3/4/5 调试、通知、告警、错误、严重、致命
	Day     int32  `protobuf:"varint,4,opt,name=day,proto3" json:"day,omitempty"`         //过去时间：天数 、  默认 0 表示永久
	Desc    string `protobuf:"bytes,5,opt,name=desc,proto3" json:"desc,omitempty"`        // 备注
	Ip      string `protobuf:"bytes,6,opt,name=ip,proto3" json:"ip,omitempty"`            //ipv6添加单个ip ：1406:3c00:0:2409:13:58:103:15 ipv4添加单个ip：192.168.1.2
	Ips     string `protobuf:"bytes,7,opt,name=ips,proto3" json:"ips,omitempty"`          //批量添加ip列表 用逗号分隔，支持三种格式：192.168.1.2、192.168.1.1/244、192.168.1.1-192.168.1.255
	Method  string `protobuf:"bytes,8,opt,name=method,proto3" json:"method,omitempty"`    //新增方式  single/batch 单个添加/批量添加 默认 单个添加
}

// CreateBlackIP 新增全局黑名单
func (this *IPListService) CreateBlackIP(ctx context.Context, req *CreateBlackIPRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if req.GroupId != 0 {
		exist, err := models.SharedIPListDAO.ExistsEnabledIPList(tx, req.GroupId)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, fmt.Errorf("该黑白名单分组[%d]不存在", req.GroupId)
		}
	} else {
		req.GroupId, err = models.SharedIPListDAO.FindOrCreateGlobalBlackIPList(tx)
		if err != nil {
			return nil, err
		}
	}
	if req.Level < 0 || int(req.Level) > len(eventLevels) {

		return nil, fmt.Errorf("错误的事件级别%d", req.Level)
	}

	type ipData struct {
		ipFrom string
		ipTo   string
	}

	var batchIPs = []*ipData{}
	switch req.Type {
	case "ipv4":
		if req.Method == "single" {
			if req.Ip == "" {
				return nil, fmt.Errorf("请输入IP")
			}
			if !utils.IsIPv4(req.Ip) {
				return nil, fmt.Errorf("请输入正确的IP")
			}
		} else if req.Method == "batch" {
			if len(req.Ips) == 0 {
				return nil, fmt.Errorf("请输入IP列表")
			}
			var lines = strings.Split(req.Ips, ",")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "/") { // CIDR
					if strings.Contains(line, ":") {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					ipFrom, ipTo, err := configutils.ParseCIDR(line)
					if err != nil {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
						ipTo:   ipTo,
					})
				} else if strings.Contains(line, "-") { // IP Range
					var pieces = strings.Split(line, "-")
					var ipFrom = strings.TrimSpace(pieces[0])
					var ipTo = strings.TrimSpace(pieces[1])

					if net.ParseIP(ipFrom) == nil || net.ParseIP(ipTo) == nil || strings.Contains(ipFrom, ":") || strings.Contains(ipTo, ":") {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					if utils.IP2Long(ipFrom) > utils.IP2Long(ipTo) {
						ipFrom, ipTo = ipTo, ipFrom
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
						ipTo:   ipTo,
					})
				} else if strings.Contains(line, ",") { // IP Range
					var pieces = strings.Split(line, ",")
					var ipFrom = strings.TrimSpace(pieces[0])
					var ipTo = strings.TrimSpace(pieces[1])

					if net.ParseIP(ipFrom) == nil || net.ParseIP(ipTo) == nil || strings.Contains(ipFrom, ":") || strings.Contains(ipTo, ":") {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					if utils.IP2Long(ipFrom) > utils.IP2Long(ipTo) {
						ipFrom, ipTo = ipTo, ipFrom
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
						ipTo:   ipTo,
					})
				} else if len(line) > 0 {
					var ipFrom = line
					if net.ParseIP(ipFrom) == nil || strings.Contains(ipFrom, ":") {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
					})
				}
			}
		}
	case "ipv6":
		if req.Method == "single" {
			if req.Ip == "" {
				return nil, fmt.Errorf("请输入IP")
			}

			// 校验IP格式（ipFrom）
			if !utils.IsIPv6(req.Ip) {
				return nil, fmt.Errorf("请输入正确的IPv6地址")
			}
		} else if req.Method == "batch" {
			if len(req.Ips) == 0 {
				return nil, fmt.Errorf("请输入IP列表")
			}
			var lines = strings.Split(req.Ips, ",")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "/") { // CIDR
					if !strings.Contains(line, ":") {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					ipFrom, ipTo, err := configutils.ParseCIDR(line)
					if err != nil {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
						ipTo:   ipTo,
					})
				} else if strings.Contains(line, "-") { // IP Range
					var pieces = strings.Split(line, "-")
					var ipFrom = strings.TrimSpace(pieces[0])
					var ipTo = strings.TrimSpace(pieces[1])

					if net.ParseIP(ipFrom) == nil || net.ParseIP(ipTo) == nil || !strings.Contains(ipFrom, ":") || !strings.Contains(ipTo, ":") {
						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					if utils.IP2Long(ipFrom) > utils.IP2Long(ipTo) {
						ipFrom, ipTo = ipTo, ipFrom
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
						ipTo:   ipTo,
					})
				} else if strings.Contains(line, ",") { // IP Range
					var pieces = strings.Split(line, ",")
					var ipFrom = strings.TrimSpace(pieces[0])
					var ipTo = strings.TrimSpace(pieces[1])

					if net.ParseIP(ipFrom) == nil || net.ParseIP(ipTo) == nil || !strings.Contains(ipFrom, ":") || !strings.Contains(ipTo, ":") {

						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					if utils.IP2Long(ipFrom) > utils.IP2Long(ipTo) {
						ipFrom, ipTo = ipTo, ipFrom
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
						ipTo:   ipTo,
					})
				} else if len(line) > 0 {
					var ipFrom = line
					if net.ParseIP(ipFrom) == nil || !strings.Contains(ipFrom, ":") {

						return nil, fmt.Errorf("\"%s\"IP格式错误", line)
					}
					batchIPs = append(batchIPs, &ipData{
						ipFrom: ipFrom,
					})
				}
			}
		}
	case "all":
		req.Ip = "0.0.0.0"
	default:
		return nil, fmt.Errorf("无效的ip类型")
	}
	var expiredAt int64
	if req.Day > 0 {
		expiredAt = time.Now().AddDate(0, 0, int(req.Day)).Unix()
	}

	if len(batchIPs) > 0 {
		for _, ip := range batchIPs {

			// 删除以前的
			err = models.SharedIPItemDAO.DeleteOldItem(tx, req.GroupId, ip.ipFrom, ip.ipTo)
			if err != nil {
				return nil, err
			}
			_, err = models.SharedIPItemDAO.CreateIPItem(tx, req.GroupId, ip.ipFrom, ip.ipTo, expiredAt, req.Desc,
				req.Type, eventLevels[req.Level], 0, 0, 0,
				0, 0, 0, 0)
			if err != nil {
				return nil, err
			}
		}

	} else {

		// 删除以前的
		err = models.SharedIPItemDAO.DeleteOldItem(tx, req.GroupId, req.Ip, "")
		if err != nil {
			return nil, err
		}

		_, err = models.SharedIPItemDAO.CreateIPItem(tx,
			req.GroupId,
			req.Ip,
			"",
			expiredAt,
			req.Desc,
			req.Type,
			eventLevels[req.Level], 0, 0, 0,
			0, 0, 0, 0)
		if err != nil {
			return nil, err
		}
	}
	return this.Success()
}

func (this *IPListService) ClearBlackIP(ctx context.Context, req *DeleteBlackIpRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if req.GroupId == 0 {
		return nil, errors.New("分组ID不能为空")
	}
	err = models.SharedIPItemDAO.DisableIPItemsWithListId(tx, req.GroupId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
func (this *IPListService) DeleteBlackIP(ctx context.Context, req *DeleteBlackIpRequest) (*pb.RPCSuccess, error) {

	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if req.GroupId == 0 {
		return nil, errors.New("分组ID不能为空")
	}
	for _, ip := range req.Ips {
		ipt := utils.IP2Long(ip)
		if ipt == 0 {
			return nil, errors.New("IP格式错误")
		}
		ipitem, err := models.SharedIPItemDAO.FindEnabledItemContainsIP(tx, req.GroupId, ipt)
		if err != nil {
			return nil, err
		}
		if ipitem == nil {
			continue
		}
		err = models.SharedIPItemDAO.DisableIPItem(tx, int64(ipitem.Id))
		if err != nil {
			return nil, err
		}
	}
	return this.Success()
}

type ListIPWithListIdResponse struct {
	Total   int64        `json:"total"`
	IpItems []*pb.IPItem `protobuf:"bytes,1,rep,name=ipItems,proto3" json:"ipItems,omitempty"`
}

// ListIPWithListId 查询指定黑白名单中的IP
func (this *IPListService) ListIPWithListId(ctx context.Context, req *pb.ListIPItemsWithListIdRequest) (*ListIPWithListIdResponse, error) {

	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedIPItemDAO.CountIPItemsWithListId(tx, req.IpListId, "", "", req.Keyword, req.EventLevel)
	if err != nil {
		return nil, err
	}

	items, err := models.SharedIPItemDAO.ListIPItemsWithListId(tx, req.IpListId, req.Keyword, "", "", req.EventLevel, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.IPItem{}
	for _, item := range items {
		if len(item.Type) == 0 {
			item.Type = models.IPItemTypeIPv4
		}

		// server
		var pbSourceServer *pb.Server
		if item.SourceServerId > 0 {
			serverName, err := models.SharedServerDAO.FindEnabledServerName(tx, int64(item.SourceServerId))
			if err != nil {
				return nil, err
			}
			pbSourceServer = &pb.Server{
				Id:   int64(item.SourceServerId),
				Name: serverName,
			}
		}

		// WAF策略
		var pbSourcePolicy *pb.HTTPFirewallPolicy
		if item.SourceHTTPFirewallPolicyId > 0 {
			policy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledHTTPFirewallPolicyBasic(tx, int64(item.SourceHTTPFirewallPolicyId))
			if err != nil {
				return nil, err
			}
			if policy != nil {
				pbSourcePolicy = &pb.HTTPFirewallPolicy{
					Id:       int64(item.SourceHTTPFirewallPolicyId),
					Name:     policy.Name,
					ServerId: int64(policy.ServerId),
				}
			}
		}

		// WAF分组
		var pbSourceGroup *pb.HTTPFirewallRuleGroup
		if item.SourceHTTPFirewallRuleGroupId > 0 {
			groupName, err := models.SharedHTTPFirewallRuleGroupDAO.FindHTTPFirewallRuleGroupName(tx, int64(item.SourceHTTPFirewallRuleGroupId))
			if err != nil {
				return nil, err
			}
			pbSourceGroup = &pb.HTTPFirewallRuleGroup{
				Id:   int64(item.SourceHTTPFirewallRuleGroupId),
				Name: groupName,
			}
		}

		// WAF规则集
		var pbSourceSet *pb.HTTPFirewallRuleSet
		if item.SourceHTTPFirewallRuleSetId > 0 {
			setName, err := models.SharedHTTPFirewallRuleSetDAO.FindHTTPFirewallRuleSetName(tx, int64(item.SourceHTTPFirewallRuleSetId))
			if err != nil {
				return nil, err
			}
			pbSourceSet = &pb.HTTPFirewallRuleSet{
				Id:   int64(item.SourceHTTPFirewallRuleSetId),
				Name: setName,
			}
		}

		result = append(result, &pb.IPItem{
			Id:                            int64(item.Id),
			IpFrom:                        item.IpFrom,
			IpTo:                          item.IpTo,
			Version:                       int64(item.Version),
			CreatedAt:                     int64(item.CreatedAt),
			ExpiredAt:                     int64(item.ExpiredAt),
			Reason:                        item.Reason,
			Type:                          item.Type,
			EventLevel:                    item.EventLevel,
			NodeId:                        int64(item.NodeId),
			ServerId:                      int64(item.ServerId),
			SourceNodeId:                  int64(item.SourceNodeId),
			SourceServerId:                int64(item.SourceServerId),
			SourceHTTPFirewallPolicyId:    int64(item.SourceHTTPFirewallPolicyId),
			SourceHTTPFirewallRuleGroupId: int64(item.SourceHTTPFirewallRuleGroupId),
			SourceHTTPFirewallRuleSetId:   int64(item.SourceHTTPFirewallRuleSetId),
			SourceServer:                  pbSourceServer,
			SourceHTTPFirewallPolicy:      pbSourcePolicy,
			SourceHTTPFirewallRuleGroup:   pbSourceGroup,
			SourceHTTPFirewallRuleSet:     pbSourceSet,
			IsRead:                        item.IsRead,
		})
	}

	return &ListIPWithListIdResponse{IpItems: result, Total: count}, nil
}
