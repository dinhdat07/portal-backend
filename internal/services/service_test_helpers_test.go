package services

import (
	"time"
)

func ptrTime(v time.Time) *time.Time {
	return &v
}

type authServiceTestDeps struct {
	tx          *txManagerMock
	auditRepo   *auditRepoMock
	userRepo    *userRepoMock
	tokenRepo   *tokenRepoMock
	roleRepo    *roleRepoMock
	refreshRepo *refreshTokenRepoMock
	sessionRepo *sessionRepoMock
	revoStore   *sessionRevocationStoreMock
	tokenMgr    *tokenIssuerMock
	email       *emailSenderMock
}

func newAdminServiceForTest(tx *txManagerMock, auditRepo *auditRepoMock, userRepo *userRepoMock, tokenRepo *tokenRepoMock, roleRepo *roleRepoMock, email *emailSenderMock) *AdminService {
	if tx == nil {
		tx = &txManagerMock{}
	}
	if auditRepo == nil {
		auditRepo = &auditRepoMock{}
	}
	if userRepo == nil {
		userRepo = &userRepoMock{}
	}
	if tokenRepo == nil {
		tokenRepo = &tokenRepoMock{}
	}
	if roleRepo == nil {
		roleRepo = &roleRepoMock{}
	}
	if email == nil {
		email = &emailSenderMock{}
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
		deps.tx = &txManagerMock{}
	}
	if deps.auditRepo == nil {
		deps.auditRepo = &auditRepoMock{}
	}
	if deps.userRepo == nil {
		deps.userRepo = &userRepoMock{}
	}
	if deps.tokenRepo == nil {
		deps.tokenRepo = &tokenRepoMock{}
	}
	if deps.roleRepo == nil {
		deps.roleRepo = &roleRepoMock{}
	}
	if deps.refreshRepo == nil {
		deps.refreshRepo = &refreshTokenRepoMock{}
	}
	if deps.sessionRepo == nil {
		deps.sessionRepo = &sessionRepoMock{}
	}
	if deps.revoStore == nil {
		deps.revoStore = &sessionRevocationStoreMock{}
	}
	if deps.tokenMgr == nil {
		deps.tokenMgr = &tokenIssuerMock{}
	}
	if deps.email == nil {
		deps.email = &emailSenderMock{}
	}

	return NewAuthService(AuthServiceDeps{
		TxManager:        deps.tx,
		AuditLogger:      NewAuditLogService(deps.auditRepo),
		UserRepo:         deps.userRepo,
		RefreshTokenRepo: deps.refreshRepo,
		TokenRepo:        deps.tokenRepo,
		RoleRepo:         deps.roleRepo,
		SessionRepo:      deps.sessionRepo,
		revoStore:        deps.revoStore,
		TokenManager:     deps.tokenMgr,
		EmailService:     deps.email,
		FrontendBaseURL:  "http://frontend.local",
		RefreshTTL:       24 * time.Hour,
	})
}
