// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build plus

package nameservers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"regexp"
	"strings"
)

// NSRecordService 域名记录相关服务
type NSRecordService struct {
	services.BaseService
}

// CreateNSRecord 创建记录
func (this *NSRecordService) CreateNSRecord(ctx context.Context, req *pb.CreateNSRecordRequest) (*pb.CreateNSRecordResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	recordId, err := nameservers.SharedNSRecordDAO.CreateRecord(tx, req.NsDomainId, req.Description, req.Name, req.Type, req.Value, req.Ttl, req.NsRouteCodes)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSRecordResponse{NsRecordId: recordId}, nil
}

// CreateNSRecords 创建记录
func (this *NSRecordService) CreateNSRecords(ctx context.Context, req *pb.CreateNSRecordsRequest) (*pb.CreateNSRecordsResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

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

	var recordIds = []int64{}
	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, name := range req.Names {
			recordId, err := nameservers.SharedNSRecordDAO.CreateRecord(tx, req.NsDomainId, req.Description, name, req.Type, req.Value, req.Ttl, req.NsRouteCodes)
			if err != nil {
				return err
			}
			recordIds = append(recordIds, recordId)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateNSRecordsResponse{NsRecordIds: recordIds}, nil
}

// CreateNSRecordsWithDomainNames 为一组域名批量创建记录
func (this *NSRecordService) CreateNSRecordsWithDomainNames(ctx context.Context, req *pb.CreateNSRecordsWithDomainNamesRequest) (*pb.RPCSuccess, error) {
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

	if len(req.RecordsJSON) == 0 {
		return this.Success()
	}

	type record struct {
		Name       string   `json:"name"`
		Type       string   `json:"type"`
		Value      string   `json:"value"`
		RouteCodes []string `json:"routeCodes"`
		TTL        int32    `json:"ttl"`
	}

	var records = []*record{}
	err = json.Unmarshal(req.RecordsJSON, &records)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return this.Success()
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, domainName := range req.NsDomainNames {
			domainName = strings.ToLower(strings.TrimSpace(domainName))
			if len(domainName) == 0 {
				continue
			}
			domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, 0, req.UserId, domainName)
			if err != nil {
				return err
			}
			if domainId <= 0 {
				continue
			}

			// 是否删除所有以往记录
			if req.RemoveAll {
				err = nameservers.SharedNSRecordDAO.DisableRecordsInDomain(tx, domainId)
				if err != nil {
					return err
				}
			}

			for _, record := range records {
				record.Type = strings.ToLower(record.Type)

				if !req.RemoveAll && req.RemoveOld {
					err = nameservers.SharedNSRecordDAO.DisableRecordsInDomainWithNameAndType(tx, domainId, record.Name, record.Type)
					if err != nil {
						return err
					}
				}

				_, err = nameservers.SharedNSRecordDAO.CreateRecord(tx, domainId, "批量创建", record.Name, strings.ToUpper(record.Type), record.Value, record.TTL, record.RouteCodes)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNSRecordsWithDomainNames 批量修改一组域名的一组记录
func (this *NSRecordService) UpdateNSRecordsWithDomainNames(ctx context.Context, req *pb.UpdateNSRecordsWithDomainNamesRequest) (*pb.RPCSuccess, error) {
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

	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, domainName := range req.NsDomainNames {
			domainName = strings.ToLower(strings.TrimSpace(domainName))
			if len(domainName) == 0 {
				continue
			}
			domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, 0, req.UserId, domainName)
			if err != nil {
				return err
			}
			if domainId <= 0 {
				continue
			}
			err = nameservers.SharedNSRecordDAO.UpdateRecordsWithDomainId(tx, domainId, req.SearchName, req.SearchType, req.SearchValue, req.SearchNSRouteCodes, req.NewName, req.NewType, req.NewValue, req.NewNSRouteCodes)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSRecordsWithDomainNames 批量删除一组域名的一组记录
func (this *NSRecordService) DeleteNSRecordsWithDomainNames(ctx context.Context, req *pb.DeleteNSRecordsWithDomainNamesRequest) (*pb.RPCSuccess, error) {
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

	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, domainName := range req.NsDomainNames {
			domainName = strings.ToLower(strings.TrimSpace(domainName))
			if len(domainName) == 0 {
				continue
			}
			domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, 0, req.UserId, domainName)
			if err != nil {
				return err
			}
			if domainId <= 0 {
				continue
			}

			err = nameservers.SharedNSRecordDAO.DisableRecordsWithDomainId(tx, domainId, req.SearchName, req.SearchType, req.SearchValue, req.SearchNSRouteCodes)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNSRecordsIsOnWithDomainNames 批量一组域名的一组记录启用状态
func (this *NSRecordService) UpdateNSRecordsIsOnWithDomainNames(ctx context.Context, req *pb.UpdateNSRecordsIsOnWithDomainNamesRequest) (*pb.RPCSuccess, error) {
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

	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, domainName := range req.NsDomainNames {
			domainName = strings.ToLower(strings.TrimSpace(domainName))
			if len(domainName) == 0 {
				continue
			}
			domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, 0, req.UserId, domainName)
			if err != nil {
				return err
			}
			if domainId <= 0 {
				continue
			}

			err = nameservers.SharedNSRecordDAO.UpdateRecordsIsOnWithDomainId(tx, domainId, req.SearchName, req.SearchType, req.SearchValue, req.SearchNSRouteCodes, req.IsOn)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ImportNSRecords 导入域名解析
func (this *NSRecordService) ImportNSRecords(ctx context.Context, req *pb.ImportNSRecordsRequest) (*pb.RPCSuccess, error) {
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

	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, record := range req.NsRecords {
			var domainName = strings.ToLower(strings.TrimSpace(record.NsDomainName))
			if len(domainName) == 0 {
				continue
			}
			domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, 0, req.UserId, domainName)
			if err != nil {
				return err
			}
			if domainId <= 0 {
				continue
			}

			if record.Ttl <= 0 {
				record.Ttl = 600
			}

			_, err = nameservers.SharedNSRecordDAO.CreateRecord(tx, domainId, "批量导入", record.Name, record.Type, record.Value, record.Ttl, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNSRecord 修改记录
func (this *NSRecordService) UpdateNSRecord(ctx context.Context, req *pb.UpdateNSRecordRequest) (*pb.RPCSuccess, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSRecordDAO.CheckUserRecord(tx, userId, req.NsRecordId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSRecordDAO.UpdateRecord(tx, req.NsRecordId, req.Description, req.Name, req.Type, req.Value, req.Ttl, req.NsRouteCodes, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSRecord 删除记录
func (this *NSRecordService) DeleteNSRecord(ctx context.Context, req *pb.DeleteNSRecordRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSRecordDAO.CheckUserRecord(tx, userId, req.NsRecordId)
		if err != nil {
			return nil, err
		}
	}

	err = nameservers.SharedNSRecordDAO.DisableNSRecord(tx, req.NsRecordId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllNSRecords 计算记录数量
func (this *NSRecordService) CountAllNSRecords(ctx context.Context, req *pb.CountAllNSRecordsRequest) (*pb.RPCCountResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	count, err := nameservers.SharedNSRecordDAO.CountAllEnabledDomainRecords(tx, req.NsDomainId, req.Type, req.Keyword, req.NsRouteCode)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountAllNSRecordsWithName 查询相同记录名的记录数
func (this *NSRecordService) CountAllNSRecordsWithName(ctx context.Context, req *pb.CountAllNSRecordsWithNameRequest) (*pb.RPCCountResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	count, err := nameservers.SharedNSRecordDAO.CountAllRecordsWithName(tx, req.NsDomainId, req.Type, req.Name)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListNSRecords 读取单页记录
func (this *NSRecordService) ListNSRecords(ctx context.Context, req *pb.ListNSRecordsRequest) (*pb.ListNSRecordsResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSDomainDAO.CheckUserDomain(tx, userId, req.NsDomainId)
		if err != nil {
			return nil, err
		}
	}

	records, err := nameservers.SharedNSRecordDAO.ListEnabledRecords(tx, req.NsDomainId, req.Type, req.Keyword, req.NsRouteCode, req.NameAsc, req.NameDesc, req.TypeAsc, req.TypeDesc, req.TtlAsc, req.TtlDesc, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbRecords = []*pb.NSRecord{}
	for _, record := range records {
		// 线路
		var pbRoutes = []*pb.NSRoute{}
		for _, routeCode := range record.DecodeRouteIds() {
			route, err := nameservers.SharedNSRouteDAO.FindEnabledRouteWithCode(tx, routeCode)
			if err != nil {
				return nil, err
			}
			if route == nil {
				continue
			}
			pbRoutes = append(pbRoutes, &pb.NSRoute{
				Id:   int64(route.Id),
				Name: route.Name,
				Code: route.Code,
			})

			// TODO 读取其他线路
		}

		pbRecords = append(pbRecords, &pb.NSRecord{
			Id:          int64(record.Id),
			Description: record.Description,
			Name:        record.Name,
			Type:        record.Type,
			Value:       record.Value,
			Ttl:         types.Int32(record.Ttl),
			Weight:      types.Int32(record.Weight),
			CreatedAt:   int64(record.CreatedAt),
			IsOn:        record.IsOn,
			NsDomain:    nil,
			NsRoutes:    pbRoutes,
		})
	}
	return &pb.ListNSRecordsResponse{NsRecords: pbRecords}, nil
}

// FindNSRecord 查询单个记录信息
func (this *NSRecordService) FindNSRecord(ctx context.Context, req *pb.FindNSRecordRequest) (*pb.FindNSRecordResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = nameservers.SharedNSRecordDAO.CheckUserRecord(tx, userId, req.NsRecordId)
		if err != nil {
			return nil, err
		}
	}

	record, err := nameservers.SharedNSRecordDAO.FindEnabledNSRecord(tx, req.NsRecordId)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &pb.FindNSRecordResponse{NsRecord: nil}, nil
	}

	// 域名
	domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, int64(record.DomainId))
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindNSRecordResponse{NsRecord: nil}, nil
	}
	var pbDomain = &pb.NSDomain{
		Id:   int64(domain.Id),
		Name: domain.Name,
		IsOn: domain.IsOn,
	}

	// 线路
	var pbRoutes = []*pb.NSRoute{}
	for _, routeCode := range record.DecodeRouteIds() {
		route, err := nameservers.SharedNSRouteDAO.FindEnabledRouteWithCode(tx, routeCode)
		if err != nil {
			return nil, err
		}
		if route == nil {
			continue
		}
		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Id:   int64(route.Id),
			Name: route.Name,
			Code: route.Code,
		})
	}

	// TODO 读取其他线路

	return &pb.FindNSRecordResponse{NsRecord: &pb.NSRecord{
		Id:          int64(record.Id),
		Description: record.Description,
		Name:        record.Name,
		Type:        record.Type,
		Value:       record.Value,
		Ttl:         types.Int32(record.Ttl),
		Weight:      types.Int32(record.Weight),
		CreatedAt:   int64(record.CreatedAt),
		IsOn:        record.IsOn,
		NsDomain:    pbDomain,
		NsRoutes:    pbRoutes,
	}}, nil
}

// FindNSRecordWithNameAndType 使用名称和类型查询单个记录信息
func (this *NSRecordService) FindNSRecordWithNameAndType(ctx context.Context, req *pb.FindNSRecordWithNameAndTypeRequest) (*pb.FindNSRecordWithNameAndTypeResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	if req.NsDomainId <= 0 {
		return &pb.FindNSRecordWithNameAndTypeResponse{
			NsRecord: nil,
		}, nil
	}

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

	record, err := nameservers.SharedNSRecordDAO.FindEnabledRecordWithName(tx, req.NsDomainId, req.Name, req.Type)
	if err != nil {
		return nil, err
	}

	if record == nil {
		return &pb.FindNSRecordWithNameAndTypeResponse{
			NsRecord: nil,
		}, nil
	}

	// 线路
	var pbRoutes = []*pb.NSRoute{}
	for _, routeCode := range record.DecodeRouteIds() {
		route, err := nameservers.SharedNSRouteDAO.FindEnabledRouteWithCode(tx, routeCode)
		if err != nil {
			return nil, err
		}
		if route == nil {
			continue
		}
		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Id:   int64(route.Id),
			Name: route.Name,
			Code: route.Code,
		})
	}

	return &pb.FindNSRecordWithNameAndTypeResponse{
		NsRecord: &pb.NSRecord{
			Id:          int64(record.Id),
			Description: record.Description,
			Name:        record.Name,
			Type:        record.Type,
			Value:       record.Value,
			Ttl:         types.Int32(record.Ttl),
			Weight:      types.Int32(record.Weight),
			CreatedAt:   int64(record.CreatedAt),
			IsOn:        record.IsOn,
			NsRoutes:    pbRoutes,
		},
	}, nil
}

// ListNSRecordsAfterVersion 根据版本列出一组记录
func (this *NSRecordService) ListNSRecordsAfterVersion(ctx context.Context, req *pb.ListNSRecordsAfterVersionRequest) (*pb.ListNSRecordsAfterVersionResponse, error) {
	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return nil, errors.New("non commercial user")
	}

	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	// 检查是否为商业用户
	if !teaconst.IsPlus {
		return &pb.ListNSRecordsAfterVersionResponse{
			NsRecords: nil,
		}, nil
	}

	// 集群ID
	var tx = this.NullTx()
	if req.Size <= 0 {
		req.Size = 2000
	}
	records, err := nameservers.SharedNSRecordDAO.ListRecordsAfterVersion(tx, req.Version, req.Size)
	if err != nil {
		return nil, err
	}

	var pbRecords []*pb.NSRecord
	for _, record := range records {
		// 线路
		pbRoutes := []*pb.NSRoute{}
		routeIds := record.DecodeRouteIds()
		for _, routeId := range routeIds {
			var routeIdInt int64 = 0
			if regexp.MustCompile(`^id:\d+$`).MatchString(routeId) {
				routeIdInt = types.Int64(routeId[strings.Index(routeId, ":")+1:])
			}

			pbRoutes = append(pbRoutes, &pb.NSRoute{
				Id:   routeIdInt,
				Code: routeId,
			})
		}

		// TODO 读取其他线路

		pbRecords = append(pbRecords, &pb.NSRecord{
			Id:          int64(record.Id),
			Description: "",
			Name:        record.Name,
			Type:        record.Type,
			Value:       record.Value,
			Ttl:         types.Int32(record.Ttl),
			Weight:      types.Int32(record.Weight),
			IsDeleted:   record.State == nameservers.NSRecordStateDisabled,
			IsOn:        record.IsOn,
			Version:     int64(record.Version),
			NsDomain:    &pb.NSDomain{Id: int64(record.DomainId)},
			NsRoutes:    pbRoutes,
		})
	}
	return &pb.ListNSRecordsAfterVersionResponse{NsRecords: pbRecords}, nil
}
