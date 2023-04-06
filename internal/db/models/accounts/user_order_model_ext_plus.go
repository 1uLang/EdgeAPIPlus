package accounts

import (
	"crypto/sha256"
	"fmt"
	"github.com/1uLang/EdgeCommon/pkg/userconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"net/url"
	"sort"
	"strings"
	"time"
)

// IsExpired 检查当前订单是否已经过期
func (this *UserOrder) IsExpired() bool {
	return this.Status == userconfigs.OrderStatusNone &&
		time.Now().Unix() > int64(this.ExpiredAt)
}

// PayURL 构造URL
func (this *UserOrder) PayURL() (string, error) {
	var tx *dbs.Tx
	method, err := SharedOrderMethodDAO.FindEnabledBasicOrderMethod(tx, int64(this.MethodId))
	if err != nil {
		return "", err
	}

	if method == nil || !method.IsOn {
		return "", errors.New("invalid method with id '" + types.String(this.MethodId) + "'")
	}

	var args = []string{}
	args = append(args, "EdgeOrderMethod="+url.QueryEscape(method.Code))
	args = append(args, "EdgeOrderCode="+url.QueryEscape(this.Code))
	args = append(args, "EdgeOrderTimestamp="+types.String(time.Now().Unix()))
	args = append(args, "EdgeOrderAmount="+types.String(this.Amount))

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
