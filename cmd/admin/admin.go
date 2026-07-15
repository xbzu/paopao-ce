// Copyright 2026 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package admin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alimy/tryst/cfg"
	"github.com/rocboss/paopao-ce/cmd"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/dao/cache"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var promoteUsername string

func init() {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "manage administrator access",
	}
	promoteCmd := &cobra.Command{
		Use:   "promote",
		Short: "promote an existing user to administrator",
		Args:  cobra.NoArgs,
		RunE:  promoteRun,
	}
	promoteCmd.Flags().StringVar(&promoteUsername, "username", "", "existing username to promote")
	_ = promoteCmd.MarkFlagRequired("username")
	adminCmd.AddCommand(promoteCmd)
	cmd.Register(adminCmd)
}

func promoteRun(_ *cobra.Command, _ []string) error {
	username := strings.TrimSpace(promoteUsername)
	if username == "" {
		return fmt.Errorf("username must not be empty")
	}

	conf.Initial(nil, false)
	db := conf.MustGormDB()
	defer conf.CloseDB()

	tableName := conf.DatabaseSetting.TablePrefix + conf.TableUser
	user, err := promoteAdmin(db, tableName, username)
	if err != nil {
		return err
	}
	if cfg.If("Redis") {
		if err := invalidatePromotedUserCache(cache.NewAppCache().Delete, user); err != nil {
			return err
		}
	}
	fmt.Printf("user %q is now an administrator\n", user.Username)
	return nil
}

type promotedUser struct {
	ID       int64
	Username string
}

func promoteAdmin(db *gorm.DB, tableName, username string) (*promotedUser, error) {
	var user promotedUser
	if err := db.Table(tableName).
		Select("id", "username").
		Where("username = ? AND is_del = ?", username, 0).
		Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user %q was not found", username)
		}
		return nil, fmt.Errorf("find administrator candidate: %w", err)
	}
	result := db.Table(tableName).
		Where("id = ? AND is_del = ?", user.ID, 0).
		Update("is_admin", true)
	if result.Error != nil {
		return nil, fmt.Errorf("promote administrator: %w", result.Error)
	}
	var verified struct {
		IsAdmin bool
	}
	if err := db.Table(tableName).
		Select("is_admin").
		Where("id = ? AND is_del = ?", user.ID, 0).
		Take(&verified).Error; err != nil {
		return nil, fmt.Errorf("verify promoted administrator: %w", err)
	}
	if !verified.IsAdmin {
		return nil, fmt.Errorf("user %q was not promoted", user.Username)
	}
	return &user, nil
}

func invalidatePromotedUserCache(deleteKeys func(...string) error, user *promotedUser) error {
	if user == nil {
		return fmt.Errorf("promoted user is missing")
	}
	keys := []string{
		fmt.Sprintf("%s%d", conf.PrefixUserInfoById, user.ID),
		conf.PrefixUserInfoByName + user.Username,
	}
	if err := deleteKeys(keys...); err != nil {
		return fmt.Errorf("administrator was promoted but user cache invalidation failed: %w", err)
	}
	return nil
}
