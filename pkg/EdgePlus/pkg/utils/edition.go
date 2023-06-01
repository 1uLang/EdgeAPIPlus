// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils

type EditionDefinition struct {
	Name        string        `json:"name"`
	Code        ComponentCode `json:"code"`
	Description string        `json:"description"`
}

func FindAllEditions() []*EditionDefinition {
	return []*EditionDefinition{
		{
			Name: "个人商业版",
			Code: EditionBasic,
		},
		{
			Name: "专业版",
			Code: EditionPro,
		},
		{
			Name: "企业版",
			Code: EditionEnt,
		},
		{
			Name: "豪华版",
			Code: EditionMax,
		},
		{
			Name: "旗舰版",
			Code: EditionUltra,
		},
	}
}

func CheckEdition(edition Edition) bool {
	for _, e := range FindAllEditions() {
		if e.Code == edition {
			return true
		}
	}
	return false
}

func CompareEdition(edition1 Edition, edition2 Edition) int {
	var index1 = -1
	var index2 = -1

	for index, edition := range FindAllEditions() {
		if edition.Code == edition1 {
			index1 = index
		}
		if edition.Code == edition2 {
			index2 = index
		}
	}
	if index2 > index1 {
		return -1
	}
	if index2 == index1 {
		return 0
	}
	return 1
}

func EditionName(edition Edition) string {
	if len(edition) == 0 {
		return ""
	}
	for _, e := range FindAllEditions() {
		if e.Code == edition {
			return e.Name
		}
	}
	return ""
}
