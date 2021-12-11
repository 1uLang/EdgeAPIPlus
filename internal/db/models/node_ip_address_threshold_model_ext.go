package models

import (
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
)

func (this *NodeIPAddressThreshold) DecodeItems() (result []*nodeconfigs.IPAddressThresholdItemConfig) {
	if len(this.Items) == 0 {
		return
	}

	err := json.Unmarshal([]byte(this.Items), &result)
	if err != nil {
		remotelogs.Error("NodeIPAddressThreshold", "decode items: "+err.Error())
	}
	return
}

func (this *NodeIPAddressThreshold) DecodeActions() (result []*nodeconfigs.IPAddressThresholdActionConfig) {
	if len(this.Actions) == 0 {
		return
	}

	err := json.Unmarshal([]byte(this.Actions), &result)
	if err != nil {
		remotelogs.Error("NodeIPAddressThreshold", "decode actions: "+err.Error())
	}
	return
}
