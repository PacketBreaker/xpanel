package service

import (
	"fmt"
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/dbtest"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
	"gorm.io/gorm"
	"path/filepath"
	"testing"
	"time"
)

func TestXpanelOwnerRepairScale(t *testing.T) {
	if testing.Short() {
		t.Skip("scale")
	}
	for _, n := range []int{1000, 5000, 20000} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			d := t.TempDir()
			t.Setenv("XUI_DB_FOLDER", d)
			dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
			db := database.GetDB()
			clients := make([]model.ClientRecord, n)
			links := make([]model.ClientInbound, n)
			traffic := make([]xray.ClientTraffic, n)
			for i := 0; i < n; i++ {
				email := fmt.Sprintf("scale%d@test", i)
				clients[i] = model.ClientRecord{Id: i + 1, Email: email}
				links[i] = model.ClientInbound{ClientId: i + 1, InboundId: i + 1}
				traffic[i] = xray.ClientTraffic{Email: email, InboundId: i + 1, Up: 123, Down: 456}
			}
			if e := db.CreateInBatches(clients, 100).Error; e != nil {
				t.Fatal(e)
			}
			if e := db.CreateInBatches(links, 100).Error; e != nil {
				t.Fatal(e)
			}
			if e := db.CreateInBatches(traffic, 100).Error; e != nil {
				t.Fatal(e)
			}
			count := int64(0)
			db.Callback().Update().Before("gorm:update").Register("test:count-owner-updates", func(tx *gorm.DB) { count++ })
			start := time.Now()
			if e := repairTrafficDisplayOwners(db); e != nil {
				t.Fatal(e)
			}
			t.Logf("healthy rows=%d elapsed=%s update_statements=%d", n, time.Since(start), count)
			if count != 0 {
				t.Fatal("unchanged owners were written")
			}
			if e := db.Model(&xray.ClientTraffic{}).Where("id % 2 = 0").Update("inbound_id", 999999).Error; e != nil {
				t.Fatal(e)
			}
			start = time.Now()
			if e := repairTrafficDisplayOwners(db); e != nil {
				t.Fatal(e)
			}
			t.Logf("half-stale rows=%d elapsed=%s", n, time.Since(start))
			var wrong int64
			db.Model(&xray.ClientTraffic{}).Where("inbound_id = ?", 999999).Count(&wrong)
			if wrong != 0 {
				t.Fatal("stale rows remain")
			}
		})
	}
}
