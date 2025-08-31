package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// App
	AppName        string `mapstructure:"APP_NAME"`
	AppVersion     string `mapstructure:"APP_VERSION"`
	AppEnvironment string `mapstructure:"APP_ENVIRONMENT"`
	AppLocale      string `mapstructure:"APP_LOCALE"`
	AppURL         string `mapstructure:"APP_URL"`
	AppFrontendURL string `mapstructure:"APP_FRONTEND_URL"`

	// Database
	DatabaseHost         string        `mapstructure:"DATABASE_HOST"`
	DatabasePort         int           `mapstructure:"DATABASE_PORT"`
	DatabaseUser         string        `mapstructure:"DATABASE_USER"`
	DatabasePassword     string        `mapstructure:"DATABASE_PASSWORD"`
	DatabaseName         string        `mapstructure:"DATABASE_DB"`
	DatabaseSSLMode      string        `mapstructure:"DATABASE_SSL_MODE"`
	DatabaseMaxOpenConns int           `mapstructure:"DATABASE_MAX_OPEN_CONNS"`
	DatabaseMaxIdleConns int           `mapstructure:"DATABASE_MAX_IDLE_CONNS"`
	DatabaseMaxLifetime  time.Duration `mapstructure:"DATABASE_MAX_LIFETIME"`
	DatabaseMaxIdleTime  time.Duration `mapstructure:"DATABASE_MAX_IDLE_TIME"`
	DatabasePingTimeout  time.Duration `mapstructure:"DATABASE_PING_TIMEOUT"`
	DatabaseDriver       string        `mapstructure:"DATABASE_DRIVER"`

	// Redis
	RedisHost         string        `mapstructure:"REDIS_HOST"`
	RedisPort         int64         `mapstructure:"REDIS_PORT"`
	RedisPassword     string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB           int           `mapstructure:"REDIS_DB"`
	RedisMinIdleConns int           `mapstructure:"REDIS_MIN_IDLE_CONNS"`
	RedisPoolSize     int           `mapstructure:"REDIS_POOL_SIZE"`
	RedisPoolTimeout  time.Duration `mapstructure:"REDIS_POOL_TIMEOUT"`
	RedisPingTimeout  time.Duration `mapstructure:"REDIS_PING_TIMEOUT"`

	// Email
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUser     string `mapstructure:"SMTP_USER"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	SMTPTimeout  string `mapstructure:"SMTP_TIMEOUT"`

	// Server
	ServerHost            string        `mapstructure:"SERVER_HOST"`
	ServerPort            int           `mapstructure:"SERVER_PORT"`
	ServerReadTimeout     time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	ServerWriteTimeout    time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
	ServerIdleTimeout     time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
	ServerShutdownTimeout time.Duration `mapstructure:"SERVER_SHUTDOWN_TIMEOUT"`

	// Auth
	AuthJWTSecret                  string        `mapstructure:"AUTH_JWT_SECRET"`
	AuthAccessTokenExpiry          time.Duration `mapstructure:"AUTH_ACCESS_TOKEN_EXPIRY"`
	AuthAccessTokenExpiryExtended  time.Duration `mapstructure:"AUTH_ACCESS_TOKEN_EXPIRY_EXTENDED"`
	AuthRefreshTokenExpiry         time.Duration `mapstructure:"AUTH_REFRESH_TOKEN_EXPIRY"`
	AuthRefreshTokenExpiryExtended time.Duration `mapstructure:"AUTH_REFRESH_TOKEN_EXPIRY_EXTENDED"`

	// Logger
	LoggerFile       string `mapstructure:"LOGGER_FILE"`
	LoggerLevel      string `mapstructure:"LOGGER_LEVEL"`
	LoggerMaxSize    int    `mapstructure:"LOGGER_MAX_SIZE"`
	LoggerMaxBackups int    `mapstructure:"LOGGER_MAX_BACKUPS"`
	LoggerMaxAge     int    `mapstructure:"LOGGER_MAX_AGE"`
	LoggerCompress   bool   `mapstructure:"LOGGER_COMPRESS"`
	LoggerOutput     string `mapstructure:"LOGGER_OUTPUT"`

	// Observability
	OtelExporterOtlpEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

// DSN generates the database connection string
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName, c.DatabaseSSLMode)
}

func (c *Config) RedisDSN() string {
	return net.JoinHostPort(c.RedisHost, fmt.Sprintf("%d", c.RedisPort))
}

// NewConfig loads configuration from a .env file (if exists) and environment variables.
func NewConfig() (*Config, error) {
	v := viper.New()

	// Set configuration file details
	v.SetConfigFile(".env")
	v.AddConfigPath(".")

	// Read .env file, if it exists
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		slog.Info("No .env file found, relying on environment variables")
	} else {
		slog.Info(".env file found and loaded")
	}

	// Enable reading from environment variables
	v.AutomaticEnv()
	// Replace dots with underscores for environment variable names
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Ensure case-sensitive matching for mapstructure tags
	v.SetEnvPrefix("")

	// Unmarshal configuration into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate critical fields
	if cfg.DatabaseHost == "" || cfg.DatabaseName == "" || cfg.DatabaseUser == "" {
		return nil, fmt.Errorf("missing required database configuration: DATABASE_HOST, DATABASE_DB, or DATABASE_USER")
	}

	return &cfg, nil
}
