package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppEnv                   string
	HTTPAddr                 string
	GRPCAddr                 string
	DBURL                    string
	RedisAddr                string
	RedisSentinels           []string
	RedisSentinelMaster      string
	ImportStorageDir         string
	JWTPrivateKeyPEM         string
	JWTPublicKeyPEM          string
	AccessTokenTTL           time.Duration
	RefreshTokenTTL          time.Duration
	OTLPEndpoint             string
	PrometheusEnabled        bool
	LogLevel                 string
	CorsAllowedOrigins       []string
	WorkerID                 string
	RetentionCleanupInterval time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                   getEnv("APP_ENV", "development"),
		HTTPAddr:                 getEnv("HTTP_ADDR", ":8080"),
		GRPCAddr:                 getEnv("GRPC_ADDR", ":9090"),
		DBURL:                    os.Getenv("DB_URL"),
		RedisAddr:                getEnv("REDIS_ADDR", "localhost:6379"),
		RedisSentinels:           splitCSV(os.Getenv("REDIS_SENTINELS")),
		RedisSentinelMaster:      getEnv("REDIS_SENTINEL_MASTER", "amss-master"),
		ImportStorageDir:         getEnv("IMPORT_STORAGE_DIR", "./data/imports"),
		JWTPrivateKeyPEM:         os.Getenv("JWT_PRIVATE_KEY_PEM"),
		JWTPublicKeyPEM:          os.Getenv("JWT_PUBLIC_KEY_PEM"),
		AccessTokenTTL:           getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:          getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		OTLPEndpoint:             os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		PrometheusEnabled:        getBool("PROMETHEUS_ENABLED", false),
		LogLevel:                 getEnv("LOG_LEVEL", "info"),
		CorsAllowedOrigins:       splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		WorkerID:                 getEnv("WORKER_ID", "worker-1"),
		RetentionCleanupInterval: getDuration("RETENTION_CLEANUP_INTERVAL", 24*time.Hour),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	missing := []string{}
	if c.DBURL == "" {
		missing = append(missing, "DB_URL")
	}
	if c.JWTPrivateKeyPEM == "" {
		missing = append(missing, "JWT_PRIVATE_KEY_PEM")
	}
	if c.JWTPublicKeyPEM == "" {
		missing = append(missing, "JWT_PUBLIC_KEY_PEM")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	if c.AccessTokenTTL <= 0 || c.RefreshTokenTTL <= 0 {
		return errors.New("token TTLs must be positive")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}

func splitCSV(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
