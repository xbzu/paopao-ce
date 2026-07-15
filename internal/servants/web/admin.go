// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package web

import (
	"context"
	"sort"
	"time"

	"github.com/alimy/tryst/cfg"
	"github.com/gin-gonic/gin"
	api "github.com/rocboss/paopao-ce/auto/api/v1"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/model/web"
	"github.com/rocboss/paopao-ce/internal/servants/base"
	"github.com/rocboss/paopao-ce/internal/servants/chain"
	"github.com/rocboss/paopao-ce/internal/sitesetting"
	"github.com/rocboss/paopao-ce/pkg/version"
	"github.com/rocboss/paopao-ce/pkg/xerror"
	"github.com/sirupsen/logrus"
)

type adminSrv struct {
	api.UnimplementedAdminServant
	*base.DaoServant
	wc           core.WebCache
	settings     *sitesetting.Service
	serverUpTime int64
}

func (s *adminSrv) Chain() gin.HandlersChain {
	return gin.HandlersChain{chain.JWT(), chain.Admin()}
}

func (s *adminSrv) ChangeUserStatus(req *web.ChangeUserStatusReq) error {
	user, err := s.Ds.GetUserByID(req.ID)
	if err != nil || user.Model == nil || user.ID <= 0 {
		return web.ErrNoExistUsername
	}
	// 执行更新
	user.Status = req.Status
	if err := s.Ds.UpdateUser(user); err != nil {
		return xerror.ServerError
	}
	return nil
}

func (s *adminSrv) SiteInfo(req *web.SiteInfoReq) (*web.SiteInfoResp, error) {
	res, err := &web.SiteInfoResp{
		ServerUpTime:     s.serverUpTime,
		DeploymentMode:   deploymentMode(),
		RunMode:          conf.RunMode(),
		Version:          runtimeVersion(),
		DatabaseProvider: databaseProvider(),
		CacheProvider:    cacheProvider(),
		SearchProvider:   searchProvider(),
		StorageProvider:  storageProvider(),
		EnabledFeatures:  enabledFeatures(),
	}, error(nil)
	res.RegisterUserCount, err = s.Ds.GetRegisterUserCount()
	if err != nil {
		logrus.Errorf("get SiteInfo[1] occurs error: %s", err)
	}
	onlineUserKeys, xerr := s.wc.Keys(conf.PrefixOnlineUser + "*")
	if xerr == nil {
		res.OnlineUserCount = len(onlineUserKeys)
		if res.HistoryMaxOnline, err = s.wc.PutHistoryMaxOnline(res.OnlineUserCount); err != nil {
			logrus.Errorf("get Siteinfo[3] occurs error: %s", err)
		}
	} else {
		logrus.Errorf("get Siteinfo[2] occurs error: %s", xerr)
	}
	if schema, xerr := s.settings.GetSchema(); xerr == nil {
		for _, item := range schema.Items {
			if item.Active {
				res.ActiveSettingsCount++
			}
		}
	} else {
		logrus.Errorf("get SiteInfo settings schema occurs error: %s", xerr)
	}
	if values, xerr := s.settings.GetValues(context.Background()); xerr == nil {
		res.HasPendingRestart = values.HasPendingRestart
	} else {
		logrus.Errorf("get SiteInfo settings values occurs error: %s", xerr)
	}
	// 错误进行宽松赦免处理
	return res, nil
}

func runtimeVersion() string {
	build := version.ReadBuildInfo()
	if build.Version == "" || build.Version == "unknown" {
		return build.Series
	}
	return build.Version
}

func deploymentMode() string {
	switch {
	case cfg.If("Sqlite3"):
		return "Simple"
	case cfg.If("MySQL"):
		return "Standard"
	default:
		return "Custom"
	}
}

func databaseProvider() string {
	switch {
	case cfg.If("Sqlite3"):
		return "SQLite"
	case cfg.Any("Postgres", "PostgreSQL"):
		return "PostgreSQL"
	case cfg.If("MySQL"):
		return "MySQL"
	default:
		return "Not configured"
	}
}

func cacheProvider() string {
	if cfg.If("Redis") {
		return "Redis"
	}
	return "Not configured"
}

func searchProvider() string {
	switch {
	case cfg.If("Meili"):
		return "Meilisearch"
	case cfg.If("Zinc"):
		return "Zinc"
	default:
		return "Not configured"
	}
}

func storageProvider() string {
	switch {
	case cfg.If("LocalOSS"):
		return "Local OSS"
	case cfg.If("MinIO"):
		return "MinIO"
	case cfg.If("S3"):
		return "Amazon S3"
	case cfg.If("COS"):
		return "Tencent COS"
	case cfg.If("HuaweiOBS"):
		return "Huawei OBS"
	case cfg.If("AliOSS"):
		return "AliOSS"
	default:
		return "Not configured"
	}
}

func enabledFeatures() []string {
	candidates := []string{
		"Web", "Frontend:EmbedWeb", "Migration", "MySQL", "Postgres", "PostgreSQL", "Sqlite3",
		"Redis", "Meili", "Zinc", "LocalOSS", "MinIO", "S3", "AliOSS", "COS", "HuaweiOBS",
		"Sms", "Alipay", "Metrics", "Pprof", "Sentry",
	}
	features := make([]string, 0, len(candidates))
	for _, feature := range candidates {
		if cfg.If(feature) {
			features = append(features, feature)
		}
	}
	sort.Strings(features)
	return features
}

func (s *adminSrv) GetSiteSettings() (*web.SiteSettingsResp, error) {
	profile, err := s.settings.GetProfile(context.Background())
	if err != nil {
		return nil, err
	}
	return &web.SiteSettingsResp{
		SiteProfileResp: *profile,
		ReadonlyFields:  sitesetting.CloneReadonlyFields(),
	}, nil
}

func (s *adminSrv) UpdateSiteSettings(req *web.SiteSettingsReq) (*web.SiteSettingsResp, error) {
	profile, err := s.settings.UpdateEditableProfile(context.Background(), sitesetting.EditableFromRequest(req))
	if err != nil {
		return nil, err
	}
	return &web.SiteSettingsResp{
		SiteProfileResp: *profile,
		ReadonlyFields:  sitesetting.CloneReadonlyFields(),
	}, nil
}

func (s *adminSrv) GetSettingsSchema() (*web.AdminSettingsSchemaResp, error) {
	return s.settings.GetSchema()
}

func (s *adminSrv) GetSettingsValues() (*web.AdminSettingsValuesResp, error) {
	return s.settings.GetValues(context.Background())
}

func (s *adminSrv) SaveSettings(req *web.AdminSettingsSaveReq) (*web.AdminSettingsSaveResp, error) {
	return s.settings.SaveValues(context.Background(), req.Items)
}

func newAdminSrv(s *base.DaoServant, wc core.WebCache, settings *sitesetting.Service) api.Admin {
	return &adminSrv{
		DaoServant:   s,
		wc:           wc,
		settings:     settings,
		serverUpTime: time.Now().Unix(),
	}
}
