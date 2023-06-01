package gmconfigs

import (
	"context"
	"errors"
	"github.com/1uLang/gmsm/gmtls"
	"github.com/1uLang/gmsm/x509"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/lists"
	"reflect"
	"time"
)

// GMCertConfig 国密证书
type GMCertConfig struct {
	Id           int64  `yaml:"id" json:"id"`
	IsOn         bool   `yaml:"isOn" json:"isOn"`
	Name         string `yaml:"name" json:"name"`
	Description  string `yaml:"description" json:"description"`   // 说明
	SignCertData []byte `yaml:"signCertData" json:"signCertData"` // 证书数据
	SignKeyData  []byte `yaml:"signKeyData" json:"signKeyData"`   // 密钥数据
	EncCertData  []byte `yaml:"encCertData" json:"encCertData"`   // 证书数据
	EncKeyData   []byte `yaml:"encKeyData" json:"encKeyData"`     // 密钥数据
	ServerName   string `yaml:"serverName" json:"serverName"`     // 证书使用的主机名，在请求TLS服务器时需要

	// 以下是从证书中分析所得
	TimeBeginAt int64    `yaml:"timeBeginAt" json:"timeBeginAt"`
	TimeEndAt   int64    `yaml:"timeEndAt" json:"timeEndAt"`
	DNSNames    []string `yaml:"dnsNames" json:"dnsNames"`
	CommonNames []string `yaml:"commonNames" json:"commonNames"`

	signCert  *gmtls.Certificate
	encCert   *gmtls.Certificate
	timeBegin time.Time
	timeEnd   time.Time
}

// Init 校验
func (this *GMCertConfig) Init(ctx context.Context) error {
	// 如果没有指定数据， 则从ctx中读取数据
	if ctx != nil && len(this.SignCertData) < 128 {
		var dataMapOne = ctx.Value("DataMap")
		if dataMapOne != nil && !reflect.ValueOf(dataMapOne).IsNil() {
			dataMap, ok := dataMapOne.(*shared.DataMap)
			if !ok {
				return errors.New("GMCertConfig.init(): invalid 'DataMap' in context")
			}
			if dataMap != nil { // 再次检查是否为nil
				this.SignKeyData = dataMap.Read(this.SignKeyData)
				this.SignCertData = dataMap.Read(this.SignCertData)
			}
		}
	}
	// 如果没有指定数据， 则从ctx中读取数据
	if ctx != nil && len(this.EncCertData) < 128 {
		var dataMapOne = ctx.Value("DataMap")
		if dataMapOne != nil && !reflect.ValueOf(dataMapOne).IsNil() {
			dataMap, ok := dataMapOne.(*shared.DataMap)
			if !ok {
				return errors.New("GMCertConfig.init(): invalid 'DataMap' in context")
			}
			if dataMap != nil { // 再次检查是否为nil
				this.EncKeyData = dataMap.Read(this.EncKeyData)
				this.EncCertData = dataMap.Read(this.EncCertData)
			}
		}
	}

	var commonNames []string // 发行组织
	var dnsNames []string    // 域名

	signCert, err := gmtls.X509KeyPair(this.SignCertData, this.SignKeyData)
	if err != nil {
		return errors.New("load sign certificate failed:" + err.Error())
	}
	encCert, err := gmtls.X509KeyPair(this.EncCertData, this.EncKeyData)
	if err != nil {
		return errors.New("load enc certificate  failed:" + err.Error())
	}

	for index, data := range signCert.Certificate {
		c, err := x509.ParseCertificate(data)
		if err != nil {
			continue
		}

		for _, dnsName := range c.DNSNames {
			if !lists.ContainsString(dnsNames, dnsName) {
				dnsNames = append(dnsNames, dnsName)
			}
		}

		commonNames = append(commonNames, c.Issuer.CommonName)

		if index == 0 {
			this.timeBegin = c.NotBefore
			this.timeEnd = c.NotAfter
		}
	}

	this.signCert = &signCert
	this.encCert = &encCert

	// 赋值分析结果
	this.DNSNames = dnsNames
	this.CommonNames = commonNames
	this.TimeBeginAt = this.timeBegin.Unix()
	this.TimeEndAt = this.timeEnd.Unix()

	return nil
}

// MatchDomain 校验是否匹配某个域名
func (this *GMCertConfig) MatchDomain(domain string) bool {
	if len(this.DNSNames) == 0 {
		return false
	}
	return configutils.MatchDomains(this.DNSNames, domain)
}

// CertObject 获取证书对象
func (this *GMCertConfig) CertObject() (*gmtls.Certificate, *gmtls.Certificate) {
	return this.signCert, this.encCert
}

// TimeBegin 开始时间
func (this *GMCertConfig) TimeBegin() time.Time {
	return this.timeBegin
}

// TimeEnd 结束时间
func (this *GMCertConfig) TimeEnd() time.Time {
	return this.timeEnd
}
