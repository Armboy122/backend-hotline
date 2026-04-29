package config_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"backend-hotlines3/internal/config"
	"github.com/spf13/viper"
)

func TestLoadConfig_ExpandsEnvironmentAndFiltersEmptyOrigins(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("MODE", "release")
	t.Setenv("DB_PASSWORD", "db-secret")
	t.Setenv("CF_ACCOUNT_ID", "acct-123")
	t.Setenv("CF_ACCESS_KEY_ID", "access-123")
	t.Setenv("CF_SECRET_ACCESS_KEY", "secret-123")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("ORIGIN_ONE", "https://one.example")
	t.Setenv("ORIGIN_TWO", "https://two.example")

	tmp := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	configPath := filepath.Join(tmp, "config.yaml")
	configContent := []byte(`server:
  port: 8081
  mode: ${MODE}

database:
  host: localhost
  port: 5432
  user: hotline
  password: ${DB_PASSWORD}
  dbname: hotline
  sslmode: require
  timezone: Asia/Bangkok
  auto_migrate: true

cloudflare:
  r2:
    account_id: ${CF_ACCOUNT_ID}
    access_key_id: ${CF_ACCESS_KEY_ID}
    secret_access_key: ${CF_SECRET_ACCESS_KEY}
    bucket_name: hotline
    public_url: https://cdn.example

jwt:
  secret: ${JWT_SECRET}
  access_token_expiry: 1h
  refresh_token_expiry: 168h

cors:
  allowed_origins:
    - ${ORIGIN_ONE}
    - ${ORIGIN_TWO}
    - ${ORIGIN_THREE}
  allowed_methods:
    - GET
  allowed_headers:
    - Content-Type
`)
	if err := os.WriteFile(configPath, configContent, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Server.Mode != "release" {
		t.Fatalf("Server.Mode = %q, want release", cfg.Server.Mode)
	}
	if cfg.Database.Password != "db-secret" {
		t.Fatalf("Database.Password = %q, want db-secret", cfg.Database.Password)
	}
	if cfg.Cloudflare.R2.AccountID != "acct-123" {
		t.Fatalf("R2.AccountID = %q, want acct-123", cfg.Cloudflare.R2.AccountID)
	}
	if cfg.JWT.Secret != "jwt-secret" {
		t.Fatalf("JWT.Secret = %q, want jwt-secret", cfg.JWT.Secret)
	}

	wantOrigins := []string{"https://one.example", "https://two.example"}
	if !reflect.DeepEqual(cfg.CORS.AllowedOrigins, wantOrigins) {
		t.Fatalf("CORS.AllowedOrigins = %#v, want %#v", cfg.CORS.AllowedOrigins, wantOrigins)
	}
}

func TestLoadConfig_ValidatesRequiredValues(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("MODE", "release")
	t.Setenv("DB_PASSWORD", "db-secret")
	t.Setenv("CF_ACCOUNT_ID", "acct-123")
	t.Setenv("CF_ACCESS_KEY_ID", "access-123")
	t.Setenv("CF_SECRET_ACCESS_KEY", "secret-123")
	t.Setenv("JWT_SECRET", "")

	tmp := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	configPath := filepath.Join(tmp, "config.yaml")
	configContent := []byte(`server:
  port: 8081
  mode: ${MODE}

database:
  host: localhost
  port: 5432
  user: hotline
  password: ${DB_PASSWORD}
  dbname: hotline
  sslmode: require
  timezone: Asia/Bangkok
  auto_migrate: true

cloudflare:
  r2:
    account_id: ${CF_ACCOUNT_ID}
    access_key_id: ${CF_ACCESS_KEY_ID}
    secret_access_key: ${CF_SECRET_ACCESS_KEY}
    bucket_name: hotline
    public_url: https://cdn.example

jwt:
  secret: ${JWT_SECRET}
  access_token_expiry: 1h
  refresh_token_expiry: 168h

cors:
  allowed_origins:
    - https://one.example
  allowed_methods:
    - GET
  allowed_headers:
    - Content-Type
`)
	if err := os.WriteFile(configPath, configContent, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err = config.LoadConfig(context.Background())
	if err == nil {
		t.Fatal("LoadConfig returned nil error, want validation failure")
	}
	if !strings.Contains(err.Error(), "jwt.secret is required") {
		t.Fatalf("LoadConfig error = %v, want jwt.secret validation failure", err)
	}
}
