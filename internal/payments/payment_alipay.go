// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/smartwalle/alipay/v3"
	"net/url"
	"strings"
)

type AlipayPayment struct {
}

func NewAlipayPayment() *AlipayPayment {
	return &AlipayPayment{}
}

// GeneratePayURL 构造支付URL
func (this *AlipayPayment) GeneratePayURL(order *accounts.UserOrder, method *accounts.OrderMethod) (string, error) {
	var params = &userconfigs.AlipayPayMethodParams{}
	err := json.Unmarshal(method.Params, params)
	if err != nil {
		return "", errors.New("decode params: " + err.Error())
	}

	client, err := this.client(params)
	if err != nil {
		return "", err
	}

	var def = userconfigs.FindPresetPayMethodWithCode(userconfigs.PayMethodAlipay)
	if def == nil {
		return "", errors.New("can not find payment method")
	}

	// 用户节点访问地址
	userNodeAddr, err := models.SharedUserNodeDAO.FindUserNodeAccessAddr(nil)
	if err != nil {
		return "", err
	}

	var p = alipay.TradeWapPay{}
	p.NotifyURL = strings.ReplaceAll(def.NotifyURL, "${baseAddr}", userNodeAddr)
	p.ReturnURL = userNodeAddr + "/finance/pay?code=" + order.Code
	p.Subject = userconfigs.FindOrderTypeName(order.Type)
	p.OutTradeNo = order.Code
	p.TotalAmount = fmt.Sprintf("%.2f", order.Amount)
	p.ProductCode = params.ProductCode

	result, err := client.TradeWapPay(p)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// Verify 校验订单
// "gmt_create=2022-09-28+08%3A39%3A15&charset=utf-8&seller_email=thpabq1166%40sandbox.com&subject=%E5%85%85%E5%80%BC&sign=PAu%2BDHv2EerXiE4fUTfwTXWL7g4N%2BVv2zBcr8jw%2FJd2ciNsDyayqMudSNcvDzWGPwIC6P0fFqGPuoA4jod3lbCXS0uPOWANfkG%2BOS8OAIweyeNwEz%2B6G6HjMIDvD07wNMe5NKOnNc82UXOY4sIYo%2Fgfbtgw96CdUWxvAjEWTNSld4YExUf6y%2BjHZqaU%2FijNIku4q3u2UsUxszf81q9HsgOVRXuk9BTjKcgK3Eqbf5LwSF2zFYXdmUSvbGJWzl6uIsWtZMeRfMmtkGZbNIqRMLDjb9QbMv2o4foWF9EoZ%2Bev3cjDkHrzJtcpyUEyf1QobqOhP%2FEZ0Hm0R4blscAt2qQ%3D%3D&buyer_id=2088622987776411&invoice_amount=0.01&notify_id=2022092800222083918076410521014666&fund_bill_list=%5B%7B%22amount%22%3A%220.01%22%2C%22fundChannel%22%3A%22ALIPAYACCOUNT%22%7D%5D&notify_type=trade_status_sync&trade_status=TRADE_SUCCESS&receipt_amount=0.01&buyer_pay_amount=0.01&app_id=2021000121667766&sign_type=RSA2&seller_id=2088621993221476&gmt_payment=2022-09-28+08%3A39%3A17&notify_time=2022-09-28+08%3A39%3A19&version=1.0&out_trade_no=10001&total_amount=0.01&trade_no=2022092822001476410502203660&auth_app_id=2021000121667766&buyer_logon_id=phy***%40sandbox.com&point_amount=0.00"
func (this *AlipayPayment) Verify(formValues url.Values) (orderId int64, err error) {
	var orderCode = formValues.Get("out_trade_no")
	if len(orderCode) == 0 {
		return 0, errors.New("'out_trade_no' required")
	}

	var tradeStatus = formValues.Get("trade_status")
	if tradeStatus != "TRADE_SUCCESS" {
		return 0, errors.New("failed to pay: trade status: " + tradeStatus)
	}

	var tx *dbs.Tx
	order, err := accounts.SharedUserOrderDAO.FindUserOrderWithCode(tx, orderCode)
	if err != nil {
		return
	}

	if order == nil {
		return 0, errors.New("could not find order with code '" + orderCode + "'")
	}

	orderId = int64(order.Id)

	method, err := accounts.SharedOrderMethodDAO.FindEnabledOrderMethod(tx, int64(order.MethodId))
	if err != nil {
		return 0, err
	}

	if method == nil {
		return 0, errors.New("can not find payment method")
	}

	if method.ParentCode != userconfigs.PayMethodAlipay {
		return 0, errors.New("invalid method parent code '" + method.ParentCode + "', should be '" + userconfigs.PayMethodAlipay + "'")
	}

	var params = &userconfigs.AlipayPayMethodParams{}
	err = json.Unmarshal(method.Params, params)
	if err != nil {
		return 0, errors.New("decode params: " + err.Error())
	}

	client, err := this.client(params)
	if err != nil {
		return 0, err
	}

	ok, err := client.VerifySign(formValues)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errors.New("verify failed")
	}

	return
}

// 获取客户端
// TODO 考虑重用性
func (this *AlipayPayment) client(params *userconfigs.AlipayPayMethodParams) (*alipay.Client, error) {
	client, err := alipay.New(params.AppId, params.PrivateKey, !params.IsSandbox)
	if err != nil {
		return nil, err
	}

	err = client.LoadAppPublicCert(params.AppPublicCert)
	if err != nil {
		return nil, errors.New("load app public cert failed: " + err.Error())
	}

	err = client.LoadAliPayRootCert(params.AlipayRootCert)
	if err != nil {
		return nil, errors.New("load alipay root cert failed: " + err.Error())
	}

	err = client.LoadAliPayPublicCert(params.AlipayPublicCert)
	if err != nil {
		return nil, errors.New("load alipay public cert failed: " + err.Error())
	}
	return client, nil
}
