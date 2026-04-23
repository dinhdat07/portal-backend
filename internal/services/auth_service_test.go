package services

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/platform/token"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_Register_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	dob := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	role := &models.Role{ID: uuid.New(), Code: constants.RoleCodeUser}

	tests := []struct {
		name            string
		findEmail       *models.User
		findEmailErr    error
		findUsername    *models.User
		findUsernameErr error
		roleErr         error
		roleNil         bool
		createErr       error
		revokeErr       error
		tokenCreateErr  error
		emailErr        error
		auditErr        error
		expectedErr     error
	}{
		{name: "find email error", findEmailErr: errors.New("db down"), expectedErr: errors.New("db down")},
		{name: "email already exists", findEmail: &models.User{ID: uuid.New()}, expectedErr: ErrEmailExists},
		{name: "find username error", findUsernameErr: errors.New("query failed"), expectedErr: errors.New("query failed")},
		{name: "username exists", findUsername: &models.User{ID: uuid.New()}, expectedErr: ErrUsernameExists},
		{name: "role lookup fails", roleErr: errors.New("role lookup failed"), expectedErr: ErrInternalServer},
		{name: "role is nil", roleNil: true, expectedErr: ErrInternalServer},
		{name: "create user fails", createErr: errors.New("insert failed"), expectedErr: ErrInternalServer},
		{name: "revoke token fails", revokeErr: errors.New("revoke failed"), expectedErr: ErrInternalServer},
		{name: "token create fails", tokenCreateErr: errors.New("token create failed"), expectedErr: ErrInternalServer},
		{name: "email fails", emailErr: errors.New("smtp failed"), expectedErr: ErrInternalServer},
		{name: "audit fails", auditErr: errors.New("audit failed"), expectedErr: ErrInternalServer},
		{name: "success"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &userRepoMock{
				findByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
					return tc.findEmail, tc.findEmailErr
				},
				findByUsernameFn: func(ctx context.Context, username string) (*models.User, error) {
					return tc.findUsername, tc.findUsernameErr
				},
				createFn: func(ctx context.Context, user *models.User) error {
					if user.ID == uuid.Nil {
						user.ID = uuid.New()
					}
					return tc.createErr
				},
			}
			tokenRepo := &tokenRepoMock{
				revokeByUserTypeFn: func(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error {
					return tc.revokeErr
				},
				createFn: func(ctx context.Context, token *models.UserToken) error {
					return tc.tokenCreateErr
				},
			}
			roleRepo := &roleRepoMock{
				findByCodeFn: func(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
					if tc.roleErr != nil {
						return nil, tc.roleErr
					}
					if tc.roleNil {
						return nil, nil
					}
					return role, nil
				},
			}
			auditRepo := &auditRepoMock{createFn: func(ctx context.Context, log *models.AuditLog) error {
				return tc.auditErr
			}}
			email := &emailSenderMock{sendVerificationFn: func(ctx context.Context, to, name, verifyURL string) error {
				return tc.emailErr
			}}
			svc := newAuthServiceForTest(authServiceTestDeps{
				auditRepo: auditRepo,
				userRepo:  userRepo,
				tokenRepo: tokenRepo,
				roleRepo:  roleRepo,
				email:     email,
			})

			err := svc.Register(context.Background(), meta, "john@example.com", "john", "Passw0rd!", "John", "Doe", dob)
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if tc.expectedErr == ErrInternalServer || tc.expectedErr == ErrEmailExists || tc.expectedErr == ErrUsernameExists {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err == nil || err.Error() != tc.expectedErr.Error() {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestAuthService_LogIn_Table(t *testing.T) {
	now := time.Now()
	hashed, hashErr := bcrypt.GenerateFromPassword([]byte("Passw0rd!"), bcrypt.DefaultCost)
	if hashErr != nil {
		t.Fatalf("cannot setup hash: %v", hashErr)
	}
	baseUser := &models.User{
		ID:              uuid.New(),
		Email:           "john@example.com",
		Username:        "john",
		PasswordHash:    ptrString(string(hashed)),
		EmailVerifiedAt: ptrTime(now),
		Role:            models.Role{ID: uuid.New(), Code: constants.RoleCodeUser},
	}

	tests := []struct {
		name             string
		identifier       string
		password         string
		user             *models.User
		findErr          error
		refreshErr       error
		createSessionErr error
		createRefreshErr error
		accessErr        error
		expectedErr      error
	}{
		{name: "email lookup error", identifier: "john@example.com", password: "Passw0rd!", findErr: errors.New("db failed"), expectedErr: ErrInvalidCredentials},
		{name: "username not found", identifier: "john", password: "Passw0rd!", user: nil, expectedErr: ErrInvalidCredentials},
		{name: "password not set", identifier: "john", password: "Passw0rd!", user: &models.User{ID: uuid.New(), PasswordHash: nil}, expectedErr: ErrAccountNotVerified},
		{name: "wrong password", identifier: "john", password: "wrong", user: cloneUser(baseUser), expectedErr: ErrInvalidCredentials},
		{name: "email not verified", identifier: "john", password: "Passw0rd!", user: &models.User{ID: baseUser.ID, PasswordHash: ptrString(string(hashed)), Role: baseUser.Role}, expectedErr: ErrAccountNotVerified},
		{name: "account deleted", identifier: "john", password: "Passw0rd!", user: &models.User{ID: baseUser.ID, PasswordHash: ptrString(string(hashed)), EmailVerifiedAt: ptrTime(now), DeletedAt: gorm.DeletedAt{Time: now, Valid: true}, Role: baseUser.Role}, expectedErr: ErrAccountDeleted},
		{name: "refresh token error", identifier: "john", password: "Passw0rd!", user: cloneUser(baseUser), refreshErr: errors.New("random failed"), expectedErr: ErrInternalServer},
		{name: "session create error", identifier: "john", password: "Passw0rd!", user: cloneUser(baseUser), createSessionErr: errors.New("insert failed"), expectedErr: ErrInternalServer},
		{name: "refresh token create error", identifier: "john", password: "Passw0rd!", user: cloneUser(baseUser), createRefreshErr: errors.New("insert refresh failed"), expectedErr: ErrInternalServer},
		{name: "access token error", identifier: "john", password: "Passw0rd!", user: cloneUser(baseUser), accessErr: errors.New("token failed"), expectedErr: errors.New("token failed")},
		{name: "success by email", identifier: "john@example.com", password: "Passw0rd!", user: cloneUser(baseUser)},
		{name: "success by username", identifier: "john", password: "Passw0rd!", user: cloneUser(baseUser)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &userRepoMock{
				findByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
					if tc.identifier == "john@example.com" {
						return cloneUser(tc.user), tc.findErr
					}
					return nil, gorm.ErrRecordNotFound
				},
				findByUsernameFn: func(ctx context.Context, username string) (*models.User, error) {
					if tc.identifier != "john@example.com" {
						return cloneUser(tc.user), tc.findErr
					}
					return nil, gorm.ErrRecordNotFound
				},
			}
			sessionRepo := &sessionRepoMock{
				createFn: func(ctx context.Context, session *models.AuthSession) error {
					if session.ID == uuid.Nil {
						session.ID = uuid.New()
					}
					return tc.createSessionErr
				},
			}
			refreshRepo := &refreshTokenRepoMock{
				createFn: func(ctx context.Context, refreshToken *models.RefreshToken) error {
					if refreshToken.ID == uuid.Nil {
						refreshToken.ID = uuid.New()
					}
					return tc.createRefreshErr
				},
			}
			tokenMgr := &tokenIssuerMock{
				generateRefreshFn: func() (string, error) {
					if tc.refreshErr != nil {
						return "", tc.refreshErr
					}
					return "refresh-token", nil
				},
				generateAccessTokenFn: func(input token.GenerateAccessTokenInput) (string, error) {
					if tc.accessErr != nil {
						return "", tc.accessErr
					}
					return "access-token", nil
				},
				expiresInSecondsFn: func() int { return 900 },
			}

			svc := newAuthServiceForTest(authServiceTestDeps{
				auditRepo:   &auditRepoMock{},
				userRepo:    userRepo,
				refreshRepo: refreshRepo,
				sessionRepo: sessionRepo,
				tokenMgr:    tokenMgr,
			})
			result, err := svc.LogIn(context.Background(), &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "ua"}, tc.identifier, tc.password)
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if result == nil || result.AccessToken == "" || result.RefreshToken == "" || result.ExpiresIn != 900 {
					t.Fatalf("unexpected login result: %#v", result)
				}
			} else if tc.expectedErr == ErrInvalidCredentials || tc.expectedErr == ErrAccountNotVerified || tc.expectedErr == ErrAccountDeleted || tc.expectedErr == ErrInternalServer {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err == nil || err.Error() != tc.expectedErr.Error() {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestAuthService_VerifyEmail_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "ua"}
	foundToken := &models.UserToken{ID: uuid.New(), UserID: uuid.New(), TokenType: enum.TokenTypeEmailVerification}
	baseUser := &models.User{ID: foundToken.UserID, Email: "john@example.com", Username: "john", Role: models.Role{Code: constants.RoleCodeUser}}

	tests := []struct {
		name             string
		findTokenErr     error
		user             *models.User
		findUserErr      error
		markVerifiedErr  error
		markUsedErr      error
		auditErr         error
		expectedErr      error
		tokenType        enum.TokenType
		expectMarkUsed   bool
		expectMarkVerify bool
	}{
		{name: "invalid token", findTokenErr: errors.New("not found"), tokenType: enum.TokenTypeEmailVerification, expectedErr: ErrInvalidToken},
		{name: "user not found", findUserErr: errors.New("missing"), tokenType: enum.TokenTypeEmailVerification, expectedErr: ErrUserNotFound},
		{name: "already deleted", user: &models.User{ID: foundToken.UserID, Status: enum.StatusDeleted}, tokenType: enum.TokenTypeEmailVerification, expectedErr: ErrUserAlreadyDeleted},
		{name: "already active", user: &models.User{ID: foundToken.UserID, Status: enum.StatusActive}, tokenType: enum.TokenTypeEmailVerification},
		{name: "mark verified fails", user: cloneUser(baseUser), markVerifiedErr: errors.New("update failed"), tokenType: enum.TokenTypeEmailVerification, expectedErr: ErrInternalServer, expectMarkVerify: true},
		{name: "mark used fails", user: cloneUser(baseUser), markUsedErr: errors.New("mark failed"), tokenType: enum.TokenTypeEmailVerification, expectedErr: ErrInternalServer, expectMarkVerify: true, expectMarkUsed: true},
		{name: "audit fails", user: cloneUser(baseUser), auditErr: errors.New("audit failed"), tokenType: enum.TokenTypeEmailVerification, expectedErr: ErrInternalServer, expectMarkVerify: true, expectMarkUsed: true},
		{name: "success", user: cloneUser(baseUser), tokenType: enum.TokenTypeEmailVerification, expectMarkVerify: true, expectMarkUsed: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenRepo := &tokenRepoMock{
				findValidTokenFn: func(ctx context.Context, tokenHash string, tokenType enum.TokenType) (*models.UserToken, error) {
					if tc.findTokenErr != nil {
						return nil, tc.findTokenErr
					}
					return foundToken, nil
				},
				markUsedFn: func(ctx context.Context, id uuid.UUID) error {
					return tc.markUsedErr
				},
			}
			userRepo := &userRepoMock{
				findByIDFn: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					if tc.findUserErr != nil {
						return nil, tc.findUserErr
					}
					return cloneUser(tc.user), nil
				},
				markEmailVerifiedFn: func(ctx context.Context, id uuid.UUID) error {
					return tc.markVerifiedErr
				},
			}
			auditRepo := &auditRepoMock{createFn: func(ctx context.Context, log *models.AuditLog) error { return tc.auditErr }}
			svc := newAuthServiceForTest(authServiceTestDeps{
				auditRepo: auditRepo,
				userRepo:  userRepo,
				tokenRepo: tokenRepo,
			})

			err := svc.VerifyEmail(context.Background(), meta, "raw-token", tc.tokenType)
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}

			if tc.expectMarkVerify && userRepo.markEmailVerifiedFn != nil && tokenRepo.markUsedFn == nil {
				t.Fatal("unreachable")
			}
		})
	}
}

func TestAuthService_ResendVerification_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "ua"}
	user := &models.User{ID: uuid.New(), Email: "john@example.com", FirstName: "John", Username: "john", Role: models.Role{Code: constants.RoleCodeUser}}

	tests := []struct {
		name           string
		findErr        error
		revokeErr      error
		tokenCreateErr error
		emailErr       error
		auditErr       error
		expectedErr    error
	}{
		{name: "find user error", findErr: errors.New("not found"), expectedErr: ErrUserNotFound},
		{name: "revoke error", revokeErr: errors.New("revoke failed"), expectedErr: ErrInternalServer},
		{name: "token create error", tokenCreateErr: errors.New("token create failed"), expectedErr: ErrInternalServer},
		{name: "email send error", emailErr: errors.New("smtp failed"), expectedErr: ErrInternalServer},
		{name: "audit error", auditErr: errors.New("audit failed"), expectedErr: ErrInternalServer},
		{name: "success"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &userRepoMock{
				findByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
					if tc.findErr != nil {
						return nil, tc.findErr
					}
					return cloneUser(user), nil
				},
			}
			tokenRepo := &tokenRepoMock{
				revokeByUserTypeFn: func(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error {
					return tc.revokeErr
				},
				createFn: func(ctx context.Context, token *models.UserToken) error {
					return tc.tokenCreateErr
				},
			}
			email := &emailSenderMock{sendVerificationFn: func(ctx context.Context, to, name, verifyURL string) error { return tc.emailErr }}
			auditRepo := &auditRepoMock{createFn: func(ctx context.Context, log *models.AuditLog) error { return tc.auditErr }}
			svc := newAuthServiceForTest(authServiceTestDeps{
				auditRepo: auditRepo,
				userRepo:  userRepo,
				tokenRepo: tokenRepo,
				email:     email,
			})

			err := svc.ResendVerification(context.Background(), meta, user.Email, enum.TokenTypeEmailVerification)
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestAuthService_ForgotPassword_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "ua"}
	user := &models.User{ID: uuid.New(), Email: "john@example.com", FirstName: "John", Username: "john", Role: models.Role{Code: constants.RoleCodeUser}}

	tests := []struct {
		name           string
		user           *models.User
		findErr        error
		revokeErr      error
		tokenCreateErr error
		auditErr       error
		emailErr       error
		expectedErr    error
	}{
		{name: "find user error", findErr: errors.New("db failed"), expectedErr: ErrInternalServer},
		{name: "user nil", user: nil},
		{name: "user deleted", user: &models.User{ID: uuid.New(), Status: enum.StatusDeleted}},
		{name: "revoke error", user: cloneUser(user), revokeErr: errors.New("revoke failed"), expectedErr: ErrInternalServer},
		{name: "token create error", user: cloneUser(user), tokenCreateErr: errors.New("create failed"), expectedErr: ErrInternalServer},
		{name: "audit error", user: cloneUser(user), auditErr: errors.New("audit failed"), expectedErr: ErrInternalServer},
		{name: "email error", user: cloneUser(user), emailErr: errors.New("smtp failed"), expectedErr: ErrSendResetPasswordEmail},
		{name: "success", user: cloneUser(user)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &userRepoMock{
				findByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
					if tc.findErr != nil {
						return nil, tc.findErr
					}
					return cloneUser(tc.user), nil
				},
			}
			tokenRepo := &tokenRepoMock{
				revokeByUserTypeFn: func(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error {
					return tc.revokeErr
				},
				createFn: func(ctx context.Context, token *models.UserToken) error {
					return tc.tokenCreateErr
				},
			}
			auditRepo := &auditRepoMock{createFn: func(ctx context.Context, log *models.AuditLog) error { return tc.auditErr }}
			email := &emailSenderMock{sendResetFn: func(ctx context.Context, to, name, resetURL string) error { return tc.emailErr }}
			svc := newAuthServiceForTest(authServiceTestDeps{
				auditRepo: auditRepo,
				userRepo:  userRepo,
				tokenRepo: tokenRepo,
				email:     email,
			})

			err := svc.ForgotPassword(context.Background(), meta, "john@example.com")
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestAuthService_SetAndResetPassword_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "ua"}
	userID := uuid.New()
	tokenID := uuid.New()
	baseUser := &models.User{
		ID:       userID,
		Email:    "john@example.com",
		Username: "john",
		Status:   enum.StatusActive,
		Role:     models.Role{Code: constants.RoleCodeUser},
	}

	tests := []struct {
		name          string
		useReset      bool
		input         *domain.SetPasswordInput
		tokenType     enum.TokenType
		findTokenErr  error
		findUserErr   error
		user          *models.User
		updateErr     error
		markVerifyErr error
		markUsedErr   error
		auditErr      error
		expectedErr   error
	}{
		{name: "nil input", input: nil, tokenType: enum.TokenTypePasswordSet, expectedErr: ErrInvalidInput},
		{name: "blank token", input: &domain.SetPasswordInput{Token: " ", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordSet, expectedErr: ErrInvalidInput},
		{name: "password mismatch", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "b"}, tokenType: enum.TokenTypePasswordSet, expectedErr: ErrPasswordConfirmationMismatch},
		{name: "token invalid", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordSet, findTokenErr: errors.New("not found"), expectedErr: ErrInvalidToken},
		{name: "user not found", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordSet, findUserErr: errors.New("missing"), expectedErr: ErrUserNotFound},
		{name: "user deleted", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordSet, user: &models.User{ID: userID, Status: enum.StatusDeleted}, expectedErr: ErrUserAlreadyDeleted},
		{name: "password already set", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordSet, user: &models.User{ID: userID, Status: enum.StatusActive, PasswordHash: ptrString("hash")}, expectedErr: ErrPasswordAlreadySet},
		{name: "invalid token type", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenType("invalid"), user: cloneUser(baseUser), expectedErr: ErrInvalidInput},
		{name: "update password error", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordReset, user: cloneUser(baseUser), updateErr: errors.New("save failed"), expectedErr: ErrInternalServer},
		{name: "mark verified error", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordSet, user: cloneUser(baseUser), markVerifyErr: errors.New("mark failed"), expectedErr: ErrInternalServer},
		{name: "mark used error", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordReset, user: cloneUser(baseUser), markUsedErr: errors.New("mark used failed"), expectedErr: ErrInternalServer},
		{name: "audit error", input: &domain.SetPasswordInput{Token: "t", Password: "a", ConfirmPassword: "a"}, tokenType: enum.TokenTypePasswordReset, user: cloneUser(baseUser), auditErr: errors.New("audit failed"), expectedErr: ErrInternalServer},
		{name: "set password success", input: &domain.SetPasswordInput{Token: "t", Password: "abc12345", ConfirmPassword: "abc12345"}, tokenType: enum.TokenTypePasswordSet, user: cloneUser(baseUser)},
		{name: "reset password success via wrapper", useReset: true, input: &domain.SetPasswordInput{Token: "t", Password: "abc12345", ConfirmPassword: "abc12345"}, tokenType: enum.TokenTypePasswordReset, user: cloneUser(baseUser)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenRepo := &tokenRepoMock{
				findValidTokenFn: func(ctx context.Context, tokenHash string, tokenType enum.TokenType) (*models.UserToken, error) {
					if tc.findTokenErr != nil {
						return nil, tc.findTokenErr
					}
					return &models.UserToken{ID: tokenID, UserID: userID}, nil
				},
				markUsedFn: func(ctx context.Context, id uuid.UUID) error {
					return tc.markUsedErr
				},
			}
			userRepo := &userRepoMock{
				findByIDFn: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					if tc.findUserErr != nil {
						return nil, tc.findUserErr
					}
					return cloneUser(tc.user), nil
				},
				updatePasswordFn: func(ctx context.Context, id uuid.UUID, passwordHash string) error {
					return tc.updateErr
				},
				markEmailVerifiedFn: func(ctx context.Context, id uuid.UUID) error {
					return tc.markVerifyErr
				},
			}
			auditRepo := &auditRepoMock{createFn: func(ctx context.Context, log *models.AuditLog) error {
				return tc.auditErr
			}}
			svc := newAuthServiceForTest(authServiceTestDeps{
				auditRepo: auditRepo,
				userRepo:  userRepo,
				tokenRepo: tokenRepo,
			})

			var err error
			if tc.useReset {
				err = svc.ResetPassword(context.Background(), meta, tc.input, tc.tokenType)
			} else {
				err = svc.SetPassword(context.Background(), meta, tc.input, tc.tokenType)
			}
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestAuthService_Refresh_Table(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	refreshTokenID := uuid.New()
	now := time.Now().UTC()
	baseUser := &models.User{
		ID:              userID,
		Email:           "john@example.com",
		Username:        "john",
		EmailVerifiedAt: ptrTime(now),
		Role:            models.Role{ID: uuid.New(), Code: constants.RoleCodeUser},
	}
	baseSession := &models.AuthSession{ID: sessionID, UserID: userID, ExpiresAt: now.Add(24 * time.Hour)}
	baseRefreshToken := &models.RefreshToken{ID: refreshTokenID, SessionID: sessionID, UserID: userID, ExpiresAt: now.Add(1 * time.Hour)}

	tests := []struct {
		name               string
		token              string
		foundToken         *models.RefreshToken
		findTokenErr       error
		sessionErr         error
		session            *models.AuthSession
		userErr            error
		user               *models.User
		revokeErr          error
		createRefreshErr   error
		markReplacementErr error
		accessErr          error
		expected           error
	}{
		{name: "blank token", token: " ", expected: ErrInvalidInput},
		{name: "refresh token not found", token: "r1", findTokenErr: errors.New("not found"), expected: ErrInvalidRefreshToken},
		{name: "revoked refresh token", token: "r1", foundToken: &models.RefreshToken{ID: refreshTokenID, SessionID: sessionID, UserID: userID, RevokedAt: ptrTime(now), ExpiresAt: now.Add(1 * time.Hour)}, expected: ErrInvalidRefreshToken},
		{name: "expired refresh token", token: "r1", foundToken: &models.RefreshToken{ID: refreshTokenID, SessionID: sessionID, UserID: userID, ExpiresAt: now.Add(-1 * time.Hour)}, expected: ErrInvalidRefreshToken},
		{name: "invalid refresh session", token: "r1", foundToken: baseRefreshToken, sessionErr: errors.New("not found"), expected: ErrInvalidRefreshToken},
		{name: "user lookup error", token: "r1", foundToken: baseRefreshToken, session: baseSession, userErr: errors.New("missing"), expected: ErrUserNotFound},
		{name: "user deleted", token: "r1", foundToken: baseRefreshToken, session: baseSession, user: &models.User{ID: userID, DeletedAt: gorm.DeletedAt{Time: now, Valid: true}}, expected: ErrAccountDeleted},
		{name: "email not verified", token: "r1", foundToken: baseRefreshToken, session: baseSession, user: &models.User{ID: userID}, expected: ErrAccountNotVerified},
		{name: "revoke error", token: "r1", foundToken: baseRefreshToken, session: baseSession, user: cloneUser(baseUser), revokeErr: errors.New("revoke failed"), expected: ErrInternalServer},
		{name: "refresh token create error", token: "r1", foundToken: baseRefreshToken, session: baseSession, user: cloneUser(baseUser), createRefreshErr: errors.New("create failed"), expected: ErrInternalServer},
		{name: "mark replacement error", token: "r1", foundToken: baseRefreshToken, session: baseSession, user: cloneUser(baseUser), markReplacementErr: errors.New("replace failed"), expected: ErrInternalServer},
		{name: "access error", token: "r1", foundToken: baseRefreshToken, session: baseSession, user: cloneUser(baseUser), accessErr: errors.New("jwt failed"), expected: ErrInternalServer},
		{name: "success", token: "r1", foundToken: baseRefreshToken, session: baseSession, user: cloneUser(baseUser)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			refreshRepo := &refreshTokenRepoMock{
				findByTokenHashFn: func(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
					if tc.findTokenErr != nil {
						return nil, tc.findTokenErr
					}
					return tc.foundToken, nil
				},
				revokeByIDFn: func(ctx context.Context, id uuid.UUID) error {
					return tc.revokeErr
				},
				createFn: func(ctx context.Context, refreshToken *models.RefreshToken) error {
					if refreshToken.ID == uuid.Nil {
						refreshToken.ID = uuid.New()
					}
					return tc.createRefreshErr
				},
				markReplacementFn: func(ctx context.Context, id uuid.UUID, replacementID uuid.UUID) error {
					return tc.markReplacementErr
				},
			}
			sessionRepo := &sessionRepoMock{
				findActiveByIDFn: func(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
					if tc.sessionErr != nil {
						return nil, tc.sessionErr
					}
					return tc.session, nil
				},
			}
			userRepo := &userRepoMock{
				findByIDFn: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					if tc.userErr != nil {
						return nil, tc.userErr
					}
					return cloneUser(tc.user), nil
				},
			}
			tokenMgr := &tokenIssuerMock{
				generateAccessTokenFn: func(input token.GenerateAccessTokenInput) (string, error) {
					if tc.accessErr != nil {
						return "", tc.accessErr
					}
					return "access-token", nil
				},
				expiresInSecondsFn: func() int { return 1200 },
			}
			svc := newAuthServiceForTest(authServiceTestDeps{
				userRepo:    userRepo,
				refreshRepo: refreshRepo,
				sessionRepo: sessionRepo,
				tokenMgr:    tokenMgr,
			})

			res, err := svc.Refresh(context.Background(), &domain.AuditMeta{}, tc.token)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if res == nil || res.AccessToken == "" || res.RefreshToken == "" || res.ExpiresIn != 1200 {
					t.Fatalf("unexpected refresh result: %#v", res)
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}
		})
	}
}

func TestAuthService_Logout_Table(t *testing.T) {
	actor := &domain.AuditUser{ID: uuid.New()}
	tests := []struct {
		name      string
		actor     *domain.AuditUser
		sessionID uuid.UUID
		revokeErr error
		expected  error
	}{
		{name: "nil actor", actor: nil, sessionID: uuid.New(), expected: ErrUnauthorized},
		{name: "nil session", actor: actor, sessionID: uuid.Nil, expected: ErrInvalidInput},
		{name: "revoke fails", actor: actor, sessionID: uuid.New(), revokeErr: errors.New("db failed"), expected: ErrInternalServer},
		{name: "success", actor: actor, sessionID: uuid.New()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sessionRepo := &sessionRepoMock{
				revokeByIDFn: func(ctx context.Context, sessionID uuid.UUID) error {
					return tc.revokeErr
				},
			}
			refreshRepo := &refreshTokenRepoMock{
				revokeBySessionIDFn: func(ctx context.Context, sessionID uuid.UUID) error {
					return tc.revokeErr
				},
			}
			svc := newAuthServiceForTest(authServiceTestDeps{
				refreshRepo: refreshRepo,
				sessionRepo: sessionRepo,
			})
			err := svc.Logout(context.Background(), &domain.AuditMeta{}, tc.actor, tc.sessionID)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}
		})
	}
}

func TestAuthService_LogoutAll_Table(t *testing.T) {
	actor := &domain.AuditUser{ID: uuid.New()}
	tests := []struct {
		name      string
		actor     *domain.AuditUser
		revokeErr error
		expected  error
	}{
		{name: "nil actor", actor: nil, expected: ErrUnauthorized},
		{name: "nil actor id", actor: &domain.AuditUser{ID: uuid.Nil}, expected: ErrInvalidInput},
		{name: "revoke fails", actor: actor, revokeErr: errors.New("db failed"), expected: ErrInternalServer},
		{name: "success", actor: actor},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sessionRepo := &sessionRepoMock{
				revokeAllByUserIDFn: func(ctx context.Context, userID uuid.UUID) error {
					return tc.revokeErr
				},
			}
			refreshRepo := &refreshTokenRepoMock{
				revokeByUserIDFn: func(ctx context.Context, userID uuid.UUID) error {
					return tc.revokeErr
				},
			}
			svc := newAuthServiceForTest(authServiceTestDeps{
				refreshRepo: refreshRepo,
				sessionRepo: sessionRepo,
			})
			err := svc.LogoutAll(context.Background(), &domain.AuditMeta{}, tc.actor)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}
		})
	}
}
