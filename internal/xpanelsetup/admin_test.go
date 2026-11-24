package xpanelsetup

import (
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/util/crypto"
	"path/filepath"
	"testing"
)

func TestFirstAdminValidationAndNoReplacement(t *testing.T) {
	if err := database.InitDB(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatal(err)
	}
	defer database.CloseDB()
	if err := CreateFirstAdmin("local", []byte("short"), []byte("short")); err == nil {
		t.Fatal("short accepted")
	}
	p := []byte("synthetic-test-password")
	if err := CreateFirstAdmin("local", p, []byte("mismatch-password")); err == nil {
		t.Fatal("mismatch accepted")
	}
	if err := CreateFirstAdmin("local", p, p); err != nil {
		t.Fatal(err)
	}
	var u model.User
	if err := database.GetDB().First(&u).Error; err != nil {
		t.Fatal(err)
	}
	if u.Password == string(p) || !crypto.CheckPasswordHash(u.Password, string(p)) {
		t.Fatal("not bcrypt")
	}
	if err := CreateFirstAdmin("replacement", p, p); err == nil {
		t.Fatal("replacement accepted")
	}
	if err := EnsureAdmin(); err != nil {
		t.Fatal("existing admin must not prompt", err)
	}
}

func TestRecoveryPreservesIdentityAndInvalidatesSessions(t *testing.T) {
	if err := database.InitDB(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatal(err)
	}
	defer database.CloseDB()
	p := []byte("synthetic-old-password")
	if err := CreateFirstAdmin("old", p, p); err != nil {
		t.Fatal(err)
	}
	var before model.User
	database.GetDB().First(&before)
	next := []byte("synthetic-new-password")
	if err := ResetFirstAdmin("new", next, next); err != nil {
		t.Fatal(err)
	}
	var after model.User
	database.GetDB().First(&after)
	if before.Id != after.Id || after.Username != "new" || !crypto.CheckPasswordHash(after.Password, string(next)) {
		t.Fatal("recovery did not preserve identity/update hash")
	}
	if after.LoginEpoch != before.LoginEpoch+1 {
		t.Fatal("sessions not invalidated")
	}
}
