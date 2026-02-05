package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration from environment
type Config struct {
	RNGServiceURL      string
	SettingsServiceURL string
	ServerPort         string
	LogFile            string
	TelegramBotToken   string
	TelegramChatID     string
}

// Load loads configuration from environment variables
func Load() Config {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	return Config{
		RNGServiceURL:      getEnv("RNG_API_URL", "http://127.0.0.1:17004/api/proxy/rng/1"),
		SettingsServiceURL: getEnv("SETTINGS_API_URL", "http://127.0.0.1:4040/get-game-settings"),
		ServerPort:         getEnv("PORT", "11400"),
		LogFile:            getEnv("LOG_FILE", "app.log"),
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:     getEnv("TELEGRAM_CHAT_ID", ""),
	}
}

// Function to get an environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// LoadAll loads both production and test configurations from environment variables
func LoadAll() (prod Config, test Config) {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	prod = Config{
		RNGServiceURL:      getEnv("PROD_RNG_API_URL", "http://127.0.0.1:17004/api/proxy/rng/1"),
		SettingsServiceURL: getEnv("PROD_SETTINGS_API_URL", "http://127.0.0.1:4040/get-game-settings"),
		ServerPort:         getEnv("PORT", "11400"),
		LogFile:            getEnv("LOG_FILE", "app.log"),
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:     getEnv("TELEGRAM_CHAT_ID", ""),
	}
	test = Config{
		RNGServiceURL:      getEnv("TEST_RNG_API_URL", "http://127.0.0.1:17004/api/proxy/rng/1"),
		SettingsServiceURL: getEnv("TEST_SETTINGS_API_URL", "http://127.0.0.1:4040/get-game-settings"),
		ServerPort:         getEnv("PORT", "11400"),
		LogFile:            getEnv("LOG_FILE", "app.log"),
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:     getEnv("TELEGRAM_CHAT_ID", ""),
	}
	return
}
