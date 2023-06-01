//go:build plus

package senders_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/senders"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestEmailMedia_Send(t *testing.T) {
	dbs.NotifyReady()

	var media = senders.NewEmailMedia()
	media.SMTP = "smtp.qq.com:587"
	media.Username = "19644627@qq.com"
	media.Password = "123456" // 换成你的邮件密码或者授权码
	media.From = "19644627@qq.com"
	//media.FromName = "测试员"
	//media.FromName = "\"测试员\""
	//var subject = "This is test subject"
	var subject = "这是中文标题"
	_, err := media.Send("iwind.liu@gmail.com", subject, "This is a test body <strong>粗体哦</strong><br/>换行哦")
	if err != nil {
		t.Fatal(err)
	}
}

func TestEmailMedia_SendMails(t *testing.T) {
	dbs.NotifyReady()

	var media = senders.NewEmailMedia()
	media.SMTP = "smtp.qq.com:587"
	media.Username = "19644627@qq.com"
	media.Password = "123456" // 换成你的邮件密码或者授权码
	media.From = "19644627@qq.com"
	//media.FromName = "测试员"
	//media.FromName = "\"测试员\""
	err := media.SendMails([]*senders.MailInfo{
		{
			To:      "iwind.liu@gmail.com",
			Subject: "This is test subject",
			Body:    "This is a test body <strong>粗体哦</strong><br/>换行哦",
		},
		{
			To:      "iwind.liu@gmail.com",
			Subject: "This is test subject 2",
			Body:    "This is a test body 2 <strong>粗体哦</strong><br/>换行哦",
		},
		{
			To:      "q@yun4s.cn",
			Subject: "This is test subject 3",
			Body:    "This is a test body 3 <strong>粗体哦</strong><br/>换行哦",
		},
		{
			To:      "root@teaos.cn",
			Subject: "This is test subject 4",
			Body:    "This is a test body 4 <strong>粗体哦</strong><br/>换行哦",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEmailMedia_Send_163(t *testing.T) {
	var media = senders.NewEmailMedia()
	media.SMTP = "smtp.163.com:465"
	media.Username = "iwind_php@163.com"
	media.Password = "123456" // 换成你的邮件密码或者授权码
	media.From = "iwind_php@163.com"
	_, err := media.Send("iwind_php@163.com", "This is test subject", "This is a test body <strong>粗体哦</strong><br/>换行哦")
	if err != nil {
		t.Fatal(err)
	}
}
