package app

import (
	"fmt"
	"portal-system/internal/config"
	"portal-system/internal/handlers"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	"portal-system/internal/services"
	"portal-system/internal/token"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Router *gin.Engine
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// migrate
	if err := AutoMigrate(db); err != nil {
		return nil, err
	}

	// init
	tokenManager := token.New(cfg.JWTSecret, cfg.JWTAccessTTL)
	userRepo := repositories.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, tokenManager)
	authHandler := handlers.NewAuthHandler(authService)

	router := setupRouter(authHandler)

	return &App{
		Config: cfg,
		DB:     db,
		Router: router,
	}, nil

}

func (a *App) Run() error {
	return a.Router.Run(fmt.Sprintf(":%s", a.Config.Port))
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
	)
}
