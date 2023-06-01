//go:build plus

package senders

import (
	"bytes"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/utils/string"
	"io"
	"net/http"
	"strings"
	"time"
)

// 钉钉群机器人媒介
type DingTalkMedia struct {
	WebHookURL string `yaml:"webHookURL" json:"webHookURL"`
}

// 获取新对象
func NewDingTalkMedia() *DingTalkMedia {
	return &DingTalkMedia{}
}

func (this *DingTalkMedia) Send(user string, subject string, body string) (resp []byte, err error) {
	if len(this.WebHookURL) == 0 {
		return nil, errors.New("webHook url should not be empty")
	}

	content := maps.Map{
		"msgtype": "text",
		"text": maps.Map{
			"content": "标题：" + subject + "\n内容：" + body,
		},
	}
	if len(user) > 0 {
		mobiles := []string{}
		for _, u := range strings.Split(user, ",") {
			u = strings.TrimSpace(u)
			if len(u) > 0 {
				mobiles = append(mobiles, u)
			}
		}

		content["at"] = maps.Map{
			"atMobiles": mobiles,
		}
	}

	reader := bytes.NewBufferString(stringutil.JSONEncode(content))
	req, err := http.NewRequest(http.MethodPost, this.WebHookURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := utils.SharedHttpClient(5 * time.Second)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	resp, err = io.ReadAll(response.Body)
	return
}

// RequireUser 是否需要用户标识
func (this *DingTalkMedia) RequireUser() bool {
	return false
}
