package services

import (
	"context"
	"time"

	repositoriesmocks "portal-system/internal/mocks/repositories"
	servicesmocks "portal-system/internal/mocks/services"
	"portal-system/internal/models"

	"github.com/stretchr/testify/mock"
)

type authServiceTestDeps struct {
	tx          *repositoriesmocks.TxManager
	auditRepo   *repositoriesmocks.AuditLogRepository
	userRepo    *repositoriesmocks.UserRepository
	tokenRepo   *repositoriesmocks.UserTokenRepository
	roleRepo    *repositoriesmocks.RoleRepository
	refreshRepo *repositoriesmocks.RefreshTokenRepository
	sessionRepo *repositoriesmocks.AuthSessionRepository
	revoStore   *servicesmocks.SessionRevocationStore
	tokenMgr    *tokenIssuerMock
	email       *emailSenderMock
}

func newAdminServiceForTest(
	tx *repositoriesmocks.TxManager,
	auditRepo *repositoriesmocks.AuditLogRepository,
	userRepo *repositoriesmocks.UserRepository,
	tokenRepo *repositoriesmocks.UserTokenRepository,
	roleRepo *repositoriesmocks.RoleRepository,
	email *emailSenderMock,
) *AdminService {
	if tx == nil {
		tx = newPassthroughTxManager()
	}
	if auditRepo == nil {
		auditRepo = newAuditLogRepo()
	}
	return NewAdminService(AdminServiceDeps{
		TxManager:   tx,
		AuditLogger: NewAuditLogService(auditRepo),
		UserRepo:    userRepo,
		TokenRepo:   tokenRepo,
		RoleRepo:    roleRepo,
		EmailSvc:    email,
		FrontendURL: "http://frontend.local",
	})
}

func newAuthServiceForTest(deps authServiceTestDeps) *AuthService {
	if deps.tx == nil {
		deps.tx = newPassthroughTxManager()
	}
	if deps.auditRepo == nil {
		deps.auditRepo = newAuditLogRepo()
	}
	if deps.revoStore == nil {
		deps.revoStore = newSessionRevocationStore()
	}
	return NewAuthService(AuthServiceDeps{
		TxManager:        deps.tx,
		AuditLogger:      NewAuditLogService(deps.auditRepo),
		UserRepo:         deps.userRepo,
		RefreshTokenRepo: deps.refreshRepo,
		TokenRepo:        deps.tokenRepo,
		RoleRepo:         deps.roleRepo,
		SessionRepo:      deps.sessionRepo,
		RevoStore:        deps.revoStore,
		TokenManager:     deps.tokenMgr,
		EmailService:     deps.email,
		FrontendBaseURL:  "http://frontend.local",
		RefreshTTL:       24 * time.Hour,
	})
}

func newPassthroughTxManager() *repositoriesmocks.TxManager {
	tx := &repositoriesmocks.TxManager{}
	tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	}).Maybe()
	return tx
}

func newAuditLogRepo() *repositoriesmocks.AuditLogRepository {
	repo := &repositoriesmocks.AuditLogRepository{}
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Maybe()
	repo.EXPECT().List(mock.Anything, mock.Anything).Return([]models.AuditLog(nil), int64(0), nil).Maybe()
	return repo
}

func newSessionRevocationStore() *servicesmocks.SessionRevocationStore {
	store := &servicesmocks.SessionRevocationStore{}
	store.EXPECT().IsRevoked(mock.Anything, mock.Anything).Return(false, nil).Maybe()
	store.EXPECT().MarkRevoked(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return store
}
