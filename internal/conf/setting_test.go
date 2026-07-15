// Copyright 2026 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package conf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExplicitConfigAndDeploymentEnvironmentOverridesPreserveSiblings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "runtime.yaml")
	config := []byte(`
Features:
  Default: ["Web", "Sqlite3"]
JWT:
  Secret: yaml-jwt-secret
  Issuer: paopao-api
  Expire: 86400
AdminSettings:
  EncryptionKey: yaml-settings-key
Meili:
  Host: meilisearch:7700
  Index: paopao-data
  ApiKey: yaml-meili-key
MySQL:
  Username: paopao
  Password: yaml-mysql-password
  Host: mysql:3306
  DBName: paopao
LocalOSS:
  SavePath: custom/data/oss
  Bucket: paopao
  Domain: yaml.example.test
`)
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("PAOPAO_CONFIG_FILE", configPath)
	t.Setenv("PAOPAO_JWT_SECRET", "environment-jwt-secret")
	t.Setenv("PAOPAO_ADMIN_SETTINGS_KEY", "environment-settings-key")
	t.Setenv("PAOPAO_MEILI_MASTER_KEY", "environment-meili-key")
	t.Setenv("PAOPAO_MYSQL_PASSWORD", "environment-mysql-password")
	t.Setenv("PAOPAO_PUBLIC_HOST", "community.example.test")
	t.Setenv("PAOPAO_PUBLIC_SECURE", "true")

	vp, err := newViper()
	if err != nil {
		t.Fatalf("newViper() error = %v", err)
	}

	var jwt jwtConf
	if err = vp.UnmarshalKey("JWT", &jwt); err != nil {
		t.Fatalf("UnmarshalKey(JWT) error = %v", err)
	}
	var adminSettings adminSettingsConf
	if err = vp.UnmarshalKey("AdminSettings", &adminSettings); err != nil {
		t.Fatalf("UnmarshalKey(AdminSettings) error = %v", err)
	}
	var meili meiliConf
	if err = vp.UnmarshalKey("Meili", &meili); err != nil {
		t.Fatalf("UnmarshalKey(Meili) error = %v", err)
	}
	var mysql mysqlConf
	if err = vp.UnmarshalKey("MySQL", &mysql); err != nil {
		t.Fatalf("UnmarshalKey(MySQL) error = %v", err)
	}
	var localOSS localossConf
	if err = vp.UnmarshalKey("LocalOSS", &localOSS); err != nil {
		t.Fatalf("UnmarshalKey(LocalOSS) error = %v", err)
	}

	if err = applyDeploymentEnvironmentOverrides(&jwt, &adminSettings, &meili, &mysql, &localOSS); err != nil {
		t.Fatalf("applyDeploymentEnvironmentOverrides() error = %v", err)
	}

	if jwt.Secret != "environment-jwt-secret" || jwt.Issuer != "paopao-api" || jwt.Expire != 86400 {
		t.Errorf("JWT override lost sibling values: %+v", jwt)
	}
	if adminSettings.EncryptionKey != "environment-settings-key" {
		t.Errorf("AdminSettings.EncryptionKey = %q, want environment value", adminSettings.EncryptionKey)
	}
	if meili.ApiKey != "environment-meili-key" || meili.Host != "meilisearch:7700" || meili.Index != "paopao-data" {
		t.Errorf("Meili override lost sibling values: %+v", meili)
	}
	if mysql.Password != "environment-mysql-password" || mysql.UserName != "paopao" || mysql.Host != "mysql:3306" || mysql.DBName != "paopao" {
		t.Errorf("MySQL override lost sibling values: %+v", mysql)
	}
	if localOSS.Domain != "community.example.test" || !localOSS.Secure || localOSS.SavePath != "custom/data/oss" || localOSS.Bucket != "paopao" {
		t.Errorf("LocalOSS override lost sibling values: %+v", localOSS)
	}
}

func TestDeploymentEnvironmentRejectsInvalidPublicSecureValue(t *testing.T) {
	t.Setenv("PAOPAO_PUBLIC_SECURE", "sometimes")
	if err := applyDeploymentEnvironmentOverrides(nil, nil, nil, nil, &localossConf{}); err == nil {
		t.Fatal("applyDeploymentEnvironmentOverrides() error = nil, want invalid boolean error")
	}
}
