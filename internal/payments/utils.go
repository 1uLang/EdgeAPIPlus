// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package payments

import (
	"crypto/sha256"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/types"
	"net/url"
	"sort"
	"strings"
	"time"
)

// GeneratePayURL PayURL 构造支付URL
func GeneratePayURL(order *accounts.UserOrder, method *accounts.OrderMethod) (payURL string, err error) {
	if method == nil || !method.IsOn {
		return "", errors.New("invalid method with id '" + types.String(method.Id) + "'")
	}

	// 内置支付方式
	switch method.ParentCode {
	case userconfigs.PayMethodAlipay: // 支付宝
		return NewAlipayPayment().GeneratePayURL(order, method)
	}

	// 自定义
	var args = []string{}
	args = append(args, "EdgeOrderMethod="+url.QueryEscape(method.Code))
	args = append(args, "EdgeOrderCode="+url.QueryEscape(order.Code))
	args = append(args, "EdgeOrderTimestamp="+types.String(time.Now().Unix()))
	args = append(args, "EdgeOrderAmount="+fmt.Sprintf("%.2f", order.Amount))

	sort.Strings(args)

	var signArgs = append([]string{}, args...)
	signArgs = append(signArgs, method.Secret)
	var sign = fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(signArgs, "&"))))
	args = append(args, "EdgeOrderSign="+sign)

	if strings.Contains(method.Url, "?") {
		return method.Url + "&" + strings.Join(args, "&"), nil
	}
	return method.Url + "?" + strings.Join(args, "&"), nil
}

func Verify(payMethod userconfigs.PayMethod, formValues url.Values) (orderId int64, err error) {
	switch payMethod {
	case userconfigs.PayMethodAlipay: // 支付宝
		return NewAlipayPayment().Verify(formValues)
	}

	return 0, errors.New("invalid payment method '" + payMethod + "'")
}
