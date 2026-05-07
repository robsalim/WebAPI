package main

import (
	"log"
	"os"
	"webapi/config"
	"webapi/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию.
	cfg := config.LoadConfig()

	// Создаем роутер
	router := gin.Default()

	// Настройка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Инициализация обработчиков
	handlers := controllers.NewHandlers(cfg)

	// Регистрация маршрутов
	router.GET("/health", handlers.Health)
	router.POST("/restart", handlers.Restart)
	router.GET("/time-diff", handlers.TimeDiff)
	router.GET("/debug/privileges", handlers.CheckPrivileges)
	router.GET("/db-health", handlers.GetDatabaseHealth)
	router.GET("/data-delays", handlers.GetDataDelays)

	// Вывод информации о запуске
	log.Printf("🚀 API запущена под пользователем: %s", os.Getenv("USERNAME"))
	log.Printf("🛡️  Права администратора: %v", cfg.IsAdmin)
	log.Printf("📁 IServer путь: %s", cfg.IServer.Path)
	log.Printf("🗄️  База данных: %s", cfg.Database.Base)
	log.Printf("⏱️  Задержка (часов): %d", cfg.Delays.DeltaHour)
	log.Printf("🕐 NTP сервер: %s", cfg.NTP.Server)
	log.Printf("🌐 Сервер запущен на порту :9200")

	// Запуск сервера
	router.Run("0.0.0.0:9200")
}
