package service

import (
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/dbtest"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
	"path/filepath"
	"testing"
)

func TestXpanelMigrationEmptySelectionCommitsLaterCleanup(t *testing.T) {
	for _, nonmatching := range []bool{false, true} {
		t.Run(map[bool]string{false: "empty", true: "nonmatching"}[nonmatching], func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XUI_DB_FOLDER", dir)
			dbtest.InitDB(t, filepath.Join(dir, "x-ui.db"))
			db := database.GetDB()
			if nonmatching {
				if err := db.Create(&model.Inbound{UserId: 1, Tag: "0.0.0.0:19899-tunnel", Port: 19899, Protocol: model.Tunnel, Settings: `{}`, StreamSettings: `{}`}).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Create(&xray.ClientTraffic{Email: "orphan-no-inbound", InboundId: 0}).Error; err != nil {
				t.Fatal(err)
			}
			if err := (&InboundService{}).MigrationRequirements(); err != nil {
				t.Fatal(err)
			}
			var n int64
			db.Model(&xray.ClientTraffic{}).Where("inbound_id = 0").Count(&n)
			if n != 0 {
				t.Fatal("later orphan cleanup did not commit")
			}
			if nonmatching {
				var in model.Inbound
				db.First(&in)
				if in.Tag != "19899-tunnel" {
					t.Fatal("later tag cleanup did not commit", in.Tag)
				}
			}
		})
	}
}
