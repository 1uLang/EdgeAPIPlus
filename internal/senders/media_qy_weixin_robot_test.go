//go:build plus

package senders

import (
	"testing"
)

func TestQyWeixinRobotMedia_Send(t *testing.T) {
	media := NewQyWeixinRobotMedia()
	media.WebHookURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=123456" //需要换成你自己的webhook
	media.TextFormat = FormatText
	resp, err := media.Send("", "这是标题", "*这是内容*")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("resp:", string(resp))
}