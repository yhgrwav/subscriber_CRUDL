package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type HTTPConfig struct {
	AppPort         string `env:"APP_PORT" envDefault:"8080"`
	ReadTimeoutSec  int    `env:"HTTP_READ_TIMEOUT_SEC" envDefault:"10"`
	WriteTimeoutSec int    `env:"HTTP_WRITE_TIMEOUT_SEC" envDefault:"10"`
}

type DBConfig struct {
	PostgresDSN        string `env:"POSTGRES_DSN,required"`
	MaxOpen            int    `env:"DB_MAX_OPEN" envDefault:"10"`
	MaxIdle            int    `env:"DB_MAX_IDLE" envDefault:"5"`
	ConnMaxLifetimeMin int    `env:"DB_CONN_MAX_LIFETIME_MIN" envDefault:"30"`
}

type LoggerConfig struct {
	Level string `env:"LOG_LEVEL" envDefault:"INFO"`
}

type SwaggerConfig struct {
	Host     string `env:"SWAGGER_HOST" envDefault:"localhost:8080"`
	BasePath string `env:"SWAGGER_BASE_PATH" envDefault:"/api/v1"`
}

type Config struct {
	HTTP    HTTPConfig
	DB      DBConfig
	Logger  LoggerConfig
	Swagger SwaggerConfig
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{}
	if err := env.Parse(&cfg.HTTP); err != nil {
		return Config{}, fmt.Errorf("ошибка парсинга конфигурации API: %w", err)
	}
	if err := env.Parse(&cfg.DB); err != nil {
		return Config{}, fmt.Errorf("ошибка парсинга конфигурации БД: %w", err)
	}
	if err := env.Parse(&cfg.Logger); err != nil {
		return Config{}, fmt.Errorf("ошибка парсинга конфигурации логгера: %w", err)
	}
	if err := env.Parse(&cfg.Swagger); err != nil {
		return Config{}, fmt.Errorf("ошибка парсинга сконфигурации сваггера: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("ошибка валидации конфига: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Logger.Level == "" {
		c.Logger.Level = "INFO"
	}
	if c.HTTP.AppPort == "" {
		c.HTTP.AppPort = "8080"
	}
	if c.HTTP.ReadTimeoutSec <= 0 {
		c.HTTP.ReadTimeoutSec = 10
	}
	if c.HTTP.WriteTimeoutSec <= 0 {
		c.HTTP.WriteTimeoutSec = 10
	}
	u, err := url.Parse(c.DB.PostgresDSN)
	if err != nil {
		return fmt.Errorf("невалидный POSTGRES_DSN: %w", err)
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return errors.New("POSTGRES_DSN должен использовать postgres или postgresql scheme")
	}
	if c.DB.MaxOpen < 1 {
		c.DB.MaxOpen = 10
	}
	if c.DB.MaxIdle < 0 {
		c.DB.MaxIdle = 5
	}
	if c.DB.ConnMaxLifetimeMin <= 0 {
		c.DB.ConnMaxLifetimeMin = 30
	}
	if c.Swagger.Host == "" {
		c.Swagger.Host = "localhost:8080"
	}
	if c.Swagger.BasePath == "" {
		c.Swagger.BasePath = "/api/v1"
	}
	return nil
}

func (c *Config) HTTPReadTimeout() time.Duration {
	return time.Duration(c.HTTP.ReadTimeoutSec) * time.Second
}

func (c *Config) HTTPWriteTimeout() time.Duration {
	return time.Duration(c.HTTP.WriteTimeoutSec) * time.Second
}
