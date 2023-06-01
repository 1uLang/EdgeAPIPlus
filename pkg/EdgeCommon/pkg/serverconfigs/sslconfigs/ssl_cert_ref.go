package sslconfigs

type SSLCertRef struct {
	IsOn     bool  `yaml:"isOn" json:"isOn"`
	CertId   int64 `yaml:"certId" json:"certId"`
	GmCertId int64 `yaml:"gmCertId" json:"gmCertId,omitempty"`
}
