package app

import (
	"fmt"
	"portal-system/internal/auth"
	"portal-system/internal/config"
	"portal-system/internal/http/handlers"
	"portal-system/internal/models"
	"portal-system/internal/platform/email"
	"portal-system/internal/platform/token"
	"portal-system/internal/repositories"
	"portal-system/internal/services"

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

	// email service
	smtpCfg, err := config.LoadSMTPConfig()
	if err != nil {
		return nil, err
	}
	emailService := email.NewSMTPEmailService(*smtpCfg)

	// init
	userRepo := repositories.NewUserRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)
	tokenRepo := repositories.NewUserTokenRepository(db)
	roleRepo := repositories.NewRoleRepository(db)

	tokenManager := token.New(cfg.JWTSecret, cfg.JWTAccessTTL)
	authenticator := auth.NewAuthenticator(tokenManager, roleRepo)
	authorizer := auth.NewAuthorizer()

	auditLogService := services.NewAuditLogService(auditLogRepo)
	authService := services.NewAuthService(db, userRepo, tokenRepo, roleRepo, tokenManager, auditLogService, emailService, cfg.FrontEndUrl)
	userService := services.NewUserService(db, userRepo, roleRepo, auditLogService)
	adminService := services.NewAdminService(db, userRepo, tokenRepo, roleRepo, auditLogService, emailService, cfg.FrontEndUrl)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	adminHandler := handlers.NewAdminHandler(adminService, userService)
	router := setupRouter(authHandler, userHandler, adminHandler, authenticator, authorizer)

	if cfg.Env == "development" {
		if err := seedPermissions(db); err != nil {
			return nil, err
		}
		if err := seedRoles(db); err != nil {
			return nil, err
		}
		if err := seedRolePermissions(db); err != nil {
			return nil, err
		}
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

func (a *App) Run() error {
	return a.Router.Run(fmt.Sprintf(":%s", a.Config.Port))
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{}, &models.AuditLog{}, &models.UserToken{}, &models.Role{}, &models.Permission{},
	)
}
