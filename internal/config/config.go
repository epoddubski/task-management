package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	Database               DBConfig
	JWTSecret              string
	TokenTTL               time.Duration
	DeadlineWorkerInterval time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Info(err.Error())
	}

	cfg := &Config{
		Port: getEnv("APP_PORT", "8080"),
		Database: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
		},
		JWTSecret:              getEnv("JWT_SECRET", ""),
		TokenTTL:               getDurationEnv("TOKEN_TTL", 15*time.Minute),
		DeadlineWorkerInterval: getDurationEnv("DEADLINE_WORKER_INTERVAL", 10*time.Minute),
	}

	if cfg.Database.User == "" || cfg.Database.Password == "" || cfg.Database.Name == "" {
		return nil, fmt.Errorf("database connection parameters are required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
