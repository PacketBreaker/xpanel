package service

import (
	"errors"
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/dbtest"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
	"gorm.io/gorm"
	"path/filepath"
	"strings"
	"testing"
)

func TestXpanelMigrationPropagatesOrphanFailure(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", d)
	dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
	db := database.GetDB()
	db.Create(&xray.ClientTraffic{Email: "later-orphan", InboundId: 999})
	if err := db.Exec(`CREATE TRIGGER fail_later_cleanup BEFORE DELETE ON client_traffics WHEN OLD.inbound_id <> 0 BEGIN SELECT RAISE(ABORT, 'injected orphan cleanup failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
	err := (&InboundService{}).MigrateDB()
	if err == nil || !strings.Contains(err.Error(), "injected orphan cleanup failure") {
		t.Fatalf("error not propagated: %v", err)
	}
}
func TestXpanelVisionRepairPropagatesLoadFailure(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", d)
	dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
	db := database.GetDB()
	if err := db.Callback().Query().Before("gorm:query").Register("test:vision-load", func(tx *gorm.DB) { tx.AddError(errors.New("injected vision load failure")) }); err != nil {
		t.Fatal(err)
	}
	err := (&InboundService{}).MigrationRestoreVisionFlow()
	if err == nil || !strings.Contains(err.Error(), "injected vision load failure") {
		t.Fatalf("error not propagated: %v", err)
	}
}
func TestXpanelVisionRepairPropagatesEffectiveFlowFailure(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", d)
	dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
	db := database.GetDB()
	if err := db.Exec("DROP TABLE client_inbounds").Error; err != nil {
		t.Fatal(err)
	}
	_, _, err := (&InboundService{}).restoreVisionFlowForEligibleInboundChecked(nil, `{"clients":[{"email":"repair@test","id":"test"}]}`, `{"network":"tcp","security":"reality"}`, model.VLESS)
	if err == nil || !strings.Contains(err.Error(), "client_inbounds") {
		t.Fatalf("error not propagated: %v", err)
	}
}
func TestXpanelVisionRepairPropagatesWriteFailure(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", d)
	dbtest.InitDB(t, filepath.Join(d, "x-ui.db"))
	db := database.GetDB()
	db.Create(&model.ClientRecord{Id: 1, Email: "repair@test", Enable: true})
	db.Create(&model.ClientInbound{ClientId: 1, InboundId: 1, FlowOverride: "xtls-rprx-vision"})
	ib := &model.Inbound{Tag: "repair-target", Port: 19991, Protocol: model.VLESS, Settings: `{"clients":[{"email":"repair@test","id":"test","flow":""}]}`, StreamSettings: `{"network":"tcp","security":"reality"}`}
	if err := db.Create(ib).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER fail_vision_write BEFORE UPDATE OF settings ON inbounds BEGIN SELECT RAISE(ABORT, 'injected vision write failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
	err := (&InboundService{}).MigrationRestoreVisionFlow()
	if err == nil || !strings.Contains(err.Error(), "injected vision write failure") {
		t.Fatalf("error not propagated: %v", err)
	}
}
