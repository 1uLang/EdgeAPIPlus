//go:build plus

package senders

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"time"
)

type MailInfo struct {
	To      string
	Subject string
	Body    string
}

// EmailMedia 邮件媒介
type EmailMedia struct {
	SMTP     string `yaml:"smtp" json:"smtp"`         // SMTP地址：host:port
	Username string `yaml:"username" json:"username"` // 用户名
	Password string `yaml:"password" json:"password"` // 密码
	From     string `yaml:"from" json:"from"`         // 发件人
	FromName string `yaml:"fromName" json:"fromName"` // 发件人名称
	Protocol string `yaml:"protocol" json:"protocol"` // 协议：tcp/tls
}

// NewEmailMedia 获取新对象
func NewEmailMedia() *EmailMedia {
	return &EmailMedia{}
}

func (this *EmailMedia) Send(user string, subject string, body string) (resp []byte, err error) {
	if len(this.SMTP) == 0 {
		return nil, errors.New("host address should be specified")
	}

	// 自动加端口

	if _, _, err := net.SplitHostPort(this.SMTP); err != nil {
		this.SMTP += ":587"
	}

	if len(this.From) == 0 {
		this.From = this.Username
	}

	var contentType = "Content-Type: text/html; charset=UTF-8"
	var senderName = this.FromName
	if len(senderName) == 0 {
		productName, err := models.SharedSysSettingDAO.ReadProductName(nil)
		if err != nil {
			return nil, errors.New("read product name failed: " + err.Error())
		}
		if len(productName) > 0 {
			senderName = productName
		} else {
			senderName = teaconst.ProductName
		}
	}
	var msg = []byte("To: " + user + "\r\nFrom: " + strconv.Quote(senderName) + " <" + this.From + ">\r\nSubject: " + "=?utf-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?=" + "\r\n" + contentType + "\r\n\r\n" + body)
	return nil, this.SendMail(this.From, []string{user}, msg)
}

// RequireUser 是否需要用户标识
func (this *EmailMedia) RequireUser() bool {
	return true
}

// SendMail 发送邮件
func (this *EmailMedia) SendMail(from string, to []string, message []byte) error {
	_, err := mail.ParseAddress(from)
	if err != nil {
		return err
	}

	if len(to) == 0 {
		return errors.New("recipients should not be empty")
	}

	for _, to1 := range to {
		_, err := mail.ParseAddress(to1)
		if err != nil {
			return err
		}
	}

	client, err := this.Connect()
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Quit()
		_ = client.Close()
	}()

	// To && From
	if err := client.Mail(from); err != nil {
		return err
	}

	for _, to1 := range to {
		if err := client.Rcpt(to1); err != nil {
			return err
		}
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(message)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

// SendMails 发送邮件
func (this *EmailMedia) SendMails(mails []*MailInfo) error {
	if len(mails) == 0 {
		return nil
	}

	_, err := mail.ParseAddress(this.From)
	if err != nil {
		return err
	}

	client, err := this.Connect()
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Quit()
		_ = client.Close()
	}()

	// 发送者
	var senderName = this.FromName
	if len(senderName) == 0 {
		productName, err := models.SharedSysSettingDAO.ReadProductName(nil)
		if err != nil {
			return errors.New("read product name failed: " + err.Error())
		}
		if len(productName) > 0 {
			senderName = productName
		} else {
			senderName = teaconst.ProductName
		}
	}

	// 创建邮件
	for _, mailInfo := range mails {
		if len(mailInfo.To) == 0 {
			continue
		}

		_, err = mail.ParseAddress(mailInfo.To)
		if err != nil {
			return err
		}

		if err := client.Mail(this.From); err != nil {
			return err
		}

		if err := client.Rcpt(mailInfo.To); err != nil {
			return err
		}

		// Data
		w, err := client.Data()
		if err != nil {
			return err
		}

		_, err = w.Write([]byte("To: " + mailInfo.To + "\r\nFrom: " + strconv.Quote(senderName) + " <" + this.From + ">\r\nSubject: " + mailInfo.Subject + "\r\n" + "Content-Type: text/html; charset=UTF-8" + "\r\n\r\n" + mailInfo.Body))
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *EmailMedia) Connect() (*smtp.Client, error) {
	var serverName = this.SMTP
	var username = this.Username
	var password = this.Password

	host, port, _ := net.SplitHostPort(serverName)

	var client *smtp.Client

	// TLS config
	var tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	if this.Protocol == "tcp" || (len(this.Protocol) == 0 && port == "25") { // 25 port: default tcp port
		conn, err := net.DialTimeout("tcp", serverName, 10*time.Second)
		if err != nil {
			return nil, err
		}
		client, err = smtp.NewClient(conn, host)
		if err != nil {
			return nil, err
		}
	} else if port == "587" { // 587 port: prefer START_TLS
		conn, err := net.DialTimeout("tcp", serverName, 10*time.Second)
		if err != nil {
			conn, err := tls.Dial("tcp", serverName, tlsConfig)
			if err != nil {
				return nil, err
			}
			client, err = smtp.NewClient(conn, host)
			if err != nil {
				return nil, err
			}
		} else {
			client, err = smtp.NewClient(conn, host)
			if err != nil {
				return nil, err
			}
			err = client.StartTLS(tlsConfig)
			if err != nil {
				return nil, err
			}
		}
	} else {
		conn, err := tls.Dial("tcp", serverName, tlsConfig)
		if err != nil {
			conn, err := net.DialTimeout("tcp", serverName, 10*time.Second)
			if err != nil {
				return nil, err
			}

			client, err = smtp.NewClient(conn, host)
			if err != nil {
				return nil, err
			}
			err = client.StartTLS(tlsConfig)
			if err != nil {
				return nil, err
			}
		} else {
			client, err = smtp.NewClient(conn, host)
			if err != nil {
				return nil, err
			}
		}
	}

	// 认证
	var auth = smtp.PlainAuth("", username, password, host)
	if err := client.Auth(auth); err != nil {
		_ = client.Quit()
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
