//go:build plus

package senders

import (
	"encoding/json"
	"errors"
	"github.com/iwind/TeaGo/maps"
	"reflect"
)

// MediaType 通知媒介类型
type MediaType = string

const (
	MediaTypeEmail         MediaType = "email"
	MediaTypeWebHook       MediaType = "webHook"
	MediaTypeScript        MediaType = "script"
	MediaTypeDingTalk      MediaType = "dingTalk"
	MediaTypeQyWeixin      MediaType = "qyWeixin"
	MediaTypeQyWeixinRobot MediaType = "qyWeixinRobot"
	MediaTypeAliyunSms     MediaType = "aliyunSms"
	MediaTypeTelegram      MediaType = "telegram"
)

// AllMediaTypes 所有媒介
func AllMediaTypes() []maps.Map {
	return []maps.Map{
		{
			"name":         "邮件",
			"code":         MediaTypeEmail,
			"supportsHTML": true,
			"instance":     new(EmailMedia),
			"description":  "通过邮件发送通知",
			"user":         "接收人邮箱地址",
		},
		{
			"name":         "WebHook",
			"code":         MediaTypeWebHook,
			"supportsHTML": false,
			"instance":     new(WebHookMedia),
			"description":  "通过HTTP请求发送通知",
			"user":         "通过${MessageUser}参数传递到URL上",
		},
		{
			"name":         "脚本",
			"code":         MediaTypeScript,
			"supportsHTML": false,
			"instance":     new(ScriptMedia),
			"description":  "通过运行脚本发送通知",
			"user":         "可以在脚本中使用${MessageUser}来获取这个标识",
		},
		{
			"name":         "钉钉群机器人",
			"code":         MediaTypeDingTalk,
			"supportsHTML": false,
			"instance":     new(DingTalkMedia),
			"description":  "通过钉钉群机器人发送通知消息",
			"user":         "要At（@）的群成员的手机号，多个手机号用英文逗号隔开，也可以为空",
		},
		{
			"name":         "企业微信应用",
			"code":         MediaTypeQyWeixin,
			"supportsHTML": false,
			"instance":     new(QyWeixinMedia),
			"description":  "通过企业微信应用发送通知消息",
			"user":         "接收消息的成员的用户账号，多个成员用竖线（|）分隔，如果所有成员使用@all。留空表示所有成员。",
		},
		{
			"name":         "企业微信群机器人",
			"code":         MediaTypeQyWeixinRobot,
			"supportsHTML": false,
			"instance":     new(QyWeixinRobotMedia),
			"description":  "通过微信群机器人发送通知消息",
			"user":         "要At（@）的群成员的手机号，多个手机号用英文逗号隔开，也可以为空",
		},
		{
			"name":         "阿里云短信",
			"code":         MediaTypeAliyunSms,
			"supportsHTML": false,
			"instance":     new(AliyunSmsMedia),
			"description":  "通过<a href=\"https://www.aliyun.com/product/sms?spm=5176.11533447.1097531.2.12055cfa6UnIix\" target=\"_blank\">阿里云短信服务</a>发送短信",
			"user":         "接收消息的手机号",
		},
		{
			"name":         "Telegram机器人",
			"code":         MediaTypeTelegram,
			"supportsHTML": false,
			"instance":     new(TelegramMedia),
			"description":  "通过机器人向群或者某个用户发送消息，需要确保所在网络能够访问Telegram API服务",
			"user":         "群或用户的Chat ID，通常是一个数字，可以通过和 @get_id_bot 建立对话并发送任意消息获得",
		},
	}
}

// FindMediaType 查找媒介类型
func FindMediaType(mediaType string) maps.Map {
	for _, m := range AllMediaTypes() {
		if m["code"] == mediaType {
			return m
		}
	}
	return nil
}

// NewMediaInstance 查找媒介实例
func NewMediaInstance(mediaType string, optionsJSON []byte) (MediaInterface, error) {
	for _, m := range AllMediaTypes() {
		if m["code"] == mediaType {
			var media = reflect.New(reflect.TypeOf(m["instance"]).Elem()).Interface().(MediaInterface)
			if len(optionsJSON) > 0 {
				err := json.Unmarshal(optionsJSON, media)
				if err != nil {
					return nil, errors.New("decode media options failed: " + err.Error())
				}
			}
			return media, nil
		}
	}
	return nil, errors.New("can not find media with type '" + mediaType + "'")
}

// FindMediaTypeName 查找媒介类型名称
func FindMediaTypeName(mediaType string) string {
	m := FindMediaType(mediaType)
	if m == nil {
		return ""
	}
	return m["name"].(string)
}

// MediaInterface 媒介接口
type MediaInterface interface {
	// Send 发送
	Send(user string, subject string, body string) (resp []byte, err error)

	// RequireUser 是否可以需要用户标识
	RequireUser() bool
}
