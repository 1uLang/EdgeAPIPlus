// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package payments

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"net/url"
)

type Payment interface {
	GeneratePayURL(order *accounts.UserOrder, method *accounts.OrderMethod) (url string, err error)
	Verify(formValues url.Values) (orderId int64, err error)
}
