// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus
// +build plus

package models

import (
	"encoding/json"
	"errors"
	"github.com/1uLang/EdgeCommon/pkg/dnsconfigs"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs/ddosconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

// CreateCluster 创建集群
func (this *NSClusterDAO) CreateCluster(tx *dbs.Tx, name string, email string, accessLogRefJSON []byte, hosts []string, soaConfig *dnsconfigs.NSSOAConfig) (int64, error) {
	var op = NewNSClusterOperator()
	op.Name = name

	if len(accessLogRefJSON) > 0 {
		op.AccessLog = accessLogRefJSON
	}

	op.IsOn = true
	op.State = NSClusterStateEnabled

	// 默认端口
	// TCP
	{
		var config = &serverconfigs.TCPProtocolConfig{}
		config.IsOn = true
		config.Listen = []*serverconfigs.NetworkAddressConfig{
			{
				Protocol:  serverconfigs.ProtocolTCP,
				PortRange: "53",
			},
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return 0, err
		}
		op.Tcp = configJSON
	}

	// UDP
	{
		var config = &serverconfigs.UDPProtocolConfig{}
		config.IsOn = true
		config.Listen = []*serverconfigs.NetworkAddressConfig{
			{
				Protocol:  serverconfigs.ProtocolUDP,
				PortRange: "53",
			},
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return 0, err
		}
		op.Udp = configJSON
	}

	// hosts
	if hosts == nil {
		hosts = []string{}
	}
	hostsJSON, err := json.Marshal(hosts)
	if err != nil {
		return 0, err
	}
	op.Hosts = hostsJSON

	// SOA
	if soaConfig == nil {
		soaConfig = dnsconfigs.DefaultNSSOAConfig()
	}
	soaJSON, err := json.Marshal(soaConfig)
	if err != nil {
		return 0, err
	}
	op.Soa = soaJSON
	op.SoaSerial = time.Now().Unix()

	// email
	op.Email = email

	return this.SaveInt64(tx, op)
}

// UpdateCluster 修改集群
func (this *NSClusterDAO) UpdateCluster(tx *dbs.Tx, clusterId int64, name string, email string, hosts []string, isOn bool, timeZone string, autoRemoteStart bool) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}

	oldCluster, err := this.FindEnabledNSCluster(tx, clusterId)
	if err != nil {
		return err
	}
	if oldCluster == nil {
		return errors.New("cluster not found")
	}

	var op = NewNSClusterOperator()
	op.Id = clusterId
	op.Name = name
	op.Email = email
	op.IsOn = isOn

	// hosts
	if hosts == nil {
		hosts = []string{}
	}
	hostsJSON, err := json.Marshal(hosts)
	if err != nil {
		return err
	}
	op.Hosts = hostsJSON

	// 检查Hosts和SOA配置是否变更
	if !utils.EqualConfig(oldCluster.DecodeHosts(), hosts) || oldCluster.Email != email {
		op.SoaSerial = time.Now().Unix()
	}

	op.TimeZone = timeZone
	op.AutoRemoteStart = autoRemoteStart

	err = this.Save(tx, op)
	if err != nil {
		return nil
	}

	return this.NotifyUpdate(tx, clusterId)
}

// UpdateClusterAccessLog 设置访问日志
func (this *NSClusterDAO) UpdateClusterAccessLog(tx *dbs.Tx, clusterId int64, accessLogJSON []byte) error {
	return this.Query(tx).
		Pk(clusterId).
		Set("accessLog", accessLogJSON).
		UpdateQuickly()
}

// FindClusterAccessLog 读取访问日志配置
func (this *NSClusterDAO) FindClusterAccessLog(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	accessLog, err := this.Query(tx).
		Pk(clusterId).
		Result("accessLog").
		FindStringCol("")
	return []byte(accessLog), err
}

// UpdateRecursion 设置递归DNS
func (this *NSClusterDAO) UpdateRecursion(tx *dbs.Tx, clusterId int64, recursionJSON []byte) error {
	err := this.Query(tx).
		Pk(clusterId).
		Set("recursion", recursionJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterRecursion 读取递归DNS配置
func (this *NSClusterDAO) FindClusterRecursion(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	recursion, err := this.Query(tx).
		Result("recursion").
		Pk(clusterId).
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	return []byte(recursion), nil
}

// FindClusterTCP 查找集群的TCP设置
func (this *NSClusterDAO) FindClusterTCP(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("tcp").
		FindBytesCol()
}

// UpdateClusterTCP 修改集群的TCP设置
func (this *NSClusterDAO) UpdateClusterTCP(tx *dbs.Tx, clusterId int64, tcpConfig *serverconfigs.TCPProtocolConfig) error {
	tcpJSON, err := json.Marshal(tcpConfig)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("tcp", tcpJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterTLS 查找集群的TLS设置
func (this *NSClusterDAO) FindClusterTLS(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("tls").
		FindBytesCol()
}

// UpdateClusterTLS 修改集群的TLS设置
func (this *NSClusterDAO) UpdateClusterTLS(tx *dbs.Tx, clusterId int64, tlsConfig *serverconfigs.TLSProtocolConfig) error {
	tlsJSON, err := json.Marshal(tlsConfig)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("tls", tlsJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterUDP 查找集群的TCP设置
func (this *NSClusterDAO) FindClusterUDP(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("udp").
		FindBytesCol()
}

// UpdateClusterUDP 修改集群的UDP设置
func (this *NSClusterDAO) UpdateClusterUDP(tx *dbs.Tx, clusterId int64, udpConfig *serverconfigs.UDPProtocolConfig) error {
	udpJSON, err := json.Marshal(udpConfig)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("udp", udpJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterHosts 查找集群的DNS主机域名
func (this *NSClusterDAO) FindClusterHosts(tx *dbs.Tx, clusterId int64) ([]string, error) {
	one, err := this.Query(tx).
		Result("hosts").
		Pk(clusterId).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*NSCluster).DecodeHosts(), nil
}

// FindClusterDDoSProtection 获取集群的DDoS设置
func (this *NSClusterDAO) FindClusterDDoSProtection(tx *dbs.Tx, clusterId int64) (*ddosconfigs.ProtectionConfig, error) {
	one, err := this.Query(tx).
		Result("ddosProtection").
		Pk(clusterId).
		Find()
	if one == nil || err != nil {
		return nil, err
	}

	return one.(*NSCluster).DecodeDDoSProtection(), nil
}

// UpdateClusterDDoSProtection 设置集群的DDoS设置
func (this *NSClusterDAO) UpdateClusterDDoSProtection(tx *dbs.Tx, clusterId int64, ddosProtection *ddosconfigs.ProtectionConfig) error {
	if clusterId <= 0 {
		return ErrNotFound
	}

	var op = NewNSClusterOperator()
	op.Id = clusterId

	if ddosProtection == nil {
		op.DdosProtection = "{}"
	} else {
		ddosProtectionJSON, err := json.Marshal(ddosProtection)
		if err != nil {
			return err
		}
		op.DdosProtection = ddosProtectionJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleDNS, clusterId, 0, NSNodeTaskTypeDDosProtectionChanged)
}

// FindClusterAnswer 查询应答模式
func (this *NSClusterDAO) FindClusterAnswer(tx *dbs.Tx, clusterId int64) (*dnsconfigs.NSAnswerConfig, error) {
	answerJSON, err := this.Query(tx).
		Pk(clusterId).
		Result("answer").
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	var config = dnsconfigs.DefaultNSAnswerConfig()
	if IsNull(answerJSON) {
		return config, nil
	}

	err = json.Unmarshal(answerJSON, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// UpdateClusterAnswer 设置应答模式
func (this *NSClusterDAO) UpdateClusterAnswer(tx *dbs.Tx, clusterId int64, answerConfig *dnsconfigs.NSAnswerConfig) error {
	if clusterId <= 0 {
		return nil
	}

	if answerConfig == nil {
		return nil
	}

	answerJSON, err := json.Marshal(answerConfig)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Pk(clusterId).
		Set("answer", answerJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterSOA 查找SOA配置
func (this *NSClusterDAO) FindClusterSOA(tx *dbs.Tx, clusterId int64) (*dnsconfigs.NSSOAConfig, error) {
	soaJSON, err := this.Query(tx).
		Pk(clusterId).
		Result("soa").
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	var config = dnsconfigs.DefaultNSSOAConfig()
	if IsNull(soaJSON) {
		return config, nil
	}

	err = json.Unmarshal(soaJSON, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// UpdateClusterSOA 修改SOA配置
func (this *NSClusterDAO) UpdateClusterSOA(tx *dbs.Tx, clusterId int64, soaConfig *dnsconfigs.NSSOAConfig) error {
	if clusterId <= 0 {
		return nil
	}

	if soaConfig == nil {
		return nil
	}

	oldSOAConfig, err := this.FindClusterSOA(tx, clusterId)
	if err != nil {
		return err
	}

	// 如果相同则不修改
	if utils.EqualConfig(soaConfig, oldSOAConfig) {
		return nil
	}

	soaJSON, err := json.Marshal(soaConfig)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Pk(clusterId).
		Set("soa", soaJSON).
		Set("soaSerial", time.Now().Unix()).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, clusterId)
}
