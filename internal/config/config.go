package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
	JWT  JWTConfig
	Log  LogConfig
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

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("db.dsn", "")
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.ttl", "24h")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

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

	return &cfg, nil
}
