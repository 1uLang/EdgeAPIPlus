package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/rpc/pb"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/sslconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

// SSLCertService SSL证书相关服务
type SSLCertService struct {
	BaseService
}

// CreateSSLCert 创建Cert
func (this *SSLCertService) CreateSSLCert(ctx context.Context, req *pb.CreateSSLCertRequest) (*pb.CreateSSLCertResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.TimeBeginAt < 0 {
		return nil, errors.New("invalid TimeBeginAt")
	}
	if req.TimeEndAt < 0 {
		return nil, errors.New("invalid TimeEndAt")
	}

	certId, err := models.SharedSSLCertDAO.CreateCert(tx, adminId, userId, req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return &pb.CreateSSLCertResponse{SslCertId: certId}, nil
}

// UpdateSSLCert 修改Cert
func (this *SSLCertService) UpdateSSLCert(ctx context.Context, req *pb.UpdateSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.TimeBeginAt < 0 {
		return nil, errors.New("invalid TimeBeginAt")
	}
	if req.TimeEndAt < 0 {
		return nil, errors.New("invalid TimeEndAt")
	}

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(tx, req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedSSLCertDAO.UpdateCert(tx, req.SslCertId, req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledSSLCertConfig 查找证书配置
func (this *SSLCertService) FindEnabledSSLCertConfig(ctx context.Context, req *pb.FindEnabledSSLCertConfigRequest) (*pb.FindEnabledSSLCertConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(tx, req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedSSLCertDAO.ComposeCertConfig(tx, req.SslCertId, nil)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledSSLCertConfigResponse{SslCertJSON: configJSON}, nil
}

// DeleteSSLCert 删除证书
func (this *SSLCertService) DeleteSSLCert(ctx context.Context, req *pb.DeleteSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(tx, req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedSSLCertDAO.DisableSSLCert(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}

	// 停止相关ACME任务
	err = acme.SharedACMETaskDAO.DisableAllTasksWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountSSLCerts 计算匹配的Cert数量
func (this *SSLCertService) CountSSLCerts(ctx context.Context, req *pb.CountSSLCertRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedSSLCertDAO.CountCerts(tx, req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, req.UserId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListSSLCerts 列出单页匹配的Cert
func (this *SSLCertService) ListSSLCerts(ctx context.Context, req *pb.ListSSLCertsRequest) (*pb.ListSSLCertsResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	certIds, err := models.SharedSSLCertDAO.ListCertIds(tx, req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, req.UserId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	certConfigs := []*sslconfigs.SSLCertConfig{}
	for _, certId := range certIds {
		certConfig, err := models.SharedSSLCertDAO.ComposeCertConfig(tx, certId, nil)
		if err != nil {
			return nil, err
		}

		// 这里不需要数据内容
		certConfig.CertData = nil
		certConfig.KeyData = nil

		certConfigs = append(certConfigs, certConfig)
	}
	certConfigsJSON, err := json.Marshal(certConfigs)
	if err != nil {
		return nil, err
	}
	return &pb.ListSSLCertsResponse{SslCertsJSON: certConfigsJSON}, nil
}

// CountAllSSLCertsWithOCSPError 计算有OCSP错误的证书数量
func (this *SSLCertService) CountAllSSLCertsWithOCSPError(ctx context.Context, req *pb.CountAllSSLCertsWithOCSPErrorRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedSSLCertDAO.CountAllSSLCertsWithOCSPError(tx, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListSSLCertsWithOCSPError 列出有OCSP错误的证书
func (this *SSLCertService) ListSSLCertsWithOCSPError(ctx context.Context, req *pb.ListSSLCertsWithOCSPErrorRequest) (*pb.ListSSLCertsWithOCSPErrorResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	certs, err := models.SharedSSLCertDAO.ListSSLCertsWithOCSPError(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbCerts = []*pb.SSLCert{}
	for _, cert := range certs {
		pbCerts = append(pbCerts, &pb.SSLCert{
			Id:            int64(cert.Id),
			IsOn:          cert.IsOn,
			Name:          cert.Name,
			TimeBeginAt:   types.Int64(cert.TimeBeginAt),
			TimeEndAt:     types.Int64(cert.TimeEndAt),
			DnsNames:      cert.DecodeDNSNames(),
			CommonNames:   cert.DecodeCommonNames(),
			IsACME:        cert.IsACME,
			AcmeTaskId:    int64(cert.AcmeTaskId),
			Ocsp:          cert.Ocsp,
			OcspIsUpdated: cert.OcspIsUpdated == 1,
			OcspError:     cert.OcspError,
			Description:   cert.Description,
			IsCA:          cert.IsCA,
			ServerName:    cert.ServerName,
			CreatedAt:     int64(cert.CreatedAt),
			UpdatedAt:     int64(cert.UpdatedAt),
		})
	}

	return &pb.ListSSLCertsWithOCSPErrorResponse{
		SslCerts: pbCerts,
	}, nil
}

// IgnoreSSLCertsWithOCSPError 忽略一组OCSP证书错误
func (this *SSLCertService) IgnoreSSLCertsWithOCSPError(ctx context.Context, req *pb.IgnoreSSLCertsWithOCSPErrorRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedSSLCertDAO.IgnoreSSLCertsWithOCSPError(tx, req.SslCertIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ResetSSLCertsWithOCSPError 重置一组证书OCSP错误状态
func (this *SSLCertService) ResetSSLCertsWithOCSPError(ctx context.Context, req *pb.ResetSSLCertsWithOCSPErrorRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedSSLCertDAO.ResetSSLCertsWithOCSPError(tx, req.SslCertIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ResetAllSSLCertsWithOCSPError 重置所有证书OCSP错误状态
func (this *SSLCertService) ResetAllSSLCertsWithOCSPError(ctx context.Context, req *pb.ResetAllSSLCertsWithOCSPErrorRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedSSLCertDAO.ResetAllSSLCertsWithOCSPError(tx)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ListUpdatedSSLCertOCSP 读取证书的OCSP
func (this *SSLCertService) ListUpdatedSSLCertOCSP(ctx context.Context, req *pb.ListUpdatedSSLCertOCSPRequest) (*pb.ListUpdatedSSLCertOCSPResponse, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	certs, err := models.SharedSSLCertDAO.ListCertOCSPAfterVersion(tx, req.Version, int64(req.Size))
	if err != nil {
		return nil, err
	}

	var result = []*pb.ListUpdatedSSLCertOCSPResponse_SSLCertOCSP{}
	for _, cert := range certs {
		result = append(result, &pb.ListUpdatedSSLCertOCSPResponse_SSLCertOCSP{
			SslCertId: int64(cert.Id),
			Data:      cert.Ocsp,
			ExpiresAt: int64(cert.OcspExpiresAt),
			Version:   int64(cert.OcspUpdatedVersion),
		})
	}

	return &pb.ListUpdatedSSLCertOCSPResponse{
		SslCertOCSP: result,
	}, nil
}

// ------- api 客户定制化接口

type CreateSSLCertAPIRequest struct {
	Username    string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"` //指定用户账号
	IsOn        bool   `protobuf:"varint,2,opt,name=isOn,proto3" json:"isOn,omitempty"`
	Name        string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	ServerName  string `protobuf:"bytes,5,opt,name=serverName,proto3" json:"serverName,omitempty"`
	IsCA        bool   `protobuf:"varint,6,opt,name=isCA,proto3" json:"isCA,omitempty"`
	CertData    string `protobuf:"bytes,7,opt,name=certData,proto3" json:"certData,omitempty"`
	KeyData     string `protobuf:"bytes,8,opt,name=keyData,proto3" json:"keyData,omitempty"`
}

type FindAllCertRequest struct {
	Offset int64 `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"` //偏移量
	Size   int64 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`     //显示条数
}

type FindAllCertResponse struct {
	Certs []*FindAllCertResponse_Cert `protobuf:"bytes,1,rep,name=certs,proto3" json:"certs,omitempty"`  //证书列表
	Total int64                       `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"` //总条数
}
type FindAllCertResponse_Cert struct {
	Id           int64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`                     //证书id
	Name         string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`                  //证书名称
	TimeBeginDay string   `protobuf:"bytes,3,opt,name=timeBeginDay,proto3" json:"timeBeginDay,omitempty"`  //证书生效时间
	TimeEndDay   string   `protobuf:"bytes,4,opt,name=timeEndDay,proto3" json:"timeEndDay,omitempty"`      //证书失效时间
	DnsNames     []string `protobuf:"bytes,5,rep,name=dnsNames,proto3" json:"dnsNames,omitempty"`          //dns
	IsCA         bool     `protobuf:"varint,6,opt,name=isCA,proto3" json:"isCA,omitempty"`                 //是否为CA证书
	CountServers int64    `protobuf:"varint,6,opt,name=countServers,proto3" json:"countServers,omitempty"` //服务引用数
}

// CreateSSLCert 创建Cert
func (this *SSLCertService) CreateSSLCertAPI(ctx context.Context, req *CreateSSLCertAPIRequest) (*pb.CreateSSLCertResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	tx := this.NullTx()
	//通过username找到userId
	if req.Username == "" {
		return nil, fmt.Errorf("username不能为空")
	}
	adminId, err = models.SharedAdminDAO.FindAdminIdWithUsername(tx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("查询用户名%s失败：%s", req.Username, err.Error())
	}
	if adminId == 0 {
		return nil, fmt.Errorf("该用户名%s不存在", req.Username)
	}
	certData, err := base64.StdEncoding.DecodeString(req.CertData)
	if err != nil {
		return nil, fmt.Errorf("证书校验错误：" + err.Error())
	}
	keyData, err := base64.StdEncoding.DecodeString(req.KeyData)
	if err != nil {
		return nil, fmt.Errorf("证书或密钥校验错误：" + err.Error())
	}
	fmt.Println("cert : ", string(certData))
	fmt.Println("key : ", string(keyData))
	// 校验
	sslConfig := &sslconfigs.SSLCertConfig{
		IsCA:     req.IsCA,
		CertData: certData,
		KeyData:  keyData,
	}
	err = sslConfig.Init()
	if err != nil {
		if req.IsCA {
			return nil, fmt.Errorf("证书校验错误：" + err.Error())
		} else {
			return nil, fmt.Errorf("证书或密钥校验错误：" + err.Error())
		}
	}
	fmt.Println(sslConfig.TimeBeginAt, sslConfig.TimeEndAt)
	certId, err := models.SharedSSLCertDAO.CreateCert(tx, adminId, 0, req.IsOn, req.Name, req.Description, req.ServerName,
		req.IsCA, certData, keyData, sslConfig.TimeBeginAt, sslConfig.TimeEndAt, sslConfig.DNSNames, sslConfig.CommonNames)
	if err != nil {
		return nil, err
	}

	return &pb.CreateSSLCertResponse{SslCertId: certId}, nil
}

func (this *SSLCertService) FindAllCert(ctx context.Context, req *FindAllCertRequest) (*FindAllCertResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	tx := this.NullTx()

	result, total, err := models.SharedSSLCertDAO.FindAllCerts(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	resp := &FindAllCertResponse{Total: total}
	for _, cert := range result {
		c := &FindAllCertResponse_Cert{Id: int64(cert.Id), Name: cert.Name, IsCA: cert.IsCA}
		c.DnsNames = []string{}
		if models.IsNotNull(cert.DnsNames) {
			err = json.Unmarshal(cert.DnsNames, &c.DnsNames)
			if err != nil {
				return nil, err
			}
		}

		policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, int64(cert.Id))
		if err != nil {
			return nil, err
		}

		if len(policyIds) == 0 {
			c.CountServers = 0
		} else {
			c.CountServers, err = models.SharedServerDAO.CountAllEnabledServersWithSSLPolicyIds(tx, policyIds)
			if err != nil {
				return nil, err
			}
		}

		c.TimeBeginDay = timeutil.FormatTime("Y-m-d", int64(cert.TimeBeginAt))
		c.TimeEndDay = timeutil.FormatTime("Y-m-d", int64(cert.TimeEndAt))
		resp.Certs = append(resp.Certs, c)
	}
	return resp, nil
}
