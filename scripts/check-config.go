package main

import (
	"fmt"
	"os"
	"diabetbot/internal/config"
)

func main() {
	fmt.Println("=== DiabetBot Configuration Check ===")
	
	cfg := config.Load()
	
	fmt.Printf("Telegram Bot Token: %s\n", maskSecret(cfg.Telegram.BotToken))
	fmt.Printf("Telegram Webhook URL: %s\n", cfg.Telegram.WebhookURL)
	fmt.Printf("WebApp URL: %s\n", cfg.Telegram.WebAppURL)
	fmt.Printf("GigaChat API Key: %s\n", maskSecret(cfg.GigaChat.APIKey))
	fmt.Printf("Database Host: %s\n", cfg.Database.Host)
	fmt.Printf("Database Port: %s\n", cfg.Database.Port)
	fmt.Printf("Server Host: %s\n", cfg.Server.Host)
	fmt.Printf("Server Port: %s\n", cfg.Server.Port)
	
	fmt.Println("\n=== Environment Variables ===")
	fmt.Printf("WEBAPP_URL: %s\n", os.Getenv("WEBAPP_URL"))
	fmt.Printf("TELEGRAM_BOT_TOKEN: %s\n", maskSecret(os.Getenv("TELEGRAM_BOT_TOKEN")))
	
	if cfg.Telegram.WebAppURL == "" {
		fmt.Println("\n❌ WARNING: WebApp URL is not configured!")
		fmt.Println("Set WEBAPP_URL environment variable to fix this.")
	} else {
		fmt.Printf("\n✅ WebApp URL is configured: %s\n", cfg.Telegram.WebAppURL)
		fmt.Printf("Generated WebApp link would be: %s/webapp\n", cfg.Telegram.WebAppURL)
	}
}

func maskSecret(secret string) string {
	if secret == "" {
		return "(not set)"
	}
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}