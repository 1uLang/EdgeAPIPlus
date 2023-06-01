// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

type ComponentCode = string

const (
	ComponentCodeUser        ComponentCode = "user"
	ComponentCodeScheduling  ComponentCode = "scheduling"
	ComponentCodeMonitor     ComponentCode = "monitor"
	ComponentCodeLog         ComponentCode = "log"
	ComponentCodeReporter    ComponentCode = "reporter"
	ComponentCodePlan        ComponentCode = "plan"
	ComponentCodeFinance     ComponentCode = "finance"
	ComponentCodeNS          ComponentCode = "ns"
	ComponentCodeL2Node      ComponentCode = "l2node"
	ComponentCodeTicket      ComponentCode = "ticket"
	ComponentCodeAntiDDoS    ComponentCode = "antiDDoS"
	ComponentCodeCloudNative ComponentCode = "cloudNative"
)

type Edition = string

const (
	EditionBasic Edition = "basic" // 个人商业版
	EditionPro   Edition = "pro"   // 专业版
	EditionEnt   Edition = "ent"   // 企业版
	EditionMax   Edition = "max"   // [待命名]
	EditionUltra Edition = "ultra" // 旗舰版
)

type ComponentDefinition struct {
	Name        string        `json:"name"`
	Code        ComponentCode `json:"code"`
	Description string        `json:"description"`
}

func FindAllComponents() []*ComponentDefinition {
	return []*ComponentDefinition{
		{
			Name: "多租户",
			Code: ComponentCodeUser,
		},
		{
			Name: "智能调度",
			Code: ComponentCodeScheduling,
		},
		{
			Name: "监控",
			Code: ComponentCodeMonitor,
		},
		{
			Name: "日志",
			Code: ComponentCodeLog,
		},
		{
			Name: "区域监控",
			Code: ComponentCodeReporter,
		},
		{
			Name: "套餐",
			Code: ComponentCodePlan,
		},
		{
			Name: "财务",
			Code: ComponentCodeFinance,
		},
		{
			Name: "智能DNS",
			Code: ComponentCodeNS,
		},
		{
			Name: "L2节点",
			Code: ComponentCodeL2Node,
		},
		{
			Name: "工单系统",
			Code: ComponentCodeTicket,
		},
		{
			Name: "高防IP",
			Code: ComponentCodeAntiDDoS,
		},
		/**{
			Name: "云原生部署",
			Code: ComponentCodeCloudNative,
		},**/
	}
}

func CheckComponent(code string) bool {
	for _, c := range FindAllComponents() {
		if c.Code == code {
			return true
		}
	}
	return false
}
