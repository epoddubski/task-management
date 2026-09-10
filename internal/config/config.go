package config

import (
	"time"
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
