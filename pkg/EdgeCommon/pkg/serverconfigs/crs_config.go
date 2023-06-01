/*
   @Author: 1usir
   @Description:
   @File: crs_config.go
   @Version: 1.0.0
   @Date: 2023/5/24 11:39
*/

package serverconfigs

// CRSConfig CRS配置
type CRSConfig struct {
	IsOn bool `yaml:"isOn" json:"isOn"`
}

func NewCRSConfig() *CRSConfig {
	return &CRSConfig{
		IsOn: true,
	}
}

func (this *CRSConfig) Init() error {
	return nil
}
