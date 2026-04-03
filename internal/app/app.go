package app

import (
	"fmt"
	"portal-system/internal/config"
	"portal-system/internal/domain/enum"
	"portal-system/internal/http/handlers"
	"portal-system/internal/models"
	"portal-system/internal/platform/email"
	"portal-system/internal/platform/token"
	"portal-system/internal/repositories"
	"portal-system/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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

	// email service
	smtpCfg, err := config.LoadSMTPConfig()
	if err != nil {
		return nil, err
	}
	emailService := email.NewSMTPEmailService(*smtpCfg)

	// init
	tokenManager := token.New(cfg.JWTSecret, cfg.JWTAccessTTL)
	userRepo := repositories.NewUserRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)
	tokenRepo := repositories.NewUserTokenRepository(db)

	auditLogService := services.NewAuditLogService(auditLogRepo)
	authService := services.NewAuthService(db, userRepo, tokenRepo, tokenManager, auditLogService, emailService, cfg.FrontEndUrl)
	userService := services.NewUserService(db, userRepo, auditLogService)
	adminService := services.NewAdminService(db, userRepo, tokenRepo, auditLogService, emailService, cfg.FrontEndUrl)

	authHandler := handlers.NewAuthHandler(authService, cfg)
	userHandler := handlers.NewUserHandler(userService)
	adminHandler := handlers.NewAdminHandler(adminService, userService)
	router := setupRouter(authHandler, userHandler, adminHandler, tokenManager, cfg)

	if cfg.Env == "development" {
		if err := seedAdmin(db, cfg); err != nil {
			return nil, err
		}
	}

	return &App{
		Config: cfg,
		DB:     db,
		Router: router,
	}, nil

}

func seedAdmin(db *gorm.DB, cfg *config.Config) error {
	var existing models.User
	err := db.Where("email = ?", cfg.AdminEmail).First(&existing).Error
	if err == nil {
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashStr := string(hash)
	now := time.Now()

	admin := &models.User{
		Email:           cfg.AdminEmail,
		Username:        "admin",
		FirstName:       "System",
		LastName:        "Admin",
		PasswordHash:    &hashStr,
		Role:            enum.RoleAdmin,
		Status:          enum.StatusActive,
		EmailVerifiedAt: &now,
	}

	return db.Create(admin).Error
}

func (a *App) Run() error {
	return a.Router.Run(fmt.Sprintf(":%s", a.Config.Port))
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{}, &models.AuditLog{}, &models.UserToken{},
	)
}
