package services

import (
	"time"
)

func ptrTime(v time.Time) *time.Time {
	return &v
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

func newAuthServiceForTest(tx *txManagerMock, auditRepo *auditRepoMock, userRepo *userRepoMock, tokenRepo *tokenRepoMock, roleRepo *roleRepoMock, sessionRepo *sessionRepoMock, tokenMgr *tokenIssuerMock, email *emailSenderMock) *AuthService {
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
	if sessionRepo == nil {
		sessionRepo = &sessionRepoMock{}
	}
	if tokenMgr == nil {
		tokenMgr = &tokenIssuerMock{}
	}
	if email == nil {
		email = &emailSenderMock{}
	}

	return NewAuthService(AuthServiceDeps{
		TxManager:       tx,
		AuditLogger:     NewAuditLogService(auditRepo),
		UserRepo:        userRepo,
		TokenRepo:       tokenRepo,
		RoleRepo:        roleRepo,
		SessionRepo:     sessionRepo,
		TokenManager:    tokenMgr,
		EmailService:    email,
		FrontendBaseURL: "http://frontend.local",
		RefreshTTL:      24 * time.Hour,
	})
}
