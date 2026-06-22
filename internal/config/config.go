package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP    HTTPConfig
	DB      DBConfig
	JWT     JWTConfig
	Log     LogConfig
	Storage StorageConfig
	Brevo   BrevoConfig
}

type StorageConfig struct {
	UploadDir   string `mapstructure:"upload_dir"`
	S3Endpoint  string `mapstructure:"s3_endpoint"`
	S3Bucket    string `mapstructure:"s3_bucket"`
	S3AccessKey string `mapstructure:"s3_access_key"`
	S3SecretKey string `mapstructure:"s3_secret_key"`
	S3Region    string `mapstructure:"s3_region"`
	PresignTTL  string `mapstructure:"presign_ttl"` // e.g. "1h"
}

type HTTPConfig struct {
	Addr string
}

type DBConfig struct {
	DSN string
}

type JWTConfig struct {
	Secret string
	TTL    string
}

type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

type BrevoConfig struct {
	APIKey string `mapstructure:"api_key"`
	From   string `mapstructure:"from"`
	Name   string `mapstructure:"name"`
	Host   string `mapstructure:"host"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("db.dsn", "")
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.ttl", "24h")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("storage.upload_dir", "./uploads")
	v.SetDefault("storage.s3_region", "ru-central1")
	v.SetDefault("storage.presign_ttl", "1h")
	v.SetDefault("brevo.api_key", "")
	v.SetDefault("brevo.from", "")
	v.SetDefault("brevo.name", "")
	v.SetDefault("brevo.host", "")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.DB.DSN == "" {
		return nil, errors.New("db.dsn is required (APP_DB_DSN)")
	}

	if cfg.JWT.Secret == "" {
		return nil, errors.New("jwt.secret is required (APP_JWT_SECRET)")
	}

	if cfg.Brevo.APIKey == "" {
		return nil, errors.New("brevo.api_key is required (APP_BREVO_API_KEY)")
	}

	return &cfg, nil
}
