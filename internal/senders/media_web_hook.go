//go:build plus

package senders

import (
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/utils/string"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebHookMedia WebHook媒介
type WebHookMedia struct {
	URL         string      `yaml:"url" json:"url"` // URL中可以使用${MessageSubject}, ${MessageBody}两个变量
	Method      string      `yaml:"method" json:"method"`
	ContentType string      `yaml:"contentType" json:"contentType"` // 内容类型：params|body
	Headers     []*Variable `yaml:"headers" json:"headers"`
	Params      []*Variable `yaml:"params" json:"params"`
	Body        string      `yaml:"body" json:"body"`
}

// NewWebHookMedia 获取新对象
func NewWebHookMedia() *WebHookMedia {
	return &WebHookMedia{}
}

// Send 发送
func (this *WebHookMedia) Send(user string, subject string, body string) (resp []byte, err error) {
	if len(this.URL) == 0 {
		return nil, errors.New("'url' should be specified")
	}

	timeout := 10 * time.Second

	if len(this.Method) == 0 {
		this.Method = http.MethodGet
	}

	this.URL = strings.Replace(this.URL, "${MessageUser}", url.QueryEscape(user), -1)
	this.URL = strings.Replace(this.URL, "${MessageSubject}", url.QueryEscape(subject), -1)
	this.URL = strings.Replace(this.URL, "${MessageBody}", url.QueryEscape(body), -1)

	var req *http.Request
	if this.Method == http.MethodGet {
		req, err = http.NewRequest(this.Method, this.URL, nil)
	} else {
		params := url.Values{
			"MessageUser":    []string{user},
			"MessageSubject": []string{subject},
			"MessageBody":    []string{body},
		}

		postBody := ""
		if this.ContentType == "params" {
			for _, param := range this.Params {
				param.Value = strings.Replace(param.Value, "${MessageUser}", user, -1)
				param.Value = strings.Replace(param.Value, "${MessageSubject}", subject, -1)
				param.Value = strings.Replace(param.Value, "$MessageBody}", body, -1)
				params.Add(param.Name, param.Value)
			}
			postBody = params.Encode()
		} else if this.ContentType == "body" {
			userJSON := stringutil.JSONEncode(user)
			subjectJSON := stringutil.JSONEncode(subject)
			bodyJSON := stringutil.JSONEncode(body)
			if len(userJSON) > 0 {
				userJSON = userJSON[1 : len(userJSON)-1]
			}
			if len(subjectJSON) > 0 {
				subjectJSON = subjectJSON[1 : len(subjectJSON)-1]
			}
			if len(bodyJSON) > 0 {
				bodyJSON = bodyJSON[1 : len(bodyJSON)-1]
			}
			postBody = strings.Replace(this.Body, "${MessageUser}", userJSON, -1)
			postBody = strings.Replace(postBody, "${MessageSubject}", subjectJSON, -1)
			postBody = strings.Replace(postBody, "${MessageBody}", bodyJSON, -1)
		} else {
			postBody = params.Encode()
		}

		req, err = http.NewRequest(this.Method, this.URL, strings.NewReader(postBody))
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)

	if len(this.Headers) > 0 {
		for _, h := range this.Headers {
			req.Header.Set(h.Name, h.Value)
		}
	}

	client := utils.SharedHttpClient(timeout)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	data, err := io.ReadAll(response.Body)
	return data, err
}

// RequireUser 是否需要用户标识
func (this *WebHookMedia) RequireUser() bool {
	return false
}

// AddHeader 添加Header
func (this *WebHookMedia) AddHeader(name string, value string) {
	this.Headers = append(this.Headers, &Variable{
		Name:  name,
		Value: value,
	})
}

// 添加参数
func (this *WebHookMedia) AddParam(name string, value string) {
	this.Params = append(this.Params, &Variable{
		Name:  name,
		Value: value,
	})
}
