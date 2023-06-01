//go:build plus

package senders

import (
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/files"
	"testing"
)

func TestScriptMedia_Send(t *testing.T) {
	script := `#!/usr/bin/env bash

echo  "subject:${MessageSubject}"
echo "body:${MessageBody}"
`

	tmp := files.NewFile(Tea.Root + "/web/tmp/media_test.sh")
	err := tmp.WriteString(script)
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Chmod(0777)
	defer func() {
		_ = tmp.Delete()
	}()

	media := NewScriptMedia()
	media.Path = tmp.Path()
	_, err = media.Send("zhangsan", "this is subject", "this is body")
	if err != nil {
		t.Fatal(err)
	}
}

func TestScriptMedia_Send2(t *testing.T) {
	media := NewScriptMedia()
	media.ScriptType = "code"
	media.Script = `#!/usr/bin/env bash

echo  "subject:${MessageSubject}"
echo "body:${MessageBody}"
`
	_, err := media.Send("zhangsan", "this is subject", "this is body")
	if err != nil {
		t.Fatal(err)
	}
}
