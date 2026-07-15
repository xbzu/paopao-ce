// Copyright 2026 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package admin

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rocboss/paopao-ce/internal/conf"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testUser struct {
	ID       int64 `gorm:"primaryKey"`
	Username string
	IsAdmin  bool
	IsDel    int8
}

func (testUser) TableName() string { return "p_user" }

func TestPromoteAdmin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err = db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if err = db.Create(&testUser{Username: "owner"}).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	promoted, err := promoteAdmin(db, "p_user", "owner")
	if err != nil {
		t.Fatalf("promoteAdmin() error = %v", err)
	}
	if promoted.ID == 0 || promoted.Username != "owner" {
		t.Fatalf("promoteAdmin() user = %+v", promoted)
	}
	var user testUser
	if err = db.First(&user, "username = ?", "owner").Error; err != nil {
		t.Fatalf("First() error = %v", err)
	}
	if !user.IsAdmin {
		t.Fatal("IsAdmin = false, want true")
	}
}

func TestPromoteAdminRejectsMissingUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err = db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if _, err = promoteAdmin(db, "p_user", "missing"); err == nil {
		t.Fatal("promoteAdmin() error = nil, want missing user error")
	}
}

func TestPromoteAdminIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err = db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if err = db.Create(&testUser{Username: "admin", IsAdmin: true}).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err = promoteAdmin(db, "p_user", "admin"); err != nil {
		t.Fatalf("promoteAdmin() repeated promotion error = %v", err)
	}
}

func TestPromoteAdminRejectsDeletedUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err = db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if err = db.Create(&testUser{Username: "deleted", IsDel: 1}).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err = promoteAdmin(db, "p_user", "deleted"); err == nil {
		t.Fatal("promoteAdmin() error = nil, want deleted user error")
	}
}

func TestInvalidatePromotedUserCacheUsesDatabaseIdentity(t *testing.T) {
	user := &promotedUser{ID: 42, Username: "ActualOwner"}
	var deleted []string
	err := invalidatePromotedUserCache(func(keys ...string) error {
		deleted = append(deleted, keys...)
		return nil
	}, user)
	if err != nil {
		t.Fatalf("invalidatePromotedUserCache() error = %v", err)
	}
	want := []string{
		conf.PrefixUserInfoById + "42",
		conf.PrefixUserInfoByName + user.Username,
	}
	if !reflect.DeepEqual(deleted, want) {
		t.Fatalf("deleted keys = %v, want %v", deleted, want)
	}
}

func TestInvalidatePromotedUserCacheReturnsDeleteError(t *testing.T) {
	wantErr := errors.New("redis unavailable")
	err := invalidatePromotedUserCache(func(...string) error { return wantErr }, &promotedUser{ID: 1, Username: "owner"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("invalidatePromotedUserCache() error = %v, want wrapped delete error", err)
	}
}
