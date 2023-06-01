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

// QyWeixinRobotMedia 企业微信群机器人媒介
type QyWeixinRobotMedia struct {
	WebHookURL string     `yaml:"webHookURL" json:"webHookURL"`
	TextFormat TextFormat `yaml:"textFormat" json:"textFormat"`
}

// NewQyWeixinRobotMedia 获取新对象
func NewQyWeixinRobotMedia() *QyWeixinRobotMedia {
	return &QyWeixinRobotMedia{}
}

func (this *QyWeixinRobotMedia) Send(user string, subject string, body string) (resp []byte, err error) {
	if len(this.WebHookURL) == 0 {
		return nil, errors.New("webHook url should not be empty")
	}

	mobiles := []string{}
	if len(user) > 0 {
		for _, u := range strings.Split(user, ",") {
			u = strings.TrimSpace(u)
			if len(u) > 0 {
				mobiles = append(mobiles, u)
			}
		}
	}

	content := maps.Map{}
	if this.TextFormat == FormatMarkdown { // markdown
		content = maps.Map{
			"msgtype": "markdown",
			"markdown": maps.Map{
				"content":               subject + "\n" + body,
				"mentioned_mobile_list": mobiles,
			},
		}
	} else {
		content = maps.Map{
			"msgtype": "text",
			"text": maps.Map{
				"content":               subject + "\n" + body,
				"mentioned_mobile_list": mobiles,
			},
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
func (this *QyWeixinRobotMedia) RequireUser() bool {
	return false
}
