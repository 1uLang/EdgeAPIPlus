// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsla

import (
	"errors"
	"github.com/iwind/TeaGo/types"
)

type BaseResponse struct {
	Status struct {
		Code    int    `json:"code"`
		Name    string `json:"name"`
		Message string `json:"message"`
	} `json:"status"`
}

func (this *BaseResponse) Success() bool {
	return this.Status.Code == 300
}

func (this *BaseResponse) Error() error {
	return errors.New("code:" + types.String(this.Status.Code) + ", name:" + this.Status.Name + ", message:" + this.Status.Message)
}
