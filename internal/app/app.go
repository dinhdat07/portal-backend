package app

import (
	"fmt"
	"portal-system/internal/auth"
	"portal-system/internal/config"
	"portal-system/internal/http/handlers"
	"portal-system/internal/models"
	"portal-system/internal/platform/email"
	"portal-system/internal/platform/storage"
	"portal-system/internal/platform/token"
	"portal-system/internal/services"
	"time"

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

	// init repo
	userRepo := storage.NewGormUserRepository(db)
	auditLogRepo := storage.NewGormAuditLogRepository(db)
	tokenRepo := storage.NewGormUserTokenRepository(db)
	roleRepo := storage.NewGormRoleRepository(db)
	sessionRepo := storage.NewGormAuthSessionRepository(db)
	txManager := storage.NewGormTxManager(db)

	// auth
	tokenManager := token.New(cfg.JWTSecret, cfg.JWTAccessTTL)
	authenticator := auth.NewAuthenticator(tokenManager, roleRepo)
	authorizer := auth.NewAuthorizer()

	// service
	auditLogService := services.NewAuditLogService(auditLogRepo)

	authService := services.NewAuthService(services.AuthServiceDeps{
		TxManager:       txManager,
		AuditLogger:     auditLogService,
		UserRepo:        userRepo,
		TokenRepo:       tokenRepo,
		RoleRepo:        roleRepo,
		SessionRepo:     sessionRepo,
		TokenManager:    tokenManager,
		EmailService:    emailService,
		FrontendBaseURL: cfg.FrontEndUrl,
		RefreshTTL:      time.Duration(cfg.RefreshTTL) * time.Second,
	})

	userService := services.NewUserService(services.UserServiceDeps{
		TxManager:   txManager,
		AuditLogger: auditLogService,
		UserRepo:    userRepo,
		RoleRepo:    roleRepo,
	})

	adminService := services.NewAdminService(services.AdminServiceDeps{
		TxManager:   txManager,
		AuditLogger: auditLogService,
		UserRepo:    userRepo,
		TokenRepo:   tokenRepo,
		RoleRepo:    roleRepo,
		EmailSvc:    emailService,
		FrontendURL: cfg.FrontEndUrl,
	})

	// handler
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	adminHandler := handlers.NewAdminHandler(adminService, userService)

	router := setupRouter(
		authHandler,
		userHandler,
		adminHandler,
		authenticator,
		authorizer,
	)

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
		&models.User{}, &models.AuditLog{}, &models.UserToken{}, &models.Role{}, &models.Permission{}, &models.AuthSession{},
	)
}
