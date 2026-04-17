package app

import (
	"portal-system/internal/auth"
	"portal-system/internal/config"
	portalgrpc "portal-system/internal/grpc"
	"portal-system/internal/http/handlers"
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

	// transport - http
	Router *gin.Engine

	// auth
	Authenticator *auth.Authenticator
	Authorizer    *auth.Authorizer

	// http handlers
	AuthHandler  *handlers.AuthHandler
	UserHandler  *handlers.UserHandler
	AdminHandler *handlers.AdminHandler

	// grpc servers
	AuthGRPC  *portalgrpc.AuthServer
	UserGRPC  *portalgrpc.UserServer
	AdminGRPC *portalgrpc.AdminServer
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

	if err := AutoMigrate(db); err != nil {
		return nil, err
	}

	smtpCfg, err := config.LoadSMTPConfig()
	if err != nil {
		return nil, err
	}
	emailService := email.NewSMTPEmailService(*smtpCfg)

	userRepo := storage.NewGormUserRepository(db)
	auditLogRepo := storage.NewGormAuditLogRepository(db)
	tokenRepo := storage.NewGormUserTokenRepository(db)
	roleRepo := storage.NewGormRoleRepository(db)
	sessionRepo := storage.NewGormAuthSessionRepository(db)
	refreshTokenRepo := storage.NewGormRefreshTokenRepository(db)
	txManager := storage.NewGormTxManager(db)

	tokenManager := token.New(cfg.JWTSecret, cfg.JWTAccessTTL)
	authenticator := auth.NewAuthenticator(tokenManager, roleRepo, sessionRepo)
	authorizer := auth.NewAuthorizer()

	auditLogService := services.NewAuditLogService(auditLogRepo)

	authService := services.NewAuthService(services.AuthServiceDeps{
		TxManager:        txManager,
		AuditLogger:      auditLogService,
		UserRepo:         userRepo,
		TokenRepo:        tokenRepo,
		RoleRepo:         roleRepo,
		SessionRepo:      sessionRepo,
		RefreshTokenRepo: refreshTokenRepo,
		TokenManager:     tokenManager,
		EmailService:     emailService,
		FrontendBaseURL:  cfg.FrontEndUrl,
		RefreshTTL:       time.Duration(cfg.RefreshTTL) * time.Second,
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

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	adminHandler := handlers.NewAdminHandler(adminService, userService)

	authGRPC := portalgrpc.NewAuthServer(authService)
	userGRPC := portalgrpc.NewUserServer(userService)
	adminGRPC := portalgrpc.NewAdminServer(adminService, userService)

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
		Config:        cfg,
		DB:            db,
		Authenticator: authenticator,
		Authorizer:    authorizer,
		AuthHandler:   authHandler,
		UserHandler:   userHandler,
		AdminHandler:  adminHandler,
		AuthGRPC:      authGRPC,
		UserGRPC:      userGRPC,
		AdminGRPC:     adminGRPC,
	}, nil
}
