package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/sslconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/types"
	"net/url"
	"regexp"
	"strings"
)

type ReverseProxyService struct {
	BaseService
}

// CreateReverseProxy 创建反向代理
func (this *ReverseProxyService) CreateReverseProxy(ctx context.Context, req *pb.CreateReverseProxyRequest) (*pb.CreateReverseProxyResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验源站
	}

	var tx = this.NullTx()

	reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, userId, req.SchedulingJSON, req.PrimaryOriginsJSON, req.BackupOriginsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateReverseProxyResponse{ReverseProxyId: reverseProxyId}, nil
}

// FindEnabledReverseProxy 查找反向代理
func (this *ReverseProxyService) FindEnabledReverseProxy(ctx context.Context, req *pb.FindEnabledReverseProxyRequest) (*pb.FindEnabledReverseProxyResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	reverseProxy, err := models.SharedReverseProxyDAO.FindEnabledReverseProxy(tx, req.ReverseProxyId)
	if err != nil {
		return nil, err
	}
	if reverseProxy == nil {
		return &pb.FindEnabledReverseProxyResponse{ReverseProxy: nil}, nil
	}

	result := &pb.ReverseProxy{
		Id:                 int64(reverseProxy.Id),
		SchedulingJSON:     reverseProxy.Scheduling,
		PrimaryOriginsJSON: reverseProxy.PrimaryOrigins,
		BackupOriginsJSON:  reverseProxy.BackupOrigins,
	}
	return &pb.FindEnabledReverseProxyResponse{ReverseProxy: result}, nil
}

// FindEnabledReverseProxyConfig 查找反向代理配置
func (this *ReverseProxyService) FindEnabledReverseProxyConfig(ctx context.Context, req *pb.FindEnabledReverseProxyConfigRequest) (*pb.FindEnabledReverseProxyConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	config, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, req.ReverseProxyId, nil)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledReverseProxyConfigResponse{ReverseProxyJSON: configData}, nil
}

// UpdateReverseProxyScheduling 修改反向代理调度算法
func (this *ReverseProxyService) UpdateReverseProxyScheduling(ctx context.Context, req *pb.UpdateReverseProxySchedulingRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	err = models.SharedReverseProxyDAO.UpdateReverseProxyScheduling(tx, req.ReverseProxyId, req.SchedulingJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateReverseProxyPrimaryOrigins 修改主要源站信息
func (this *ReverseProxyService) UpdateReverseProxyPrimaryOrigins(ctx context.Context, req *pb.UpdateReverseProxyPrimaryOriginsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	err = models.SharedReverseProxyDAO.UpdateReverseProxyPrimaryOrigins(tx, req.ReverseProxyId, req.OriginsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateReverseProxyBackupOrigins 修改备用源站信息
func (this *ReverseProxyService) UpdateReverseProxyBackupOrigins(ctx context.Context, req *pb.UpdateReverseProxyBackupOriginsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	err = models.SharedReverseProxyDAO.UpdateReverseProxyBackupOrigins(tx, req.ReverseProxyId, req.OriginsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateReverseProxy 修改是否启用
func (this *ReverseProxyService) UpdateReverseProxy(ctx context.Context, req *pb.UpdateReverseProxyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	// 校验参数
	var connTimeout = &shared.TimeDuration{}
	if len(req.ConnTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ConnTimeoutJSON, connTimeout)
		if err != nil {
			return nil, err
		}
	}

	var readTimeout = &shared.TimeDuration{}
	if len(req.ReadTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ReadTimeoutJSON, readTimeout)
		if err != nil {
			return nil, err
		}
	}

	var idleTimeout = &shared.TimeDuration{}
	if len(req.IdleTimeoutJSON) > 0 {
		err = json.Unmarshal(req.IdleTimeoutJSON, idleTimeout)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedReverseProxyDAO.UpdateReverseProxy(tx, req.ReverseProxyId, types.Int8(req.RequestHostType), req.RequestHost, req.RequestHostExcludingPort, req.RequestURI, req.StripPrefix, req.AutoFlush, req.AddHeaders, connTimeout, readTimeout, idleTimeout, req.MaxConns, req.MaxIdleConns, req.ProxyProtocolJSON, req.FollowRedirects)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ------- api 客户定制化接口

// 创建代理源站信息
type CreateProxyOriginsRequest struct {
	ServerId     int64    `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"`       //服务id
	Protocol     string   `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`        //源站协议
	Addr         string   `protobuf:"bytes,3,opt,name=addr,proto3" json:"addr,omitempty"`                //地址
	Domains      []string `protobuf:"bytes,4,rep,name=domains,proto3" json:"domains,omitempty"`          //专属域名数组
	Weight       int32    `protobuf:"varint,5,opt,name=weight,proto3" json:"weight,omitempty"`           //权重
	Name         string   `protobuf:"bytes,6,opt,name=name,proto3" json:"name,omitempty"`                //名称
	ConnTimeout  int32    `protobuf:"varint,7,opt,name=connTimeout,proto3" json:"connTimeout,omitempty"` //连接超时
	ReadTimeout  int32    `protobuf:"varint,8,opt,name=readTimeout,proto3" json:"readTimeout,omitempty"`
	MaxConns     int32    `protobuf:"varint,9,opt,name=maxConns,proto3" json:"maxConns,omitempty"`          //最大连接数
	MaxIdleConns int32    `protobuf:"varint,10,opt,name=maxIdleConns,proto3" json:"maxIdleConns,omitempty"` //最大空闲连接数
	IdleTimeout  int32    `protobuf:"varint,11,opt,name=idleTimeout,proto3" json:"idleTimeout,omitempty"`   //最大空闲超时
	IsOn         bool     `protobuf:"varint,12,opt,name=isOn,proto3" json:"isOn,omitempty"`                 //是否开启
	Desc         string   `protobuf:"bytes,13,opt,name=desc,proto3" json:"desc,omitempty"`                  //备注
	IsPrimary    bool     `protobuf:"varint,14,opt,name=isPrimary,proto3" json:"isPrimary,omitempty"`       //是否是主要源站
	CertId       int64    `protobuf:"varint,15,opt,name=certId,proto3" json:"certId,omitempty"`             //源站HTTPS时，访问证书id
	FollowPort   bool     `protobuf:"varint,16,opt,name=followPort,proto3" json:"followPort,omitempty"`     //跟随端口
}

type CreateProxyOriginsResponse struct {
	OriginId int64 `protobuf:"varint,1,opt,name=originId,proto3" json:"originId,omitempty"` //源站id
}

// 修改代理源站信息
type UpdateProxyOriginsRequest struct {
	OriginId     int64    `protobuf:"varint,1,opt,name=originId,proto3" json:"originId,omitempty"`       //源站id
	Protocol     string   `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`        //源站协议
	Addr         string   `protobuf:"bytes,3,opt,name=addr,proto3" json:"addr,omitempty"`                //地址
	Domains      []string `protobuf:"bytes,4,rep,name=domains,proto3" json:"domains,omitempty"`          //专属域名数组
	Weight       int32    `protobuf:"varint,5,opt,name=weight,proto3" json:"weight,omitempty"`           //权重
	Name         string   `protobuf:"bytes,6,opt,name=name,proto3" json:"name,omitempty"`                //名称
	ConnTimeout  int32    `protobuf:"varint,7,opt,name=connTimeout,proto3" json:"connTimeout,omitempty"` //连接超时
	ReadTimeout  int32    `protobuf:"varint,8,opt,name=readTimeout,proto3" json:"readTimeout,omitempty"`
	MaxConns     int32    `protobuf:"varint,9,opt,name=maxConns,proto3" json:"maxConns,omitempty"`          //最大连接数
	MaxIdleConns int32    `protobuf:"varint,10,opt,name=maxIdleConns,proto3" json:"maxIdleConns,omitempty"` //最大空闲连接数
	IdleTimeout  int32    `protobuf:"varint,11,opt,name=idleTimeout,proto3" json:"idleTimeout,omitempty"`   //最大空闲超时
	IsOn         bool     `protobuf:"varint,12,opt,name=isOn,proto3" json:"isOn,omitempty"`                 //是否开启
	Desc         string   `protobuf:"bytes,13,opt,name=desc,proto3" json:"desc,omitempty"`                  //备注
	IsPrimary    bool     `protobuf:"varint,14,opt,name=isPrimary,proto3" json:"isPrimary,omitempty"`       //是否是主要源站
	CertId       int64    `protobuf:"varint,15,opt,name=certId,proto3" json:"certId,omitempty"`             //源站HTTPS时，访问证书id
	FollowPort   bool     `protobuf:"varint,16,opt,name=followPort,proto3" json:"followPort,omitempty"`     //跟随端口
}

type DeleteProxyOriginsRequest struct {
	ServerId  int64 `protobuf:"varint,1,opt,name=serverId,proto3" json:"serverId,omitempty"`   //服务id
	OriginId  int64 `protobuf:"varint,2,opt,name=originId,proto3" json:"originId,omitempty"`   //源站id
	IsPrimary bool  `protobuf:"varint,3,opt,name=isPrimary,proto3" json:"isPrimary,omitempty"` //是否是主要源站
}

// CreateProxyOrigins 创建代理源站地址
func (this *ReverseProxyService) CreateProxyOrigins(ctx context.Context, req *CreateProxyOriginsRequest) (*CreateProxyOriginsResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	failed := true
	tx := this.NullTx()
	if req.Addr == "" {
		return nil, fmt.Errorf("请输入源站地址")
	}
	if req.ServerId == 0 {
		return nil, fmt.Errorf("请输入服务id")
	}
	addr := req.Addr
	// 是否是完整的地址
	if (req.Protocol == "http" || req.Protocol == "https") && regexp.MustCompile(`^(http|https)://`).MatchString(addr) {
		u, err := url.Parse(addr)
		if err == nil {
			addr = u.Host
		}
	}

	addr = strings.ReplaceAll(addr, "：", ":")
	addr = regexp.MustCompile(`\s+`).ReplaceAllString(addr, "")
	portIndex := strings.LastIndex(addr, ":")
	if portIndex < 0 {
		if req.Protocol == "http" {
			addr += ":80"
		} else if req.Protocol == "https" {
			addr += ":443"
		} else {
			return nil, fmt.Errorf("地址中需要带有端口")
		}
		portIndex = strings.LastIndex(addr, ":")
	}
	host := addr[:portIndex]
	port := addr[portIndex+1:]
	if port == "0" {
		return nil, fmt.Errorf("端口号不能为0")
	}

	connTimeout := &shared.TimeDuration{
		Count: int64(req.ConnTimeout),
		Unit:  shared.TimeDurationUnitSecond,
	}

	readTimeout := &shared.TimeDuration{
		Count: int64(req.ReadTimeout),
		Unit:  shared.TimeDurationUnitSecond,
	}

	idleTimeout := &shared.TimeDuration{
		Count: int64(req.IdleTimeout),
		Unit:  shared.TimeDurationUnitSecond,
	}
	var domains = []string{}
	if len(req.Domains) > 0 {

		// 去除可能误加的斜杠
		for index, domain := range req.Domains {
			domains[index] = strings.TrimSuffix(domain, "/")
		}
	}
	addrJSON, err := json.Marshal(pb.NetworkAddress{
		Protocol:  req.Protocol,
		Host:      host,
		PortRange: port,
	})
	if err != nil {
		return nil, err
	}
	var sslCert *sslconfigs.SSLCertRef

	if req.CertId != 0 {
		sslCert = &sslconfigs.SSLCertRef{
			IsOn:   true,
			CertId: req.CertId,
		}
	}
	originId, err := models.SharedOriginDAO.CreateOrigin(tx, adminId, 0, req.Name, string(addrJSON), req.Desc, req.Weight, req.IsOn,
		connTimeout, readTimeout, idleTimeout, req.MaxConns, req.MaxIdleConns, sslCert, domains, "", req.FollowPort)
	if err != nil {
		return nil, err
	}
	defer func() {
		if failed { //创建失败 删除已创建的源站记录
			_ = models.SharedOriginDAO.DisableOrigin(tx, originId)
		}
	}()
	originRef := &serverconfigs.OriginRef{
		IsOn:     true,
		OriginId: originId,
	}
	serverReverseProxy, err := models.SharedServerDAO.FindReverseProxyRef(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	reverseProxy, err := models.SharedReverseProxyDAO.FindEnabledReverseProxy(tx, serverReverseProxy.ReverseProxyId)
	if err != nil {
		return nil, err
	}
	if reverseProxy == nil {
		return nil, fmt.Errorf("reverse proxy should not be nil")
	}

	origins := []*serverconfigs.OriginRef{}
	if req.IsPrimary {
		if len(reverseProxy.PrimaryOrigins) > 0 {
			err = json.Unmarshal(reverseProxy.PrimaryOrigins, &origins)
			if err != nil {
				return nil, err
			}
		}
	} else {

		if len(reverseProxy.BackupOrigins) > 0 {
			err = json.Unmarshal(reverseProxy.BackupOrigins, &origins)
			if err != nil {
				return nil, err
			}
		}
	}

	origins = append(origins, originRef)
	originsData, err := json.Marshal(origins)
	if err != nil {
		return nil, err
	}
	if req.IsPrimary {
		err = models.SharedReverseProxyDAO.UpdateReverseProxyPrimaryOrigins(tx, int64(reverseProxy.Id), originsData)
	} else {
		err = models.SharedReverseProxyDAO.UpdateReverseProxyBackupOrigins(tx, int64(reverseProxy.Id), originsData)
	}
	if err != nil {
		return nil, err
	}
	failed = false
	return &CreateProxyOriginsResponse{OriginId: originId}, nil
}

// UpdateProxyOrigins 修改代理源站信息
func (this *ReverseProxyService) UpdateProxyOrigins(ctx context.Context, req *UpdateProxyOriginsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	var orginHost string
	if req.Addr == "" {
		return nil, fmt.Errorf("请输入源站地址")
	}
	if req.OriginId == 0 {
		return nil, fmt.Errorf("请输入源站id")
	} else {
		orgin, err := models.SharedOriginDAO.FindEnabledOrigin(tx, req.OriginId)
		if err != nil {
			return nil, fmt.Errorf("查询该源站id[%v]失败：%s", req.OriginId, err.Error())
		}
		orginHost = orgin.Host
	}
	addr := req.Addr

	// 是否是完整的地址
	if (req.Protocol == "http" || req.Protocol == "https") && regexp.MustCompile(`^(http|https)://`).MatchString(addr) {
		u, err := url.Parse(addr)
		if err == nil {
			addr = u.Host
		}
	}

	addr = strings.ReplaceAll(addr, "：", ":")
	addr = regexp.MustCompile(`\s+`).ReplaceAllString(addr, "")
	portIndex := strings.LastIndex(addr, ":")
	if portIndex < 0 {
		if req.Protocol == "http" {
			addr += ":80"
		} else if req.Protocol == "https" {
			addr += ":443"
		} else {
			return nil, fmt.Errorf("地址中需要带有端口")
		}
		portIndex = strings.LastIndex(addr, ":")
	}
	host := addr[:portIndex]
	port := addr[portIndex+1:]
	if port == "0" {
		return nil, fmt.Errorf("端口号不能为0")
	}

	connTimeout := &shared.TimeDuration{
		Count: int64(req.ConnTimeout),
		Unit:  shared.TimeDurationUnitSecond,
	}

	readTimeout := &shared.TimeDuration{
		Count: int64(req.ReadTimeout),
		Unit:  shared.TimeDurationUnitSecond,
	}

	idleTimeout := &shared.TimeDuration{
		Count: int64(req.IdleTimeout),
		Unit:  shared.TimeDurationUnitSecond,
	}
	var domains = []string{}
	if len(req.Domains) > 0 {

		// 去除可能误加的斜杠
		for index, domain := range req.Domains {
			domains[index] = strings.TrimSuffix(domain, "/")
		}
	}
	addrJSON, err := json.Marshal(pb.NetworkAddress{
		Protocol:  req.Protocol,
		Host:      host,
		PortRange: port,
	})
	if err != nil {
		return nil, err
	}
	// 证书
	var sslCert *sslconfigs.SSLCertRef

	if req.CertId != 0 {
		sslCert = &sslconfigs.SSLCertRef{
			IsOn:   true,
			CertId: req.CertId,
		}
	}
	err = models.SharedOriginDAO.UpdateOrigin(tx, req.OriginId, req.Name, string(addrJSON), req.Desc, req.Weight, req.IsOn,
		connTimeout, readTimeout, idleTimeout, req.MaxConns, req.MaxIdleConns, sslCert, domains, orginHost, req.FollowPort)
	if err != nil {
		return nil, err
	}

	return &pb.RPCSuccess{}, nil
}

// DeleteProxyOrigins 删除代理源站信息
func (this *ReverseProxyService) DeleteProxyOrigins(ctx context.Context, req *DeleteProxyOriginsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	if req.ServerId == 0 || req.OriginId == 0 {
		return nil, fmt.Errorf("参数错误，服务id或源站id不能为空")
	}
	serverReverseProxy, err := models.SharedServerDAO.FindReverseProxyRef(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	reverseProxy, err := models.SharedReverseProxyDAO.FindEnabledReverseProxy(tx, serverReverseProxy.ReverseProxyId)
	if err != nil {
		return nil, err
	}
	if reverseProxy == nil {
		return nil, fmt.Errorf("reverse proxy should not be nil")
	}

	origins := []*serverconfigs.OriginRef{}
	if req.IsPrimary {
		err = json.Unmarshal(reverseProxy.PrimaryOrigins, &origins)
		if err != nil {
			return nil, err
		}
	} else {
		err = json.Unmarshal(reverseProxy.BackupOrigins, &origins)
		if err != nil {
			return nil, err
		}
	}
	result := []*serverconfigs.OriginRef{}
	for _, origin := range origins {
		if origin.OriginId == req.OriginId {
			continue
		}
		result = append(result, origin)
	}
	resultData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	if req.IsPrimary {
		err = models.SharedReverseProxyDAO.UpdateReverseProxyPrimaryOrigins(tx, int64(reverseProxy.Id), resultData)
	} else {
		err = models.SharedReverseProxyDAO.UpdateReverseProxyBackupOrigins(tx, int64(reverseProxy.Id), resultData)

	}
	return &pb.RPCSuccess{}, err
}
