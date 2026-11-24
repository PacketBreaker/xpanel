package service

import (
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/dbtest"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
	"gorm.io/gorm"
	"path/filepath"
	"testing"
)

func TestXpanelDetachInboundPreservesGlobalTotalsWithoutStaleOwner(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", d)
	dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
	db := database.GetDB()
	cs := &ClientService{}
	for _, c := range []model.ClientRecord{{Id: 1, Email: "solo@x"}, {Id: 2, Email: "shared@x"}} {
		if e := db.Create(&c).Error; e != nil {
			t.Fatal(e)
		}
	}
	for _, l := range []model.ClientInbound{{ClientId: 1, InboundId: 1}, {ClientId: 2, InboundId: 1}, {ClientId: 2, InboundId: 2}} {
		db.Create(&l)
	}
	for _, email := range []string{"solo@x", "shared@x"} {
		db.Create(&xray.ClientTraffic{Email: email, InboundId: 1, Up: 123, Down: 456})
	}
	if e := db.Transaction(func(tx *gorm.DB) error {
		if e := cs.DetachInbound(tx, 1); e != nil {
			return e
		}
		return repairTrafficDisplayOwners(tx)
	}); e != nil {
		t.Fatal(e)
	}
	for email, want := range map[string]int{"solo@x": -1, "shared@x": 2} {
		var st xray.ClientTraffic
		if e := db.Where("email = ?", email).First(&st).Error; e != nil {
			t.Fatal(e)
		}
		if st.InboundId != want || st.Up != 123 || st.Down != 456 {
			t.Fatalf("wrong owner/totals: %+v", st)
		}
	}
}
func TestXpanelLegacyReusedOwnerHiddenFromUnrelatedInbound(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", d)
	dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
	db := database.GetDB()
	ib := model.Inbound{UserId: 1, Tag: "reused", Port: 19898, Protocol: model.VLESS, Settings: `{"clients":[]}`, StreamSettings: `{}`}
	db.Create(&ib)
	db.Create(&model.ClientRecord{Email: "old@x"})
	db.Create(&xray.ClientTraffic{Email: "old@x", InboundId: ib.Id, Up: 123, Down: 456})
	if e := (&InboundService{}).MigrationRequirements(); e != nil {
		t.Fatal(e)
	}
	rows, e := (&InboundService{}).GetInbounds(1)
	if e != nil {
		t.Fatal(e)
	}
	if len(rows) != 1 || len(rows[0].ClientStats) != 0 {
		t.Fatalf("stale traffic visible %+v", rows)
	}
	var count int64
	db.Model(&xray.ClientTraffic{}).Where("email = ?", "old@x").Count(&count)
	if count != 1 {
		t.Fatal("global accounting deleted")
	}
}
func TestXpanelDetachedLifecycleDoesNotAffectReusedInbound(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", d)
	dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
	db := database.GetDB()
	svc := &InboundService{}
	old := model.ClientRecord{Email: "expired-old@x", Enable: true}
	db.Create(&old)
	db.Create(&xray.ClientTraffic{Email: old.Email, InboundId: 1, Enable: true, Up: 123, Down: 456, Total: 100, ExpiryTime: 1, Reset: 0})
	ib := model.Inbound{Id: 1, UserId: 1, Tag: "new-owner", Port: 19897, Protocol: model.VLESS, Settings: `{"clients":[{"email":"new@x","id":"11111111-1111-1111-1111-111111111111","enable":true}]}`, StreamSettings: `{}`}
	db.Create(&ib)
	if e := (&ClientService{}).SyncInbound(nil, 1, []model.Client{{Email: "new@x", ID: "11111111-1111-1111-1111-111111111111", Enable: true}}); e != nil {
		t.Fatal(e)
	}
	db.Create(&xray.ClientTraffic{Email: "new@x", InboundId: 1, Enable: true, Up: 7, Down: 8})
	if e := repairTrafficDisplayOwners(db); e != nil {
		t.Fatal(e)
	}
	if _, _, e := svc.AddTraffic(nil, nil); e != nil {
		t.Fatal(e)
	}
	n := trafficOf(t, "new@x")
	if !n.Enable || n.Up != 7 || n.Down != 8 {
		t.Fatalf("new client affected: %+v", n)
	}
	if e := svc.ResetInboundTraffic(1); e != nil {
		t.Fatal(e)
	}
	o := trafficOf(t, old.Email)
	if o.Up != 123 || o.Down != 456 || o.Total != 100 || o.ExpiryTime != 1 || o.InboundId != -1 {
		t.Fatalf("old history/lifecycle changed by unrelated reset: %+v", o)
	}
	if e := svc.ResetClientTrafficByEmail("new@x"); e != nil {
		t.Fatal(e)
	}
	o = trafficOf(t, old.Email)
	if o.Up != 123 || o.Down != 456 {
		t.Fatalf("new-client reset changed old usage: %+v", o)
	}
}
