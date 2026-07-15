package sitesetting

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/alimy/tryst/cfg"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/model/web"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var errUnexpectedDryRunQuery = errors.New("unexpected database call during dry run")

type dryRunConnPool struct{}

func (dryRunConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, errUnexpectedDryRunQuery
}

func (dryRunConnPool) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, errUnexpectedDryRunQuery
}

func (dryRunConnPool) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, errUnexpectedDryRunQuery
}

func (dryRunConnPool) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return &sql.Row{}
}

func TestDeleteOverrideQuotesReservedKeyForMySQL(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      dryRunConnPool{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing:   true,
		DryRun:                 true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open dry-run MySQL database: %v", err)
	}

	result := deleteOverrideByKey(db, "web_profile.allow_user_register")
	if result.Error != nil {
		t.Fatalf("build dry-run delete: %v", result.Error)
	}
	stmt := result.Statement
	if got := stmt.SQL.String(); !strings.Contains(got, ".`key` = ?") {
		t.Fatalf("delete SQL = %q, want quoted key column", got)
	}
}

func TestGetProfileUsesBootstrapDefaultsWhenNoOverride(t *testing.T) {
	svc := newTestService(t)

	profile, err := svc.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if !profile.AllowUserRegister {
		t.Fatalf("AllowUserRegister = false, want bootstrap true")
	}
	if profile.DefaultTweetVisibility != "friend" {
		t.Fatalf("DefaultTweetVisibility = %q, want friend", profile.DefaultTweetVisibility)
	}
	if profile.CopyrightRight != "fallback-right" {
		t.Fatalf("CopyrightRight = %q, want fallback-right", profile.CopyrightRight)
	}
}

func TestUpdateEditableProfilePersistsOnlyEditableKeys(t *testing.T) {
	svc := newTestService(t)

	profile, err := svc.UpdateEditableProfile(context.Background(), EditableProfile{
		UseFriendship:             false,
		EnableTrendsBar:           true,
		EnableWallet:              true,
		AllowTweetAttachment:      false,
		AllowTweetAttachmentPrice: false,
		AllowTweetVideo:           false,
		DefaultTweetMaxLength:     1200,
		TweetWebEllipsisSize:      300,
		TweetMobileEllipsisSize:   200,
		DefaultTweetVisibility:    "public",
		DefaultMsgLoopInterval:    3000,
		CopyrightTop:              "top",
		CopyrightLeft:             "left",
		CopyrightLeftLink:         "https://left.example.com",
		CopyrightRight:            "right",
		CopyrightRightLink:        "https://right.example.com",
	})
	if err != nil {
		t.Fatalf("UpdateEditableProfile() error = %v", err)
	}
	if !profile.AllowUserRegister {
		t.Fatalf("AllowUserRegister = false, want bootstrap true")
	}
	if !profile.AllowPhoneBind {
		t.Fatalf("AllowPhoneBind = false, want bootstrap true")
	}
	if profile.DefaultTweetVisibility != "friend" {
		t.Fatalf("DefaultTweetVisibility = %q, want current effective friend", profile.DefaultTweetVisibility)
	}
	values, err := svc.GetValues(context.Background())
	if err != nil {
		t.Fatalf("GetValues() error = %v", err)
	}
	if !hasValue(values.Items, "web_profile.enable_wallet", true, true) {
		t.Fatalf("web_profile.enable_wallet override not found")
	}
	if !values.HasPendingRestart {
		t.Fatal("HasPendingRestart = false, want true")
	}
}

func TestSaveValuesEncryptsSecretsAtRest(t *testing.T) {
	svc := newTestService(t)
	conf.AdminSettingsSetting.EncryptionKey = "bootstrap-test-encryption-key"
	svc.codec = newSecretCodec()

	_, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "meili.api_key",
		Value: []byte(`"top-secret-key"`),
	}})
	if err != nil {
		t.Fatalf("SaveValues() error = %v", err)
	}
	var record settingRecord
	if err := svc.db.WithContext(context.Background()).First(&record, "key = ?", "meili.api_key").Error; err != nil {
		t.Fatalf("load record error = %v", err)
	}
	if !record.IsEncrypted {
		t.Fatalf("IsEncrypted = false, want true")
	}
	if strings.Contains(record.Value, "top-secret-key") {
		t.Fatalf("record.Value stored plaintext = %q", record.Value)
	}
}

func TestEnvironmentManagedMeiliKeyIsReadonly(t *testing.T) {
	t.Setenv("PAOPAO_MEILI_MASTER_KEY", "environment-managed-meili-key")
	svc := newTestService(t)

	stale := settingRecord{Key: "meili.api_key", Value: `"stale-database-key"`}
	if err := svc.db.Create(&stale).Error; err != nil {
		t.Fatalf("create stale meili override: %v", err)
	}
	Bootstrap(context.Background(), svc.db)
	if conf.MeiliSetting.ApiKey != "environment-managed-meili-key" {
		t.Fatalf("Meili ApiKey = %q, want environment value", conf.MeiliSetting.ApiKey)
	}

	schema, err := svc.GetSchema()
	if err != nil {
		t.Fatalf("GetSchema() error = %v", err)
	}
	found := false
	for _, item := range schema.Items {
		if item.Key != "meili.api_key" {
			continue
		}
		found = true
		if !item.Readonly || item.ApplyMode != string(ApplyModeBootstrapOnly) {
			t.Fatalf("meili.api_key schema = readonly:%t apply_mode:%s", item.Readonly, item.ApplyMode)
		}
	}
	if !found {
		t.Fatal("schema missing meili.api_key")
	}
	if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "meili.api_key",
		Value: []byte(`"database-key"`),
	}}); err == nil {
		t.Fatal("SaveValues() error = nil, want environment-managed key rejection")
	}
	values, err := svc.GetValues(context.Background())
	if err != nil {
		t.Fatalf("GetValues() error = %v", err)
	}
	for _, item := range values.Items {
		if item.Key == "meili.api_key" && item.Source != "bootstrap" {
			t.Fatalf("meili.api_key source = %q, want bootstrap", item.Source)
		}
	}
}

func TestRestartRequiredValuesReportPendingRestart(t *testing.T) {
	svc := newTestService(t)

	resp, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "meili.host",
		Value: []byte(`"pending-restart:7700"`),
	}})
	if err != nil {
		t.Fatalf("SaveValues() error = %v", err)
	}
	if !resp.HasPendingRestart {
		t.Fatal("HasPendingRestart = false, want true")
	}
	if !hasValue(resp.Items, "meili.host", "pending-restart:7700", true) {
		t.Fatalf("meili.host pending_restart not reported")
	}
}

func TestSaveValuesRejectsInvalidSingleIntUpdate(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "web_profile.default_tweet_max_length",
		Value: []byte(`120`),
	}})
	if err == nil {
		t.Fatal("SaveValues() error = nil, want invalid params error")
	}
}

func TestSaveValuesAcceptsCoupledLimitExpansion(t *testing.T) {
	svc := newTestService(t)
	conf.WebProfileSetting.DefaultTweetMaxLength = 100
	conf.WebProfileSetting.TweetWebEllipsisSize = 80
	conf.WebProfileSetting.TweetMobileEllipsisSize = 70
	conf.AppSetting.DefaultPageSize = 20
	conf.AppSetting.MaxPageSize = 100

	resp, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{
		{Key: "web_profile.default_tweet_max_length", Value: []byte(`300`)},
		{Key: "web_profile.tweet_web_ellipsis_size", Value: []byte(`240`)},
		{Key: "web_profile.tweet_mobile_ellipsis_size", Value: []byte(`220`)},
		{Key: "app.default_page_size", Value: []byte(`200`)},
		{Key: "app.max_page_size", Value: []byte(`300`)},
	})
	if err != nil {
		t.Fatalf("SaveValues() error = %v", err)
	}
	if !hasValue(resp.Items, "web_profile.default_tweet_max_length", 300, true) {
		t.Fatalf("web_profile.default_tweet_max_length override not found")
	}
	if !hasValue(resp.Items, "app.default_page_size", 200, true) {
		t.Fatalf("app.default_page_size override not found")
	}
	if !resp.HasPendingRestart {
		t.Fatal("HasPendingRestart = false, want true")
	}
	if conf.WebProfileSetting.DefaultTweetMaxLength != 100 || conf.WebProfileSetting.TweetWebEllipsisSize != 80 || conf.WebProfileSetting.TweetMobileEllipsisSize != 70 {
		t.Fatalf("effective tweet limits changed before restart = max:%d web:%d mobile:%d", conf.WebProfileSetting.DefaultTweetMaxLength, conf.WebProfileSetting.TweetWebEllipsisSize, conf.WebProfileSetting.TweetMobileEllipsisSize)
	}
	if conf.AppSetting.DefaultPageSize != 20 || conf.AppSetting.MaxPageSize != 100 {
		t.Fatalf("effective page limits changed before restart = default:%d max:%d", conf.AppSetting.DefaultPageSize, conf.AppSetting.MaxPageSize)
	}
}

func TestSaveValuesAcceptsCoupledLimitContraction(t *testing.T) {
	svc := newTestService(t)
	conf.WebProfileSetting.DefaultTweetMaxLength = 500
	conf.WebProfileSetting.TweetWebEllipsisSize = 400
	conf.WebProfileSetting.TweetMobileEllipsisSize = 300
	conf.AppSetting.DefaultPageSize = 100
	conf.AppSetting.MaxPageSize = 200

	resp, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{
		{Key: "web_profile.default_tweet_max_length", Value: []byte(`120`)},
		{Key: "web_profile.tweet_web_ellipsis_size", Value: []byte(`100`)},
		{Key: "web_profile.tweet_mobile_ellipsis_size", Value: []byte(`90`)},
		{Key: "app.default_page_size", Value: []byte(`20`)},
		{Key: "app.max_page_size", Value: []byte(`30`)},
	})
	if err != nil {
		t.Fatalf("SaveValues() error = %v", err)
	}
	if !hasValue(resp.Items, "web_profile.default_tweet_max_length", 120, true) {
		t.Fatalf("web_profile.default_tweet_max_length override not found")
	}
	if !hasValue(resp.Items, "app.max_page_size", 30, true) {
		t.Fatalf("app.max_page_size override not found")
	}
	if !resp.HasPendingRestart {
		t.Fatal("HasPendingRestart = false, want true")
	}
	if conf.WebProfileSetting.DefaultTweetMaxLength != 500 || conf.WebProfileSetting.TweetWebEllipsisSize != 400 || conf.WebProfileSetting.TweetMobileEllipsisSize != 300 {
		t.Fatalf("effective tweet limits changed before restart = max:%d web:%d mobile:%d", conf.WebProfileSetting.DefaultTweetMaxLength, conf.WebProfileSetting.TweetWebEllipsisSize, conf.WebProfileSetting.TweetMobileEllipsisSize)
	}
	if conf.AppSetting.DefaultPageSize != 100 || conf.AppSetting.MaxPageSize != 200 {
		t.Fatalf("effective page limits changed before restart = default:%d max:%d", conf.AppSetting.DefaultPageSize, conf.AppSetting.MaxPageSize)
	}
}

func TestSaveValuesRejectsInvalidSequentialPendingLimits(t *testing.T) {
	t.Run("pagination", func(t *testing.T) {
		svc := newTestService(t)
		if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
			Key:   "app.default_page_size",
			Value: []byte(`80`),
		}}); err != nil {
			t.Fatalf("save pending default page size: %v", err)
		}
		if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
			Key:   "app.max_page_size",
			Value: []byte(`50`),
		}}); err == nil {
			t.Fatal("second SaveValues() error = nil, want persisted default/max relationship rejection")
		}
	})

	t.Run("tweet ellipsis", func(t *testing.T) {
		svc := newTestService(t)
		if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
			Key:   "web_profile.default_tweet_max_length",
			Value: []byte(`500`),
		}}); err != nil {
			t.Fatalf("save pending tweet max: %v", err)
		}
		if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
			Key:   "web_profile.tweet_web_ellipsis_size",
			Value: []byte(`600`),
		}}); err == nil {
			t.Fatal("second SaveValues() error = nil, want persisted tweet relationship rejection")
		}
	})
}

func TestApplyPersistedOverridesRejectsInvalidCoupledStateAtomically(t *testing.T) {
	svc := newTestService(t)
	records := []settingRecord{
		{Key: "app.default_page_size", Value: `80`},
		{Key: "app.max_page_size", Value: `50`},
	}
	if err := svc.db.Create(&records).Error; err != nil {
		t.Fatalf("create invalid persisted settings: %v", err)
	}
	originalDefault := conf.AppSetting.DefaultPageSize
	originalMax := conf.AppSetting.MaxPageSize
	if err := svc.ApplyPersistedOverrides(context.Background()); err == nil {
		t.Fatal("ApplyPersistedOverrides() error = nil, want coupled-state rejection")
	}
	if conf.AppSetting.DefaultPageSize != originalDefault || conf.AppSetting.MaxPageSize != originalMax {
		t.Fatalf("invalid persisted state partially applied = default:%d max:%d", conf.AppSetting.DefaultPageSize, conf.AppSetting.MaxPageSize)
	}
}

func TestBootstrapAppliesPersistedOverrides(t *testing.T) {
	svc := newTestService(t)
	conf.AdminSettingsSetting.EncryptionKey = "bootstrap-test-encryption-key"
	svc.codec = newSecretCodec()

	_, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "web_profile.enable_wallet",
		Value: []byte(`true`),
	}})
	if err != nil {
		t.Fatalf("SaveValues() error = %v", err)
	}
	conf.WebProfileSetting.EnableWallet = false
	Bootstrap(context.Background(), svc.db)
	if !conf.WebProfileSetting.EnableWallet {
		t.Fatal("Bootstrap() did not apply persisted web_profile.enable_wallet override")
	}
}

func TestSaveValuesPublishesRuntimePolicyImmediately(t *testing.T) {
	svc := newTestService(t)

	resp, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{
		{Key: "web_profile.allow_user_register", Value: []byte(`false`)},
		{Key: "web_profile.allow_phone_bind", Value: []byte(`false`)},
	})
	if err != nil {
		t.Fatalf("SaveValues() error = %v", err)
	}
	policy := CurrentRuntimePolicy()
	if policy.AllowUserRegister {
		t.Fatal("AllowUserRegister = true, want false")
	}
	if policy.AllowPhoneBind {
		t.Fatal("AllowPhoneBind = true, want false")
	}
	if !hasValue(resp.Items, "web_profile.allow_user_register", false, false) {
		t.Fatal("web_profile.allow_user_register live value not reported")
	}
	if !hasValue(resp.Items, "web_profile.allow_phone_bind", false, false) {
		t.Fatal("web_profile.allow_phone_bind live value not reported")
	}
	profile, err := svc.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if profile.AllowUserRegister || profile.AllowPhoneBind {
		t.Fatalf("GetProfile() policy = register:%t phone:%t, want both false", profile.AllowUserRegister, profile.AllowPhoneBind)
	}
}

func TestRegistrationAndPhoneSettingsAreLiveAndEditable(t *testing.T) {
	svc := newTestService(t)
	schema, err := svc.GetSchema()
	if err != nil {
		t.Fatalf("GetSchema() error = %v", err)
	}
	for _, key := range []string{"web_profile.allow_user_register", "web_profile.allow_phone_bind"} {
		found := false
		for _, item := range schema.Items {
			if item.Key != key {
				continue
			}
			found = true
			if item.ApplyMode != string(ApplyModeLive) || item.Readonly {
				t.Fatalf("schema %s = apply_mode:%s readonly:%t, want live/editable", key, item.ApplyMode, item.Readonly)
			}
		}
		if !found {
			t.Fatalf("schema missing %s", key)
		}
	}
	if fields := CloneReadonlyFields(); len(fields) != 0 {
		t.Fatalf("ReadonlyFields = %v, want empty", fields)
	}
	for _, item := range schema.Items {
		if item.ApplyMode != string(ApplyModeLive) {
			continue
		}
		if item.Key != "web_profile.allow_user_register" && item.Key != "web_profile.allow_phone_bind" {
			t.Fatalf("unsafe live setting remains: %s", item.Key)
		}
	}
}

func TestPhoneBindingIsInactiveAndRejectedWithoutSMS(t *testing.T) {
	svc := newTestService(t)
	cfg.Initial(map[string][]string{
		"default": {"Web", "Meili", "LocalOSS"},
	}, nil)
	conf.WebProfileSetting.AllowPhoneBind = true
	bootstrapConfig = nil
	resetRuntimePolicyForTest()
	svc = NewService(svc.db)

	if CurrentRuntimePolicy().AllowPhoneBind {
		t.Fatal("AllowPhoneBind = true without Sms feature")
	}
	profile, err := svc.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if profile.AllowPhoneBind {
		t.Fatal("profile exposes phone binding without Sms feature")
	}
	schema, err := svc.GetSchema()
	if err != nil {
		t.Fatalf("GetSchema() error = %v", err)
	}
	for _, item := range schema.Items {
		if item.Key == "web_profile.allow_phone_bind" && item.Active {
			t.Fatal("phone binding setting is active without Sms feature")
		}
	}
	if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "web_profile.allow_phone_bind",
		Value: []byte(`true`),
	}}); err == nil {
		t.Fatal("SaveValues() accepted phone binding without Sms feature")
	}
}

func TestManagedPublicAddressIsEnvironmentOwned(t *testing.T) {
	t.Setenv("PAOPAO_PUBLIC_HOST", "community.example.test")
	t.Setenv("PAOPAO_PUBLIC_SECURE", "true")
	svc := newTestService(t)
	for _, record := range []settingRecord{
		{Key: "local_oss.domain", Value: "stale.example.test"},
		{Key: "local_oss.secure", Value: "false"},
	} {
		if err := svc.db.Create(&record).Error; err != nil {
			t.Fatalf("create stale %s override: %v", record.Key, err)
		}
	}
	Bootstrap(context.Background(), svc.db)
	if conf.LocalOSSSetting.Domain != "community.example.test" || !conf.LocalOSSSetting.Secure {
		t.Fatalf("environment-owned public address was overridden = domain:%q secure:%t", conf.LocalOSSSetting.Domain, conf.LocalOSSSetting.Secure)
	}

	schema, err := svc.GetSchema()
	if err != nil {
		t.Fatalf("GetSchema() error = %v", err)
	}
	for _, key := range []string{"local_oss.domain", "local_oss.secure"} {
		found := false
		for _, item := range schema.Items {
			if item.Key != key {
				continue
			}
			found = true
			if !item.Readonly || item.ApplyMode != string(ApplyModeBootstrapOnly) {
				t.Fatalf("schema %s = readonly:%t apply_mode:%s, want environment owned", key, item.Readonly, item.ApplyMode)
			}
		}
		if !found {
			t.Fatalf("schema missing %s", key)
		}
	}
	if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "local_oss.domain",
		Value: []byte(`"stale.example.test"`),
	}}); err == nil {
		t.Fatal("SaveValues() accepted an environment-owned public domain")
	}
}

func TestLegacyDisallowRegistrationIsOnlyBootstrapDefault(t *testing.T) {
	svc := newTestService(t)
	cfg.Initial(map[string][]string{
		"default": {"Web:DisallowUserRegister"},
	}, nil)
	t.Cleanup(func() {
		cfg.Initial(nil, nil)
		resetRuntimePolicyForTest()
	})
	bootstrapConfig = nil
	resetRuntimePolicyForTest()
	svc = NewService(svc.db)

	if CurrentRuntimePolicy().AllowUserRegister {
		t.Fatal("AllowUserRegister = true, want legacy bootstrap default false")
	}
	if bootstrapConfig.WebProfile.AllowUserRegister {
		t.Fatal("bootstrap snapshot registration = true, want false")
	}
	if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{{
		Key:   "web_profile.allow_user_register",
		Value: []byte(`true`),
	}}); err != nil {
		t.Fatalf("SaveValues() error = %v", err)
	}
	if !CurrentRuntimePolicy().AllowUserRegister {
		t.Fatal("AllowUserRegister = false after live override, want true")
	}
}

func TestRuntimePolicyConcurrentReadersAndGenericSaves(t *testing.T) {
	svc := newTestService(t)
	const readers = 8
	const updates = 20

	done := make(chan struct{})
	var readersWG sync.WaitGroup
	readersWG.Add(readers)
	for range readers {
		go func() {
			defer readersWG.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = CurrentRuntimePolicy()
				}
			}
		}()
	}

	for i := 0; i < updates; i++ {
		value := i%2 == 0
		raw := []byte(`false`)
		if value {
			raw = []byte(`true`)
		}
		if _, err := svc.SaveValues(context.Background(), []web.AdminSettingValueInput{
			{Key: "web_profile.allow_user_register", Value: raw},
			{Key: "web_profile.allow_phone_bind", Value: raw},
		}); err != nil {
			close(done)
			readersWG.Wait()
			t.Fatalf("SaveValues() update %d error = %v", i, err)
		}
	}
	close(done)
	readersWG.Wait()

	policy := CurrentRuntimePolicy()
	if policy.AllowUserRegister || policy.AllowPhoneBind {
		t.Fatalf("final policy = register:%t phone:%t, want both false", policy.AllowUserRegister, policy.AllowPhoneBind)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(root); err != nil {
		t.Fatalf("os.Chdir(%q) error = %v", root, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	t.Setenv("PAOPAO_CONFIG_FILE", filepath.Join(root, "config.yaml.sample"))
	conf.Initial(nil, false)
	cfg.Initial(map[string][]string{
		"default": {"Web", "Sms", "Meili", "LocalOSS"},
	}, nil)
	conf.WebProfileSetting = &conf.WebProfileConf{
		UseFriendship:             true,
		EnableTrendsBar:           false,
		EnableWallet:              false,
		AllowTweetAttachment:      true,
		AllowTweetAttachmentPrice: true,
		AllowTweetVideo:           true,
		AllowUserRegister:         true,
		AllowPhoneBind:            true,
		DefaultTweetMaxLength:     2000,
		TweetWebEllipsisSize:      400,
		TweetMobileEllipsisSize:   300,
		DefaultTweetVisibility:    "friend",
		DefaultMsgLoopInterval:    5000,
		CopyrightTop:              "fallback-top",
		CopyrightLeft:             "fallback-left",
		CopyrightLeftLink:         "",
		CopyrightRight:            "fallback-right",
		CopyrightRightLink:        "https://fallback.example.com",
	}
	bootstrapConfig = nil
	resetRuntimePolicyForTest()
	databaseName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:sitesetting_"+databaseName+"?mode=memory&cache=shared"), &gorm.Config{NamingStrategy: schema.NamingStrategy{TablePrefix: "p_", SingularTable: true}})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if err := db.AutoMigrate(&settingRecord{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return NewService(db)
}

func hasValue(items []web.AdminSettingValue, key string, expected any, pending bool) bool {
	for _, item := range items {
		if item.Key != key {
			continue
		}
		return item.Value == expected && item.PendingRestart == pending
	}
	return false
}
