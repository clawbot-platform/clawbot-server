package config

import (
	"fmt"
	"os"
	"time"
)

type Foundation struct {
	PostgresHost string
	PostgresPort string
	PostgresDB   string
	PostgresUser string

	RedisHost string
	RedisPort string

	NATSHost string
	NATSPort string

	MinIOURL      string
	OmniRouteURL  string
	ZeroClawURL   string
	PrometheusURL string
	GrafanaURL    string

	Timeout time.Duration
}

func LoadFoundationFromEnv() (Foundation, error) {
	timeout := getEnv("STACK_SMOKE_TIMEOUT", "5s")
	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		return Foundation{}, fmt.Errorf("parse STACK_SMOKE_TIMEOUT: %w", err)
	}

	cfg := Foundation{
		PostgresHost: getEnv("POSTGRES_HOST", "127.0.0.1"),
		PostgresPort: getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:   getEnv("POSTGRES_DB", "clawbot"),
		PostgresUser: getEnv("POSTGRES_USER", "clawbot"),
		RedisHost:    getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:    getEnv("REDIS_PORT", "6379"),
		NATSHost:     getEnv("NATS_HOST", "127.0.0.1"),
		NATSPort:     getEnv("NATS_PORT", "4222"),
		MinIOURL:     getEnv("STACK_SMOKE_MINIO_URL", fmt.Sprintf("http://%s:%s/minio/health/ready", getEnv("MINIO_HOST", "127.0.0.1"), getEnv("MINIO_PORT", "9000"))),
		OmniRouteURL: getEnv("STACK_SMOKE_OMNIROUTE_URL", fmt.Sprintf("http://%s:%s/", getEnv("OMNIROUTE_HOST", "127.0.0.1"), getEnv("OMNIROUTE_PORT", "20128"))),
		ZeroClawURL:  getEnv("STACK_SMOKE_ZEROCLAW_URL", fmt.Sprintf("http://%s:%s/health", getEnv("ZEROCLAW_HOST", "127.0.0.1"), getEnv("ZEROCLAW_PORT", "3000"))),
		PrometheusURL: getEnv(
			"STACK_SMOKE_PROMETHEUS_URL",
			fmt.Sprintf("http://%s:%s/-/ready", getEnv("PROMETHEUS_HOST", "127.0.0.1"), getEnv("PROMETHEUS_PORT", "9090")),
		),
		GrafanaURL: getEnv(
			"STACK_SMOKE_GRAFANA_URL",
			fmt.Sprintf("http://%s:%s/api/health", getEnv("GRAFANA_HOST", "127.0.0.1"), getEnv("GRAFANA_PORT", "3001")),
		),
		Timeout: parsedTimeout,
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
