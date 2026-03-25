package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"portal-system/internal/config"
	"portal-system/internal/handlers"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	"portal-system/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// migrate
	db.AutoMigrate(&models.User{})

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register", userHandler.Register)
	}

	r.Run(":8080")
}
