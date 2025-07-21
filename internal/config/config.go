package config

import (
	"os"
)

type Config struct {
	Telegram TelegramConfig
	GigaChat GigaChatConfig
	Database DatabaseConfig
	Server   ServerConfig
}

type TelegramConfig struct {
	BotToken   string
	WebhookURL string
}

type GigaChatConfig struct {
	APIKey  string
	BaseURL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

func Load() *Config {
	return &Config{
		Telegram: TelegramConfig{
			BotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
			WebhookURL: getEnv("TELEGRAM_WEBHOOK_URL", ""),
		},
		GigaChat: GigaChatConfig{
			APIKey:  getEnv("GIGACHAT_API_KEY", ""),
			BaseURL: getEnv("GIGACHAT_BASE_URL", "https://gigachat.devices.sberbank.ru"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "diabetbot"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "diabetbot"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Env:  getEnv("ENVIRONMENT", "development"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}