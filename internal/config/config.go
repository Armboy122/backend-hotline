package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Cloudflare  CloudflareConfig  `mapstructure:"cloudflare"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	CORS        CORSConfig        `mapstructure:"cors"`
	Integration IntegrationConfig `mapstructure:"integration"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	SSLMode     string `mapstructure:"sslmode"`
	TimeZone    string `mapstructure:"timezone"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

type CloudflareConfig struct {
	R2 R2Config `mapstructure:"r2"`
}

type R2Config struct {
	AccountID       string `mapstructure:"account_id"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	BucketName      string `mapstructure:"bucket_name"`
	PublicURL       string `mapstructure:"public_url"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpiry  string `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry string `mapstructure:"refresh_token_expiry"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type IntegrationConfig struct {
	ClinicToolKey string `mapstructure:"clinic_tool_key"`
}

// LoadConfig reads the application configuration from config.yaml file.
// It searches for the config file in the current directory and parent directories.
// Values in config.yaml can reference environment variables using ${VAR_NAME} syntax.
// Returns a Config struct populated with values from the YAML file, or an error if loading fails.
func LoadConfig(ctx context.Context) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")

	// โหลด .env ถ้ามี (ไม่ error ถ้าไม่มีไฟล์)
	if f, err := os.Open(".env"); err == nil {
		_ = gotenv.Apply(f)
		f.Close()
	}

	// อ่าน config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand ${VAR} placeholders จาก environment variables
	config.Server.Mode = expandEnv(config.Server.Mode)
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}
	config.Database.Password = expandEnv(config.Database.Password)
	config.Cloudflare.R2.AccountID = expandEnv(config.Cloudflare.R2.AccountID)
	config.Cloudflare.R2.AccessKeyID = expandEnv(config.Cloudflare.R2.AccessKeyID)
	config.Cloudflare.R2.SecretAccessKey = expandEnv(config.Cloudflare.R2.SecretAccessKey)
	config.JWT.Secret = expandEnv(config.JWT.Secret)
	config.Integration.ClinicToolKey = expandEnv(config.Integration.ClinicToolKey)

	// Expand CORS origins และกรอง empty string ออก
	expanded := make([]string, 0, len(config.CORS.AllowedOrigins))
	for _, origin := range config.CORS.AllowedOrigins {
		if v := expandEnv(origin); v != "" {
			expanded = append(expanded, v)
		}
	}
	config.CORS.AllowedOrigins = expanded

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Validate checks that required config values are present and well-formed.
func (c *Config) Validate() error {
	var errs []error

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be between 1 and 65535"))
	}
	switch strings.ToLower(strings.TrimSpace(c.Server.Mode)) {
	case "debug", "release", "test":
	default:
		errs = append(errs, fmt.Errorf("server.mode must be debug, release, or test"))
	}

	if strings.TrimSpace(c.Database.Host) == "" {
		errs = append(errs, fmt.Errorf("database.host is required"))
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errs = append(errs, fmt.Errorf("database.port must be between 1 and 65535"))
	}
	if strings.TrimSpace(c.Database.User) == "" {
		errs = append(errs, fmt.Errorf("database.user is required"))
	}
	if strings.TrimSpace(c.Database.Password) == "" {
		errs = append(errs, fmt.Errorf("database.password is required"))
	}
	if strings.TrimSpace(c.Database.DBName) == "" {
		errs = append(errs, fmt.Errorf("database.dbname is required"))
	}
	if strings.TrimSpace(c.Database.SSLMode) == "" {
		errs = append(errs, fmt.Errorf("database.sslmode is required"))
	}
	if strings.TrimSpace(c.Database.TimeZone) == "" {
		errs = append(errs, fmt.Errorf("database.timezone is required"))
	}

	if strings.TrimSpace(c.Cloudflare.R2.AccountID) == "" {
		errs = append(errs, fmt.Errorf("cloudflare.r2.account_id is required"))
	}
	if strings.TrimSpace(c.Cloudflare.R2.AccessKeyID) == "" {
		errs = append(errs, fmt.Errorf("cloudflare.r2.access_key_id is required"))
	}
	if strings.TrimSpace(c.Cloudflare.R2.SecretAccessKey) == "" {
		errs = append(errs, fmt.Errorf("cloudflare.r2.secret_access_key is required"))
	}
	if strings.TrimSpace(c.Cloudflare.R2.BucketName) == "" {
		errs = append(errs, fmt.Errorf("cloudflare.r2.bucket_name is required"))
	}
	if strings.TrimSpace(c.Cloudflare.R2.PublicURL) == "" {
		errs = append(errs, fmt.Errorf("cloudflare.r2.public_url is required"))
	}

	if strings.TrimSpace(c.JWT.Secret) == "" {
		errs = append(errs, fmt.Errorf("jwt.secret is required"))
	}
	if _, err := time.ParseDuration(c.JWT.AccessTokenExpiry); err != nil {
		errs = append(errs, fmt.Errorf("jwt.access_token_expiry must be a valid duration: %w", err))
	}
	if _, err := time.ParseDuration(c.JWT.RefreshTokenExpiry); err != nil {
		errs = append(errs, fmt.Errorf("jwt.refresh_token_expiry must be a valid duration: %w", err))
	}

	if len(c.CORS.AllowedOrigins) == 0 {
		errs = append(errs, fmt.Errorf("cors.allowed_origins must include at least one origin"))
	}
	if len(c.CORS.AllowedMethods) == 0 {
		errs = append(errs, fmt.Errorf("cors.allowed_methods must include at least one method"))
	}
	if len(c.CORS.AllowedHeaders) == 0 {
		errs = append(errs, fmt.Errorf("cors.allowed_headers must include at least one header"))
	}

	return errors.Join(errs...)
}

// expandEnv แทนที่ ${VAR_NAME} ด้วยค่าจาก environment variable
func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		key := s[2 : len(s)-1]
		return os.Getenv(key)
	}
	return s
}
