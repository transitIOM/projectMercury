package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          string
	GTFSFilePath     string
	MessagesFilePath string
	SignalRExpiry    time.Duration
	CuraOwner        string
	CuraRepo         string
	LinearAPIKey     string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Could not load .env file, using environment variables")
	}

	return &Config{
		AppPort:          getEnv("APP_PORT", ":8090"),
		GTFSFilePath:     getEnv("GTFS_FILE_PATH", "GTFS.zip"),
		MessagesFilePath: getEnv("MESSAGES_FILE_PATH", "messages.json"),
		SignalRExpiry:    getEnvDuration("SIGNALR_EXPIRY", 2*time.Minute),
		CuraOwner:        getEnv("CURA_OWNER", "transitIOM"),
		CuraRepo:         getEnv("CURA_REPO", "projectCura"),
		LinearAPIKey:     os.Getenv("LINEAR_API_KEY"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return fallback
}
