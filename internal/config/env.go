// internal/config/env.go
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration from environment
type Config struct {
	RNGServiceURL      string
	SettingsServiceURL string
	ServerPort         string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	return Config{
		RNGServiceURL:      getEnv("RNG_SERVICE_URL", "http://159.89.235.166:17003/api/proxy/rng/1"),
		SettingsServiceURL: getEnv("SETTINGS_SERVICE_URL", "https://t2.ibibe.africa/get-game-settings"),
		ServerPort:         getEnv("SERVER_PORT", "11400"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}