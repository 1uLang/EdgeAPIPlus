package models

import "github.com/iwind/TeaGo/dbs"

// GMCert GM证书
type GMCert struct {
	Id           uint32   `field:"id"`           // ID
	AdminId      uint32   `field:"adminId"`      // 管理员ID
	UserId       uint32   `field:"userId"`       // 用户ID
	State        uint8    `field:"state"`        // 状态
	CreatedAt    uint64   `field:"createdAt"`    // 创建时间
	UpdatedAt    uint64   `field:"updatedAt"`    // 修改时间
	IsOn         bool     `field:"isOn"`         // 是否启用
	Name         string   `field:"name"`         // 证书名
	Description  string   `field:"description"`  // 描述
	SignCertData []byte   `field:"signCertData"` // 签名证书内容
	SignKeyData  []byte   `field:"signKeyData"`  // 签名密钥内容
	EncCertData  []byte   `field:"encCertData"`  // 加密证书内容
	EncKeyData   []byte   `field:"encKeyData"`   // 加密密钥内容
	ServerName   string   `field:"serverName"`   // 证书使用的主机名
	GroupIds     dbs.JSON `field:"groupIds"`     // 证书分组
	TimeBeginAt  uint64   `field:"timeBeginAt"`  // 开始时间
	TimeEndAt    uint64   `field:"timeEndAt"`    // 结束时间
	DnsNames     dbs.JSON `field:"dnsNames"`     // DNS名称列表
	CommonNames  dbs.JSON `field:"commonNames"`  // 发行单位列表
	NotifiedAt   uint64   `field:"notifiedAt"`   // 最后通知时间
}

type GMCertOperator struct {
	Id           interface{} // ID
	AdminId      interface{} // 管理员ID
	UserId       interface{} // 用户ID
	State        interface{} // 状态
	CreatedAt    interface{} // 创建时间
	UpdatedAt    interface{} // 修改时间
	IsOn         interface{} // 是否启用
	Name         interface{} // 证书名
	Description  interface{} // 描述
	SignCertData interface{} // 签名证书内容
	SignKeyData  interface{} // 签名密钥内容
	EncCertData  interface{} // 加密证书内容
	EncKeyData   interface{} // 加密密钥内容
	ServerName   interface{} // 证书使用的主机名
	GroupIds     interface{} // 证书分组
	TimeBeginAt  interface{} // 开始时间
	TimeEndAt    interface{} // 结束时间
	DnsNames     interface{} // DNS名称列表
	CommonNames  interface{} // 发行单位列表
	NotifiedAt   interface{} // 最后通知时间
}

func NewGMCertOperator() *GMCertOperator {
	return &GMCertOperator{}
}
