package composer

import (
	"time"

	"portal-system/internal/auth"
	"portal-system/internal/config"
	"portal-system/internal/services"
)

type Services struct {
	// services
	AuditLog services.AuditLogger
	Auth     services.AuthService
	User     services.UserService
	Admin    services.AdminService

	// auth layer
	Authenticator *auth.Authenticator
	Authorizer    *auth.Authorizer
}

func newServices(
	cfg *config.Config,
	infra *Infra,
	repos *Repositories,
) *Services {

	// audit
	auditLogService := services.NewAuditLogService(repos.AuditLog)

	// auth service
	authService := services.NewAuthService(services.AuthServiceDeps{
		TxManager:        repos.TxManager,
		AuditLogger:      auditLogService,
		UserRepo:         repos.UserRepo,
		TokenRepo:        repos.TokenRepo,
		RoleRepo:         repos.RoleRepo,
		RefreshTokenRepo: repos.RefreshRepo,
		SessionRepo:      repos.SessionRepo,
		RevoStore:        infra.RevocationStore,
		TokenManager:     infra.TokenManager,
		EmailService:     infra.EmailService,
		FrontendBaseURL:  cfg.FrontEndUrl,
		RefreshTTL:       time.Duration(cfg.RefreshTTL) * time.Second,
	})

	// user service
	userService := services.NewUserService(services.UserServiceDeps{
		TxManager:   repos.TxManager,
		AuditLogger: auditLogService,
		UserRepo:    repos.UserRepo,
		RoleRepo:    repos.RoleRepo,
	})

	// admin service
	adminService := services.NewAdminService(services.AdminServiceDeps{
		TxManager:    repos.TxManager,
		AuditLogger:  auditLogService,
		UserRepo:     repos.UserRepo,
		TokenManager: infra.TokenManager,
		TokenRepo:    repos.TokenRepo,
		RoleRepo:     repos.RoleRepo,
		EmailSvc:     infra.EmailService,
		FrontendURL:  cfg.FrontEndUrl,
	})

	// auth layer
	authenticator := auth.NewAuthenticator(
		infra.TokenManager,
		repos.RoleRepo,
		repos.SessionRepo,
		infra.RevocationStore,
	)

	authorizer := auth.NewAuthorizer()

	return &Services{
		AuditLog: auditLogService,
		Auth:     authService,
		User:     userService,
		Admin:    adminService,

		Authenticator: authenticator,
		Authorizer:    authorizer,
	}
}
