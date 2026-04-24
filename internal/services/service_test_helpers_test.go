package services_test

import (
	"context"
	"time"

	repositoriesmocks "portal-system/internal/mocks/repositories"
	servicesmocks "portal-system/internal/mocks/services"
	. "portal-system/internal/services"

	"github.com/stretchr/testify/mock"
)

type authServiceTestDeps struct {
	tx          *repositoriesmocks.TxManager
	auditLogger *servicesmocks.AuditLogger
	userRepo    *repositoriesmocks.UserRepository
	tokenRepo   *repositoriesmocks.UserTokenRepository
	roleRepo    *repositoriesmocks.RoleRepository
	refreshRepo *repositoriesmocks.RefreshTokenRepository
	sessionRepo *repositoriesmocks.AuthSessionRepository
	revoStore   *servicesmocks.SessionRevocationStore
	tokenMgr    *servicesmocks.TokenIssuer
	email       *servicesmocks.EmailSender
}

func newAdminServiceForTest(
	tx *repositoriesmocks.TxManager,
	auditLogger *servicesmocks.AuditLogger,
	userRepo *repositoriesmocks.UserRepository,
	tokenRepo *repositoriesmocks.UserTokenRepository,
	roleRepo *repositoriesmocks.RoleRepository,
	tokenMgr *servicesmocks.TokenIssuer,
	email *servicesmocks.EmailSender,
) AdminService {
	if tx == nil {
		tx = newPassthroughTxManager()
	}
	if auditLogger == nil {
		auditLogger = newAuditLoggerMock()
	}
	if tokenRepo == nil {
		tokenRepo = &repositoriesmocks.UserTokenRepository{}
	}
	if roleRepo == nil {
		roleRepo = &repositoriesmocks.RoleRepository{}
	}
	if tokenMgr == nil {
		tokenMgr = newTokenIssuerMock()
	}
	if email == nil {
		email = newEmailSenderMock()
	}
	return NewAdminService(AdminServiceDeps{
		TxManager:    tx,
		AuditLogger:  auditLogger,
		UserRepo:     userRepo,
		TokenManager: tokenMgr,
		TokenRepo:    tokenRepo,
		RoleRepo:     roleRepo,
		EmailSvc:     email,
		FrontendURL:  "http://frontend.local",
	})
}

func newAuthServiceForTest(deps authServiceTestDeps) AuthService {
	if deps.tx == nil {
		deps.tx = newPassthroughTxManager()
	}
	if deps.auditLogger == nil {
		deps.auditLogger = newAuditLoggerMock()
	}
	if deps.revoStore == nil {
		deps.revoStore = newSessionRevocationStore()
	}
	if deps.tokenMgr == nil {
		deps.tokenMgr = newTokenIssuerMock()
	}
	if deps.email == nil {
		deps.email = newEmailSenderMock()
	}
	return NewAuthService(AuthServiceDeps{
		TxManager:        deps.tx,
		AuditLogger:      deps.auditLogger,
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

func newAuditLoggerMock() *servicesmocks.AuditLogger {
	logger := &servicesmocks.AuditLogger{}
	logger.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	logger.EXPECT().LogWithMetadata(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return logger
}

func newSessionRevocationStore() *servicesmocks.SessionRevocationStore {
	store := &servicesmocks.SessionRevocationStore{}
	store.EXPECT().IsRevoked(mock.Anything, mock.Anything).Return(false, nil).Maybe()
	store.EXPECT().MarkRevoked(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return store
}

func newTokenIssuerMock() *servicesmocks.TokenIssuer {
	tokenMgr := &servicesmocks.TokenIssuer{}
	tokenMgr.EXPECT().GenerateAccessToken(mock.Anything).Return("access-token", nil).Maybe()
	tokenMgr.EXPECT().GenerateRefreshToken().Return("refresh-token", nil).Maybe()
	tokenMgr.EXPECT().ExpiresInSeconds().Return(3600).Maybe()
	tokenMgr.EXPECT().Parse(mock.Anything).Return(nil, nil).Maybe()
	tokenMgr.EXPECT().HashToken(mock.Anything).RunAndReturn(func(raw string) string {
		return "hashed-" + raw
	}).Maybe()
	tokenMgr.EXPECT().GenerateHashToken().Return("token-hash", "raw-token", nil).Maybe()
	return tokenMgr
}

func newEmailSenderMock() *servicesmocks.EmailSender {
	email := &servicesmocks.EmailSender{}
	email.EXPECT().SendVerificationEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	email.EXPECT().SendResetPasswordEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	email.EXPECT().SendSetPasswordEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return email
}
