// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus
// +build plus

package nameservers

import (
	"context"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"strings"
	"time"
)

// NSDomainService 域名相关服务
type NSDomainService struct {
	services.BaseService
}

// CreateNSDomain 创建域名
func (this *NSDomainService) CreateNSDomain(ctx context.Context, req *pb.CreateNSDomainRequest) (*pb.CreateNSDomainResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	var isAdminRequest = adminId > 0

	if userId > 0 {
		req.UserId = userId
	}

	req.Name = strings.ToLower(req.Name)

	// 检查 req.NsDomainGroupIds 有效性
	var tx = this.NullTx()
	if req.UserId > 0 && len(req.NsDomainGroupIds) > 0 {
		for _, groupId := range req.NsDomainGroupIds {
			err = nameservers.SharedNSDomainGroupDAO.CheckUserGroup(tx, req.UserId, groupId)
			if err != nil {
				return nil, err
			}
		}
	}

	// 检查clusterId
	if req.UserId > 0 && !isAdminRequest {
		userConfig, err := models.SharedSysSettingDAO.ReadNSUserConfig(tx)
		if err != nil {
			return nil, err
		}
		req.NsClusterId = userConfig.DefaultClusterId
	}

	if req.NsClusterId <= 0 {
		return nil, errors.New("'nsClusterId' required")
	}

	// 是否已经存在
	exists, err := nameservers.SharedNSDomainDAO.ExistUserDomain(tx, req.UserId, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("domain '" + req.Name + "' already exists")
	}

	exists, err = nameservers.SharedNSDomainDAO.ExistVerifiedDomain(tx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("domain '" + req.Name + "' already exists")
	}

	// 状态
	var status = dnsconfigs.NSDomainStatusNone
	if isAdminRequest {
		// 管理员添加的直接通过审核
		status = dnsconfigs.NSDomainStatusVerified
	}

	// 创建
	domainId, err := nameservers.SharedNSDomainDAO.CreateDomain(tx, req.NsClusterId, req.UserId, req.NsDomainGroupIds, req.Name, status)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSDomainResponse{NsDomainId: domainId}, nil
}

// CreateNSDomains 批量创建域名
func (this *NSDomainService) CreateNSDomains(ctx context.Context, req *pb.CreateNSDomainsRequest) (*pb.CreateNSDomainsResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	var isAdminRequest = adminId > 0

	if userId > 0 {
		req.UserId = userId
	}

	// 检查分组有效性
	var tx = this.NullTx()
	if req.UserId > 0 {
		for _, groupId := range req.NsDomainGroupIds {
			err = nameservers.SharedNSDomainGroupDAO.CheckUserGroup(tx, req.UserId, groupId)
			if err != nil {
				return nil, err
			}
		}
	}

	// 检查集群
	if req.UserId > 0 && !isAdminRequest {
		userConfig, err := models.SharedSysSettingDAO.ReadNSUserConfig(tx)
		if err != nil {
			return nil, err
		}
		req.NsClusterId = userConfig.DefaultClusterId
	}

	if req.NsClusterId <= 0 {
		return nil, errors.New("'nsClusterId' required")
	}

	var domainIds = []int64{}
	var domainMap = map[string]bool{} // domainName => bool
	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, name := range req.Names {
			name = strings.ToLower(name)

			// 检查是否已添加
			_, ok := domainMap[name]
			if ok {
				return nil
			}

			domainMap[name] = true

			// 是否已经存在
			exists, err := nameservers.SharedNSDomainDAO.ExistUserDomain(tx, req.UserId, name)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("domain '" + name + "' already exists")
			}

			exists, err = nameservers.SharedNSDomainDAO.ExistVerifiedDomain(tx, name)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("domain '" + name + "' already exists")
			}

			// 状态
			var status = dnsconfigs.NSDomainStatusNone
			if isAdminRequest {
				// 管理员添加的直接通过验证
				status = dnsconfigs.NSDomainStatusVerified
			}

			// 创建
			domainId, err := nameservers.SharedNSDomainDAO.CreateDomain(tx, req.NsClusterId, req.UserId, req.NsDomainGroupIds, name, status)
			if err != nil {
				return err
			}
			domainIds = append(domainIds, domainId)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateNSDomainsResponse{NsDomainIds: domainIds}, nil
}

// UpdateNSDomain 修改域名
func (this *NSDomainService) UpdateNSDomain(ctx context.Context, req *pb.UpdateNSDomainRequest) (*pb.RPCSuccess, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	if req.UserId > 0 {
		//  检查分组有效性
		for _, groupId := range req.NsDomainGroupIds {
			err = nameservers.SharedNSDomainGroupDAO.CheckUserGroup(tx, req.UserId, groupId)
			if err != nil {
				return nil, err
			}
		}
	}

	err = nameservers.SharedNSDomainDAO.UpdateDomain(tx, req.NsDomainId, req.NsClusterId, req.UserId, req.NsDomainGroupIds, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateNSDomainStatus 修改域名状态
func (this *NSDomainService) UpdateNSDomainStatus(ctx context.Context, req *pb.UpdateNSDomainStatusRequest) (*pb.RPCSuccess, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.NsDomainId <= 0 {
		return nil, errors.New("invalid nsDomainId")
	}

	if !dnsconfigs.NSDomainStatusIsValid(req.Status) {
		return nil, errors.New("invalid status '" + req.Status + "'")
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSDomainDAO.UpdateDomainStatus(tx, req.NsDomainId, req.Status)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteNSDomain 删除域名
func (this *NSDomainService) DeleteNSDomain(ctx context.Context, req *pb.DeleteNSDomainRequest) (*pb.RPCSuccess, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSDomainDAO.DisableNSDomain(tx, req.NsDomainId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSDomains 批量删除域名
func (this *NSDomainService) DeleteNSDomains(ctx context.Context, req *pb.DeleteNSDomainsRequest) (*pb.RPCSuccess, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	for _, name := range req.Names {
		name = strings.ToLower(strings.TrimSpace(name))
		if len(name) == 0 {
			continue
		}

		if req.UserId > 0 {
			err = nameservers.SharedNSDomainDAO.DisableUserDomainWithName(tx, req.UserId, name)
		} else {
			err = nameservers.SharedNSDomainDAO.DisableDomainWithName(tx, name)
		}
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// FindNSDomain 查找单个域名
func (this *NSDomainService) FindNSDomain(ctx context.Context, req *pb.FindNSDomainRequest) (*pb.FindNSDomainResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, req.NsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindNSDomainResponse{NsDomain: nil}, nil
	}

	// 集群
	cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(domain.ClusterId))
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return &pb.FindNSDomainResponse{NsDomain: nil}, nil
	}

	// 用户
	var pbUser *pb.User
	if domain.UserId > 0 {
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(domain.UserId), nil)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return &pb.FindNSDomainResponse{NsDomain: nil}, nil
		}
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	// groups
	var pbGroups = []*pb.NSDomainGroup{}
	var groupIds = domain.DecodeGroupIds()
	for _, groupId := range groupIds {
		group, err := nameservers.SharedNSDomainGroupDAO.FindEnabledNSDomainGroup(tx, groupId)
		if err != nil {
			return nil, err
		}
		if group != nil && group.IsOn {
			pbGroups = append(pbGroups, &pb.NSDomainGroup{
				Id:     int64(group.Id),
				Name:   group.Name,
				IsOn:   group.IsOn,
				UserId: int64(group.UserId),
			})
		}
	}

	return &pb.FindNSDomainResponse{
		NsDomain: &pb.NSDomain{
			Id:        int64(domain.Id),
			Name:      domain.Name,
			IsOn:      domain.IsOn,
			TsigJSON:  domain.Tsig,
			CreatedAt: int64(domain.CreatedAt),
			Status:    domain.Status,
			NsCluster: &pb.NSCluster{
				Id:   int64(cluster.Id),
				IsOn: cluster.IsOn,
				Name: cluster.Name,
			},
			User:             pbUser,
			NsDomainGroupIds: groupIds,
			NsDomainGroups:   pbGroups,
		},
	}, nil
}

// FindNSDomainWithName 根据域名名称查找域名
func (this *NSDomainService) FindNSDomainWithName(ctx context.Context, req *pb.FindNSDomainWithNameRequest) (*pb.FindNSDomainWithNameResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	var isUserRequest = userId > 0

	if len(req.Name) == 0 {
		return &pb.FindNSDomainWithNameResponse{
			NsDomain: nil,
		}, nil
	}

	var tx = this.NullTx()
	domain, err := nameservers.SharedNSDomainDAO.FindDomainWithName(tx, userId, req.Name)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindNSDomainWithNameResponse{NsDomain: nil}, nil
	}

	// 集群
	var pbCluster *pb.NSCluster
	if !isUserRequest {
		cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(domain.ClusterId))
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			return &pb.FindNSDomainWithNameResponse{NsDomain: nil}, nil
		}
		pbCluster = &pb.NSCluster{
			Id:   int64(cluster.Id),
			IsOn: cluster.IsOn,
			Name: cluster.Name,
		}
	}

	// 用户
	var pbUser *pb.User
	if domain.UserId > 0 && !isUserRequest {
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(domain.UserId), nil)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return &pb.FindNSDomainWithNameResponse{NsDomain: nil}, nil
		}
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	// groups
	var pbGroups = []*pb.NSDomainGroup{}
	var groupIds = domain.DecodeGroupIds()
	for _, groupId := range groupIds {
		group, err := nameservers.SharedNSDomainGroupDAO.FindEnabledNSDomainGroup(tx, groupId)
		if err != nil {
			return nil, err
		}
		if group != nil && group.IsOn {
			pbGroups = append(pbGroups, &pb.NSDomainGroup{
				Id:     int64(group.Id),
				Name:   group.Name,
				IsOn:   group.IsOn,
				UserId: int64(group.UserId),
			})
		}
	}

	return &pb.FindNSDomainWithNameResponse{
		NsDomain: &pb.NSDomain{
			Id:               int64(domain.Id),
			Name:             domain.Name,
			IsOn:             domain.IsOn,
			TsigJSON:         domain.Tsig,
			Status:           domain.Status,
			CreatedAt:        int64(domain.CreatedAt),
			NsCluster:        pbCluster,
			User:             pbUser,
			NsDomainGroupIds: groupIds,
			NsDomainGroups:   pbGroups,
		},
	}, nil
}

// CountAllNSDomains 计算域名数量
func (this *NSDomainService) CountAllNSDomains(ctx context.Context, req *pb.CountAllNSDomainsRequest) (*pb.RPCCountResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId

		// 检查分组
		if req.NsDomainGroupId > 0 {
			err = nameservers.SharedNSDomainGroupDAO.CheckUserGroup(tx, userId, req.NsDomainGroupId)
			if err != nil {
				return nil, err
			}
		}
	}

	count, err := nameservers.SharedNSDomainDAO.CountAllEnabledDomains(tx, req.NsClusterId, req.UserId, req.NsDomainGroupId, req.Status, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListNSDomains 列出单页域名
func (this *NSDomainService) ListNSDomains(ctx context.Context, req *pb.ListNSDomainsRequest) (*pb.ListNSDomainsResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	var isUserRequest = userId > 0

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId

		// 检查分组
		if req.NsDomainGroupId > 0 {
			err = nameservers.SharedNSDomainGroupDAO.CheckUserGroup(tx, userId, req.NsDomainGroupId)
			if err != nil {
				return nil, err
			}
		}
	}

	domains, err := nameservers.SharedNSDomainDAO.ListEnabledDomains(tx, req.NsClusterId, req.UserId, req.NsDomainGroupId, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbDomains = []*pb.NSDomain{}
	var cacheMap = utils.NewCacheMap()
	for _, domain := range domains {
		// 集群
		cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(domain.ClusterId))
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			continue
		}

		// 用户
		var pbUser *pb.User
		if domain.UserId > 0 && !isUserRequest {
			user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(domain.UserId), cacheMap)
			if err != nil {
				return nil, err
			}
			if user == nil {
				continue
			}
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}

		// groups
		var pbGroups = []*pb.NSDomainGroup{}
		var groupIds = domain.DecodeGroupIds()
		for _, groupId := range groupIds {
			group, err := nameservers.SharedNSDomainGroupDAO.FindEnabledNSDomainGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if group != nil && group.IsOn {
				pbGroups = append(pbGroups, &pb.NSDomainGroup{
					Id:     int64(group.Id),
					Name:   group.Name,
					IsOn:   group.IsOn,
					UserId: int64(group.UserId),
				})
			}
		}

		var pbCluster *pb.NSCluster
		if !isUserRequest {
			pbCluster = &pb.NSCluster{
				Id:   int64(cluster.Id),
				IsOn: cluster.IsOn,
				Name: cluster.Name,
			}
		}

		pbDomains = append(pbDomains, &pb.NSDomain{
			Id:               int64(domain.Id),
			Name:             domain.Name,
			Status:           domain.Status,
			IsOn:             domain.IsOn,
			CreatedAt:        int64(domain.CreatedAt),
			TsigJSON:         domain.Tsig,
			NsCluster:        pbCluster,
			User:             pbUser,
			NsDomainGroupIds: groupIds,
			NsDomainGroups:   pbGroups,
		})
	}

	return &pb.ListNSDomainsResponse{NsDomains: pbDomains}, nil
}

// ListNSDomainsAfterVersion 根据版本列出一组域名
func (this *NSDomainService) ListNSDomainsAfterVersion(ctx context.Context, req *pb.ListNSDomainsAfterVersionRequest) (*pb.ListNSDomainsAfterVersionResponse, error) {
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return &pb.ListNSDomainsAfterVersionResponse{
			NsDomains: nil,
		}, nil
	}

	// 集群ID
	var tx = this.NullTx()
	if req.Size <= 0 {
		req.Size = 2000
	}
	domains, err := nameservers.SharedNSDomainDAO.ListDomainsAfterVersion(tx, req.Version, req.Size)
	if err != nil {
		return nil, err
	}

	var pbDomains []*pb.NSDomain
	for _, domain := range domains {
		pbDomains = append(pbDomains, &pb.NSDomain{
			Id:        int64(domain.Id),
			Name:      domain.Name,
			IsOn:      domain.IsOn,
			Status:    domain.Status,
			IsDeleted: domain.State == nameservers.NSDomainStateDisabled || domain.Status != dnsconfigs.NSDomainStatusVerified,
			Version:   int64(domain.Version),
			TsigJSON:  domain.Tsig,
			NsCluster: &pb.NSCluster{Id: int64(domain.ClusterId)},
			User:      nil,
		})
	}
	return &pb.ListNSDomainsAfterVersionResponse{NsDomains: pbDomains}, nil
}

// FindNSDomainTSIG 查找TSIG配置
func (this *NSDomainService) FindNSDomainTSIG(ctx context.Context, req *pb.FindNSDomainTSIGRequest) (*pb.FindNSDomainTSIGResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	tsig, err := nameservers.SharedNSDomainDAO.FindEnabledDomainTSIG(tx, req.NsDomainId)
	if err != nil {
		return nil, err
	}
	return &pb.FindNSDomainTSIGResponse{TsigJSON: tsig}, nil
}

// UpdateNSDomainTSIG 修改TSIG配置
func (this *NSDomainService) UpdateNSDomainTSIG(ctx context.Context, req *pb.UpdateNSDomainTSIGRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSDomainDAO.UpdateDomainTSIG(tx, req.NsDomainId, req.TsigJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ExistNSDomains 检查一组域名是否在用户账户中存在
func (this *NSDomainService) ExistNSDomains(ctx context.Context, req *pb.ExistNSDomainsRequest) (*pb.ExistNSDomainsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var existingDomains = []string{}

	var tx = this.NullTx()
	for _, domainName := range req.Names {
		domainName = strings.ToLower(domainName)
		b, err := nameservers.SharedNSDomainDAO.ExistUserDomain(tx, req.UserId, domainName)
		if err != nil {
			return nil, err
		}
		if b {
			existingDomains = append(existingDomains, domainName)
		}
	}

	return &pb.ExistNSDomainsResponse{ExistingNames: existingDomains}, nil
}

// ExistVerifiedNSDomains 检查一组域名是否已通过验证
func (this *NSDomainService) ExistVerifiedNSDomains(ctx context.Context, req *pb.ExistVerifiedNSDomainsRequest) (*pb.ExistVerifiedNSDomainsResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var existingDomains = []string{}

	var tx = this.NullTx()
	for _, domainName := range req.Names {
		domainName = strings.ToLower(domainName)
		b, err := nameservers.SharedNSDomainDAO.ExistVerifiedDomain(tx, domainName)
		if err != nil {
			return nil, err
		}
		if b {
			existingDomains = append(existingDomains, domainName)
		}
	}

	return &pb.ExistVerifiedNSDomainsResponse{ExistingNames: existingDomains}, nil
}

// FindNSDomainVerifyingInfo 获取域名验证信息
func (this *NSDomainService) FindNSDomainVerifyingInfo(ctx context.Context, req *pb.FindNSDomainVerifyingInfoRequest) (*pb.FindNSDomainVerifyingInfoResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	// 检查权限
	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	domain, err := nameservers.SharedNSDomainDAO.FindDomainVerifyingInfo(tx, req.NsDomainId, true)
	if err != nil {
		return nil, err
	}

	if domain == nil {
		return &pb.FindNSDomainVerifyingInfoResponse{
			Txt:       "",
			ExpiresAt: 0,
			Status:    "",
		}, nil
	}

	return &pb.FindNSDomainVerifyingInfoResponse{
		Txt:       domain.VerifyTXT,
		ExpiresAt: int64(domain.VerifyExpiresAt),
		Status:    domain.Status,
	}, nil
}

// VerifyNSDomain 验证域名信息
func (this *NSDomainService) VerifyNSDomain(ctx context.Context, req *pb.VerifyNSDomainRequest) (*pb.VerifyNSDomainResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	// 检查权限
	var tx = this.NullTx()
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	// 客户端需要处理的错误代号列表
	const (
		ErrorCodeDomainNotFound      = "DomainNotFound"
		ErrorCodeInvalidStatus       = "InvalidStatus"
		ErrorCodeInvalidDNSHosts     = "InvalidDNSHosts"
		ErrorCodeInvalidTXT          = "InvalidTXT"
		ErrorCodeTXTNotFound         = "TXTNotFound"
		ErrorCodeTXTExpired          = "TXTExpired"
		ErrorCodeVerifiedByOtherUser = "VerifiedByOtherUser"
	)

	// 验证状态
	domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, req.NsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.VerifyNSDomainResponse{
			ErrorCode: ErrorCodeDomainNotFound,
		}, nil
	}

	if domain.Status != dnsconfigs.NSDomainStatusNone {
		return &pb.VerifyNSDomainResponse{
			ErrorCode:    ErrorCodeInvalidStatus,
			ErrorMessage: "invalid status '" + domain.Status + "'",
		}, nil
	}

	if len(domain.VerifyTXT) == 0 || int64(domain.VerifyExpiresAt) < time.Now().Unix() {
		return &pb.VerifyNSDomainResponse{
			ErrorCode: ErrorCodeTXTExpired,
		}, nil
	}

	// 检查是否已被别的用户所验证
	if userId > 0 {
		verifiedDomain, err := nameservers.SharedNSDomainDAO.FindVerifiedDomainWithName(tx, domain.Name)
		if err != nil {
			return nil, err
		}
		if verifiedDomain != nil {
			return &pb.VerifyNSDomainResponse{
				ErrorCode: ErrorCodeVerifiedByOtherUser,
			}, nil
		}
	}

	// 当前集群的DNS Hosts
	userConfig, err := models.SharedSysSettingDAO.ReadNSUserConfig(tx)
	if err != nil {
		return nil, err
	}
	if userConfig == nil {
		return nil, errors.New("invalid NSUserConfig")
	}

	if userConfig.DefaultClusterId <= 0 {
		return nil, errors.New("invalid NSUserConfig.DefaultClusterId")
	}

	userHosts, err := models.SharedNSClusterDAO.FindClusterHosts(tx, userConfig.DefaultClusterId)
	if err != nil {
		return nil, err
	}
	if len(userHosts) == 0 {
		return nil, errors.New("no hosts in the ns cluster")
	}

	// 验证主机地址
	hosts, err := utils.LookupNS(domain.Name)
	if err != nil {
		remotelogs.Error("NSDomainService", "lookup NS '"+domain.Name+"' failed: "+err.Error())
		return &pb.VerifyNSDomainResponse{
			ErrorCode: ErrorCodeInvalidDNSHosts,
		}, nil
	}

	if len(hosts) == 0 {
		return &pb.VerifyNSDomainResponse{
			ErrorCode: ErrorCodeInvalidDNSHosts,
		}, nil
	}

	if len(hosts) > 0 {
		for _, host := range hosts {
			host = strings.TrimSuffix(host, ".")
			if !lists.ContainsString(userHosts, host) {
				return &pb.VerifyNSDomainResponse{
					ErrorCode:    ErrorCodeInvalidDNSHosts,
					ErrorMessage: "invalid dns host '" + host + "'",
				}, nil
			}
		}
	}

	// 验证TXT
	txtList, err := utils.LookupTXT("yanzheng." + domain.Name)
	if err != nil {
		return nil, err
	}

	if len(txtList) == 0 {
		return &pb.VerifyNSDomainResponse{
			ErrorCode:    ErrorCodeTXTNotFound,
			ErrorMessage: "",
		}, nil
	}

	if !lists.ContainsString(txtList, domain.VerifyTXT) {
		return &pb.VerifyNSDomainResponse{
			ErrorCode:    ErrorCodeInvalidTXT,
			ErrorMessage: strings.Join(txtList, ", "),
		}, nil
	}

	// 验证通过
	err = nameservers.SharedNSDomainDAO.UpdateDomainStatus(tx, req.NsDomainId, dnsconfigs.NSDomainStatusVerified)
	if err != nil {
		return nil, err
	}

	return &pb.VerifyNSDomainResponse{
		IsOk: true,
	}, nil
}
