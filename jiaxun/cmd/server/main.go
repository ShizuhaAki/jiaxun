package main

import (
	"fmt"
	"log"

	"jiaxun/internal/config"
	"jiaxun/internal/handler"
	"jiaxun/internal/middleware"
	"jiaxun/internal/repository"
	"jiaxun/internal/service"

	// Import swagger docs
	_ "jiaxun/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Jiaxun API
// @version 1.0
// @description This is the REST API for the Jiaxun application
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.jiaxun.example.com/support
// @contact.email support@jiaxun.example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	// Load configuration from environment variables or a config file
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Validate the loaded configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize DB connection using GORM
	db, err := repository.InitDB(cfg.Database.Driver, cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Create Gin router
	r := gin.Default()

	// Apply CORS middleware globally
	r.Use(middleware.CORSMiddleware())

	// Add Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Apply JWT middleware except for public routes
	// NOTE: We're NOT applying this globally to avoid blocking Swagger docs and public routes
	// r.Use(middleware.AuthMiddleware())

	// Initialize repositories, services, and handlers
	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(*userRepository)
	handler.NewUserHandler(r, userService)

	// Start the server on the configured port
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting server on %s...", port)
	log.Printf("Swagger documentation available at http://localhost%s/swagger/index.html", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
