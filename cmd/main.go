package main

import (
	"log"
	"os"

	"diabetbot/internal/app"
	"diabetbot/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize and start application
	application := app.New(cfg)
	
	log.Printf("Starting DiabetBot server on %s:%s", cfg.Server.Host, cfg.Server.Port)
	
	if err := application.Run(); err != nil {
		log.Fatal("Failed to start application:", err)
		os.Exit(1)
	}
}