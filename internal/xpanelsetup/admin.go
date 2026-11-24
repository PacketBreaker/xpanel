// Package xpanelsetup handles local first-admin setup without logging credentials.
package xpanelsetup

import (
	"bufio"
	"fmt"
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/util/crypto"
	"golang.org/x/term"
	"gorm.io/gorm"
	"os"
	"strings"
)

func EnsureAdmin() error {
	var count int64
	if err := database.GetDB().Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("Xpanel first-admin setup needs a local interactive terminal; no credential generated or logged")
	}
	defer tty.Close()
	if !term.IsTerminal(int(tty.Fd())) {
		return fmt.Errorf("Xpanel first-admin setup requires a terminal")
	}
	username, password, confirmation, err := promptCredentials(tty)
	if err != nil {
		return err
	}
	defer clear(password)
	defer clear(confirmation)
	if err = CreateFirstAdmin(username, password, confirmation); err != nil {
		return err
	}
	fmt.Fprintln(tty, "Admin created. Password was not logged. Open the local panel to sign in.")
	return nil
}
func CreateFirstAdmin(username string, password, confirmation []byte) error {
	username = strings.TrimSpace(username)
	if username == "" || len(username) > 64 {
		return fmt.Errorf("username must be1..64characters")
	}
	if len(password) < 12 || len(password) > 72 {
		return fmt.Errorf("password must be12..72bytes")
	}
	if string(password) != string(confirmation) {
		return fmt.Errorf("password confirmation does not match")
	}
	hash, err := crypto.HashPasswordAsBcrypt(string(password))
	if err != nil {
		return err
	}
	return database.GetDB().Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.User{}).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("an administrator already exists; refusing replacement")
		}
		return tx.Create(&model.User{Username: username, Password: hash}).Error
	})
}

func readHidden(tty *os.File, prompt string) ([]byte, error) {
	fd := int(tty.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, state)
	fmt.Fprint(tty, prompt)
	value, err := term.ReadPassword(fd)
	fmt.Fprint(tty, "\r\n")
	return value, err
}

func promptCredentials(tty *os.File) (string, []byte, []byte, error) {
	fmt.Fprint(tty, "Xpanel local admin setup. Username: ")
	username, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil {
		return "", nil, nil, err
	}
	username = strings.TrimSpace(username)
	if username == "" || len(username) > 64 {
		return "", nil, nil, fmt.Errorf("username must be1..64characters")
	}
	password, err := readHidden(tty, "Password (12..72bytes, hidden): ")
	if err != nil {
		return "", nil, nil, err
	}
	confirmation, err := readHidden(tty, "Confirm password (hidden): ")
	if err != nil {
		clear(password)
		return "", nil, nil, err
	}
	return username, password, confirmation, nil
}

// ResetAdmin is a local-console recovery action, never an HTTP endpoint.
func ResetAdmin() error {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("admin recovery requires a local interactive terminal")
	}
	defer tty.Close()
	if !term.IsTerminal(int(tty.Fd())) {
		return fmt.Errorf("admin recovery requires a terminal")
	}
	username, password, confirmation, err := promptCredentials(tty)
	if err != nil {
		return err
	}
	defer clear(password)
	defer clear(confirmation)
	if err = ResetFirstAdmin(username, password, confirmation); err != nil {
		return err
	}
	fmt.Fprintln(tty, "Local admin credentials changed; prior sessions invalidated. Password was not logged.")
	return nil
}
func ResetFirstAdmin(username string, password, confirmation []byte) error {
	username = strings.TrimSpace(username)
	if username == "" || len(username) > 64 {
		return fmt.Errorf("username must be1..64characters")
	}
	if len(password) < 12 || len(password) > 72 || string(password) != string(confirmation) {
		return fmt.Errorf("password must be12..72bytes and confirmation must match")
	}
	hash, err := crypto.HashPasswordAsBcrypt(string(password))
	if err != nil {
		return err
	}
	return database.GetDB().Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.User{}).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("recovery supports exactly one existing admin; refusing ambiguous or empty database")
		}
		var user model.User
		if err := tx.First(&user).Error; err != nil {
			return err
		}
		return tx.Model(&user).Updates(map[string]any{"username": username, "password": hash, "login_epoch": gorm.Expr("login_epoch + 1")}).Error
	})
}
