package models

import (
	"github.com/1uLang/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestUserDAO_UpdateUserFeatures(t *testing.T) {
	var dao = NewUserDAO()
	var tx *dbs.Tx
	err := dao.UpdateUsersFeatures(tx, []string{
		userconfigs.UserFeatureCodeServerACME,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
}
