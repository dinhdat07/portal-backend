package services

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
)

type userRepoMock struct {
	createFn             func(ctx context.Context, user *models.User) error
	findByIDFn           func(ctx context.Context, id uuid.UUID) (*models.User, error)
	findByIDUnscopedFn   func(ctx context.Context, id uuid.UUID) (*models.User, error)
	findByEmailFn        func(ctx context.Context, email string) (*models.User, error)
	findByUsernameFn     func(ctx context.Context, username string) (*models.User, error)
	updateFn             func(ctx context.Context, user *models.User) error
	updatePasswordFn     func(ctx context.Context, id uuid.UUID, passwordHash string) error
	updateRoleFn         func(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error
	markEmailVerifiedFn  func(ctx context.Context, id uuid.UUID) error
	deleteFn             func(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	restoreFn            func(ctx context.Context, id uuid.UUID) error
	listUsersFn          func(ctx context.Context, filter domain.UsersFilter) ([]models.User, int64, error)
	updatePasswordCalled int
	updateCalled         int
}

func (m *userRepoMock) Create(ctx context.Context, user *models.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *userRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *userRepoMock) FindByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.findByIDUnscopedFn != nil {
		return m.findByIDUnscopedFn(ctx, id)
	}
	return nil, nil
}

func (m *userRepoMock) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *userRepoMock) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.findByUsernameFn != nil {
		return m.findByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *userRepoMock) Update(ctx context.Context, user *models.User) error {
	m.updateCalled++
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}

func (m *userRepoMock) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	m.updatePasswordCalled++
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, id, passwordHash)
	}
	return nil
}

func (m *userRepoMock) UpdateRole(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, id, roleID)
	}
	return nil
}

func (m *userRepoMock) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	if m.markEmailVerifiedFn != nil {
		return m.markEmailVerifiedFn(ctx, id)
	}
	return nil
}

func (m *userRepoMock) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, deletedBy)
	}
	return nil
}

func (m *userRepoMock) Restore(ctx context.Context, id uuid.UUID) error {
	if m.restoreFn != nil {
		return m.restoreFn(ctx, id)
	}
	return nil
}

func (m *userRepoMock) ListUsers(ctx context.Context, filter domain.UsersFilter) ([]models.User, int64, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx, filter)
	}
	return nil, 0, nil
}

var _ repositories.UserRepository = (*userRepoMock)(nil)

type roleRepoMock struct {
	findByCodeFn func(ctx context.Context, code constants.RoleCode) (*models.Role, error)
}

func (m *roleRepoMock) FindByCode(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
	if m.findByCodeFn != nil {
		return m.findByCodeFn(ctx, code)
	}
	return nil, nil
}

func (m *roleRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return nil, nil
}

func (m *roleRepoMock) List(ctx context.Context) ([]models.Role, error) {
	return nil, nil
}

func (m *roleRepoMock) GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	return nil, nil
}

func (m *roleRepoMock) AssignPermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return nil
}

func (m *roleRepoMock) RemovePermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return nil
}

var _ repositories.RoleRepository = (*roleRepoMock)(nil)

type auditRepoMock struct {
	createFn      func(ctx context.Context, log *models.AuditLog) error
	listFn        func(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error)
	createdLogs   []*models.AuditLog
	createInvoked int
}

func (m *auditRepoMock) Create(ctx context.Context, log *models.AuditLog) error {
	m.createInvoked++
	m.createdLogs = append(m.createdLogs, log)
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}

func (m *auditRepoMock) List(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, 0, nil
}

var _ repositories.AuditLogRepository = (*auditRepoMock)(nil)

type txManagerMock struct {
	withTxFn        func(ctx context.Context, fn func(ctx context.Context) error) error
	withTxCallCount int
}

func (m *txManagerMock) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	m.withTxCallCount++
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

var _ repositories.TxManager = (*txManagerMock)(nil)

func cloneUser(u *models.User) *models.User {
	if u == nil {
		return nil
	}
	cpy := *u
	return &cpy
}

func ptrString(s string) *string {
	return &s
}

type tokenRepoMock struct {
	createFn            func(ctx context.Context, token *models.UserToken) error
	findValidTokenFn    func(ctx context.Context, tokenHash string, tokenType enum.TokenType) (*models.UserToken, error)
	markUsedFn          func(ctx context.Context, id uuid.UUID) error
	revokeFn            func(ctx context.Context, id uuid.UUID) error
	revokeByUserTypeFn  func(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error
	createCalled        int
	markUsedCalled      int
	revokeCalled        int
	revokeByTypeCalled  int
	findValidTokenCalls int
}

func (m *tokenRepoMock) Create(ctx context.Context, token *models.UserToken) error {
	m.createCalled++
	if m.createFn != nil {
		return m.createFn(ctx, token)
	}
	return nil
}

func (m *tokenRepoMock) FindValidToken(ctx context.Context, tokenHash string, tokenType enum.TokenType) (*models.UserToken, error) {
	m.findValidTokenCalls++
	if m.findValidTokenFn != nil {
		return m.findValidTokenFn(ctx, tokenHash, tokenType)
	}
	return nil, nil
}

func (m *tokenRepoMock) MarkUsed(ctx context.Context, id uuid.UUID) error {
	m.markUsedCalled++
	if m.markUsedFn != nil {
		return m.markUsedFn(ctx, id)
	}
	return nil
}

func (m *tokenRepoMock) Revoke(ctx context.Context, id uuid.UUID) error {
	m.revokeCalled++
	if m.revokeFn != nil {
		return m.revokeFn(ctx, id)
	}
	return nil
}

func (m *tokenRepoMock) RevokeByUserAndType(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error {
	m.revokeByTypeCalled++
	if m.revokeByUserTypeFn != nil {
		return m.revokeByUserTypeFn(ctx, userID, tokenType)
	}
	return nil
}

var _ repositories.UserTokenRepository = (*tokenRepoMock)(nil)

type sessionRepoMock struct {
	createFn                        func(ctx context.Context, session *models.AuthSession) error
	findActiveByRefreshTokenHashFn  func(ctx context.Context, hashToken string) (*models.AuthSession, error)
	findActiveByIDFn                func(ctx context.Context, id uuid.UUID) (*models.AuthSession, error)
	rotateRefreshTokenFn            func(ctx context.Context, in domain.RefreshInput) error
	revokeByIDFn                    func(ctx context.Context, sessionID uuid.UUID) error
	revokeAllByUserIDFn             func(ctx context.Context, userID uuid.UUID) error
	createCalled                    int
	findActiveByRefreshTokenCalled  int
	rotateRefreshTokenCalled        int
	revokeByIDCalled                int
	revokeAllByUserIDCalled         int
}

func (m *sessionRepoMock) Create(ctx context.Context, session *models.AuthSession) error {
	m.createCalled++
	if m.createFn != nil {
		return m.createFn(ctx, session)
	}
	return nil
}

func (m *sessionRepoMock) FindActiveByRefreshTokenHash(ctx context.Context, hashToken string) (*models.AuthSession, error) {
	m.findActiveByRefreshTokenCalled++
	if m.findActiveByRefreshTokenHashFn != nil {
		return m.findActiveByRefreshTokenHashFn(ctx, hashToken)
	}
	return nil, nil
}

func (m *sessionRepoMock) FindActiveByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
	if m.findActiveByIDFn != nil {
		return m.findActiveByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *sessionRepoMock) RotateRefreshToken(ctx context.Context, in domain.RefreshInput) error {
	m.rotateRefreshTokenCalled++
	if m.rotateRefreshTokenFn != nil {
		return m.rotateRefreshTokenFn(ctx, in)
	}
	return nil
}

func (m *sessionRepoMock) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	m.revokeByIDCalled++
	if m.revokeByIDFn != nil {
		return m.revokeByIDFn(ctx, sessionID)
	}
	return nil
}

func (m *sessionRepoMock) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	m.revokeAllByUserIDCalled++
	if m.revokeAllByUserIDFn != nil {
		return m.revokeAllByUserIDFn(ctx, userID)
	}
	return nil
}

var _ repositories.AuthSessionRepository = (*sessionRepoMock)(nil)

type emailSenderMock struct {
	sendVerificationFn func(ctx context.Context, to, name, verifyURL string) error
	sendResetFn        func(ctx context.Context, to, name, resetURL string) error
	sendSetFn          func(ctx context.Context, to, name, setPasswordURL string) error
	verificationCalled int
	resetCalled        int
	setCalled          int
}

func (m *emailSenderMock) SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error {
	m.verificationCalled++
	if m.sendVerificationFn != nil {
		return m.sendVerificationFn(ctx, to, name, verifyURL)
	}
	return nil
}

func (m *emailSenderMock) SendResetPasswordEmail(ctx context.Context, to, name, resetURL string) error {
	m.resetCalled++
	if m.sendResetFn != nil {
		return m.sendResetFn(ctx, to, name, resetURL)
	}
	return nil
}

func (m *emailSenderMock) SendSetPasswordEmail(ctx context.Context, to, name, setPasswordURL string) error {
	m.setCalled++
	if m.sendSetFn != nil {
		return m.sendSetFn(ctx, to, name, setPasswordURL)
	}
	return nil
}

var _ emailSender = (*emailSenderMock)(nil)

type tokenIssuerMock struct {
	generateAccessTokenFn func(userID uuid.UUID, sessionID uuid.UUID, roleID uuid.UUID, roleCode string, email string, username string) (string, error)
	generateRefreshFn     func() (string, error)
	expiresInSecondsFn    func() int
	accessCalled          int
	refreshCalled         int
}

func (m *tokenIssuerMock) GenerateAccessToken(userID uuid.UUID, sessionID uuid.UUID, roleID uuid.UUID, roleCode string, email string, username string) (string, error) {
	m.accessCalled++
	if m.generateAccessTokenFn != nil {
		return m.generateAccessTokenFn(userID, sessionID, roleID, roleCode, email, username)
	}
	return "access-token", nil
}

func (m *tokenIssuerMock) GenerateRefreshToken() (string, error) {
	m.refreshCalled++
	if m.generateRefreshFn != nil {
		return m.generateRefreshFn()
	}
	return "refresh-token", nil
}

func (m *tokenIssuerMock) ExpiresInSeconds() int {
	if m.expiresInSecondsFn != nil {
		return m.expiresInSecondsFn()
	}
	return 3600
}

var _ tokenIssuer = (*tokenIssuerMock)(nil)

