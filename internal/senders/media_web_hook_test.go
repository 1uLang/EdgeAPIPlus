//go:build plus

package senders

import (
	"testing"
)

func TestMediaWebHook_Send(t *testing.T) {
	media := NewWebHookMedia()
	media.URL = "http://127.0.0.1:9991/webhook?subject=${MessageSubject}&body=${MessageBody}"
	resp, err := media.Send("zhangsan", "this is subject", "this is body")
	t.Log(string(resp), err)
}
