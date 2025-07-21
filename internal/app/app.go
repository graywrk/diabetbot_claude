package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"diabetbot/internal/config"
	"diabetbot/internal/database"
	"diabetbot/internal/handlers"
	"diabetbot/internal/services"
	"diabetbot/internal/telegram"

	"github.com/gin-gonic/gin"
)

type App struct {
	config   *config.Config
	db       *database.Database
	bot      *telegram.Bot
	server   *http.Server
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Run() error {
	// Инициализация базы данных
	db, err := database.New(&a.config.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.db = db

	// Миграция базы данных
	if err := db.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Инициализация сервисов
	gigachatService := services.NewGigaChatService(&a.config.GigaChat)
	
	// Инициализация бота
	bot, err := telegram.NewBot(&a.config.Telegram, db.DB, gigachatService)
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}
	a.bot = bot

	// Инициализация веб-сервера
	if err := a.setupServer(); err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}

	// Запуск бота в горутине
	go func() {
		log.Println("Starting Telegram bot...")
		if err := a.bot.Start(); err != nil {
			log.Printf("Bot error: %v", err)
		}
	}()

	// Запуск веб-сервера в горутине
	go func() {
		log.Printf("Starting web server on %s:%s", a.config.Server.Host, a.config.Server.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	return a.waitForShutdown()
}

func (a *App) setupServer() error {
	if a.config.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Telegram-Init-Data")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Инициализация обработчиков API
	apiHandler := handlers.NewAPIHandler(a.db.DB)
	
	// API роуты
	api := router.Group("/api/v1")
	{
		api.GET("/user/:telegram_id", apiHandler.GetUser)
		api.PUT("/user/:telegram_id", apiHandler.UpdateUser)
		api.PUT("/user/:telegram_id/diabetes-info", apiHandler.UpdateDiabetesInfo)
		api.DELETE("/user/:telegram_id/data", apiHandler.DeleteUserData)
		
		api.GET("/glucose/:user_id", apiHandler.GetGlucoseRecords)
		api.POST("/glucose", apiHandler.CreateGlucoseRecord)
		api.PUT("/glucose/:id", apiHandler.UpdateGlucoseRecord)
		api.DELETE("/glucose/:id", apiHandler.DeleteGlucoseRecord)
		api.GET("/glucose/:user_id/stats", apiHandler.GetGlucoseStats)
		
		api.GET("/food/:user_id", apiHandler.GetFoodRecords)
		api.POST("/food", apiHandler.CreateFoodRecord)
		api.PUT("/food/:id", apiHandler.UpdateFoodRecord)
		api.DELETE("/food/:id", apiHandler.DeleteFoodRecord)
	}

	// Статические файлы для веб-приложения
	router.Static("/webapp", "./web/dist")
	
	// SPA fallback - обслуживает index.html для всех маршрутов веб-приложения
	router.NoRoute(func(c *gin.Context) {
		// Если путь начинается с /webapp/ но файл не найден, отдаем index.html для SPA маршрутизации
		if len(c.Request.URL.Path) > 7 && c.Request.URL.Path[:7] == "/webapp" {
			c.File("./web/dist/index.html")
		} else {
			c.JSON(404, gin.H{"error": "Not found"})
		}
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	a.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", a.config.Server.Host, a.config.Server.Port),
		Handler: router,
	}

	return nil
}

func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	if err := a.db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
		return err
	}

	log.Println("Server exited")
	return nil
}