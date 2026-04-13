package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/platform/email"
	"portal-system/internal/platform/token"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	txManager       repositories.TxManager
	auditLogger     *AuditLogService
	userRepo        repositories.UserRepository
	tokenRepo       repositories.UserTokenRepository
	roleRepo        repositories.RoleRepository
	sessionRepo     repositories.AuthSessionRepository
	tokenManager    *token.Manager
	emailService    *email.SMTPEmailService
	frontendBaseURL string
	refreshTTL      time.Duration
}

func NewAuthService(txManager repositories.TxManager,
	userRepo repositories.UserRepository,
	tokenRepo repositories.UserTokenRepository,
	roleRepo repositories.RoleRepository,
	sessionRepo repositories.AuthSessionRepository,
	manager *token.Manager,
	logger *AuditLogService,
	emailService *email.SMTPEmailService,
	frontendUrl string,
	refreshSec int) *AuthService {

	refreshTTL := time.Duration(refreshSec) * time.Second

	return &AuthService{
		txManager:       txManager,
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		roleRepo:        roleRepo,
		sessionRepo:     sessionRepo,
		tokenManager:    manager,
		auditLogger:     logger,
		emailService:    emailService,
		frontendBaseURL: frontendUrl,
		refreshTTL:      refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, meta *domain.AuditMeta, email, username, password, firstName, lastName string, dob time.Time) error {
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing != nil && existing.ID != uuid.Nil {
		return ErrEmailExists
	}

	existing, err = s.userRepo.FindByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing != nil && existing.ID != uuid.Nil {
		return ErrUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalServer
	}
	hashStr := string(hash)

	tokenHash, rawToken, err := generateHashToken()
	if err != nil {
		return err
	}

	role, err := s.roleRepo.FindByCode(ctx, constants.RoleCodeUser)
	if role == nil || err != nil {
		return ErrInternalServer
	}

	user := &models.User{
		Email:        email,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		DOB:          &dob,
		PasswordHash: &hashStr,
		RoleID:       role.ID,
		Role:         *role,
		Status:       enum.StatusPending,
	}

	// transaction, critical
	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}

		if err := s.tokenRepo.
			RevokeByUserAndType(ctx, user.ID, enum.TokenTypeEmailVerification); err != nil {
			return err
		}

		verifyToken := &models.UserToken{
			UserID:    user.ID,
			TokenType: enum.TokenTypeEmailVerification,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}

		if err := s.tokenRepo.Create(ctx, verifyToken); err != nil {
			return err
		}

		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.frontendBaseURL, url.QueryEscape(rawToken))

		if err := s.emailService.SendVerificationEmail(ctx, email, firstName, verifyURL); err != nil {
			return ErrSendVerificationEmail
		}

		target := domain.MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, enum.ActionRegister, nil, target)
	})

	if err != nil {
		return ErrInternalServer
	}

	return nil
}

func (s *AuthService) LogIn(ctx context.Context, meta *domain.AuditMeta, identifier, password string) (*domain.LoginResult, error) {
	var user *models.User
	var err error

	identifier = strings.TrimSpace(strings.ToLower(identifier))

	if isEmail(identifier) {
		user, err = s.userRepo.FindByEmail(ctx, identifier)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, identifier)
	}
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrAccountNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.EmailVerifiedAt == nil {
		return nil, ErrAccountNotVerified
	}

	if user.DeletedAt.Valid {
		return nil, ErrAccountDeleted
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, ErrInternalServer
	}

	now := time.Now()
	refreshTokenHash := token.HashToken(refreshToken)
	refreshExpiresAt := now.Add(s.refreshTTL)
	session := &models.AuthSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        refreshExpiresAt,
		LastUsedAt:       &now,
	}

	if meta != nil {
		session.UserAgent = meta.UserAgent
		session.IPAddress = meta.IPAddress
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, ErrInternalServer
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, session.ID, user.Role.ID, string(user.Role.Code), user.Email, user.Username)
	if err != nil {
		return nil, err
	}

	// best-effort
	actor := domain.MapUserToAuditUser(user)
	s.auditLogger.Log(ctx, meta, enum.ActionLogin, actor, nil)

	return &domain.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.tokenManager.ExpiresInSeconds(),
		User:         user,
	}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, meta *domain.AuditMeta, rawToken string, tokenType enum.TokenType) error {
	tokenHash := token.HashToken(rawToken)

	found, err := s.tokenRepo.FindValidToken(ctx, tokenHash, tokenType)
	if err != nil {
		return ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, found.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Status == enum.StatusDeleted {
		return ErrUserAlreadyDeleted
	}

	if user.Status == enum.StatusActive {
		return nil
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
			return err
		}

		if err := s.tokenRepo.MarkUsed(ctx, found.ID); err != nil {
			return err
		}

		actor := domain.MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, enum.ActionVerifyEmail, actor, actor)
	})
	if err != nil {
		return ErrInternalServer
	}

	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, meta *domain.AuditMeta, email string, tokenType enum.TokenType) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	// revoke old token
	if err = s.tokenRepo.RevokeByUserAndType(ctx, user.ID, tokenType); err != nil {
		return ErrInternalServer
	}

	// generate new token
	tokenHash, rawToken, err := generateHashToken()
	if err != nil {
		return err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		verifyToken := &models.UserToken{
			UserID:    user.ID,
			TokenType: enum.TokenTypeEmailVerification,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}

		if err := s.tokenRepo.Create(ctx, verifyToken); err != nil {
			return err
		}

		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.frontendBaseURL, url.QueryEscape(rawToken))

		if err := s.emailService.SendVerificationEmail(ctx, email, user.FirstName, verifyURL); err != nil {
			return ErrSendVerificationEmail
		}

		actor := domain.MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, enum.ActionResendVerification, actor, actor)
	})

	if err != nil {
		return ErrInternalServer
	}

	return nil

}

func (s *AuthService) ForgotPassword(ctx context.Context, meta *domain.AuditMeta, email string) error {

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrInternalServer
	}

	if user == nil {
		return nil
	}

	if user.Status == enum.StatusDeleted {
		return nil
	}

	// generate token
	tokenHash, rawToken, err := generateHashToken()
	if err != nil {
		return err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {

		// revoke old token
		if err := s.tokenRepo.
			RevokeByUserAndType(ctx, user.ID, enum.TokenTypePasswordReset); err != nil {
			return err
		}

		// create token
		resetToken := &models.UserToken{
			UserID:    user.ID,
			TokenType: enum.TokenTypePasswordReset,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		}

		if err := s.tokenRepo.Create(ctx, resetToken); err != nil {
			return err
		}
		actor := domain.MapUserToAuditUser(user)
		return s.auditLogger.
			Log(ctx, meta, enum.ActionForgotPassword, actor, actor)
	})

	if err != nil {
		return ErrInternalServer
	}

	// build link
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.frontendBaseURL, url.QueryEscape(rawToken))

	// send mail
	if err := s.emailService.SendResetPasswordEmail(ctx, user.Email, user.FirstName, resetURL); err != nil {
		return ErrSendResetPasswordEmail
	}

	return nil
}

func (s *AuthService) SetPassword(ctx context.Context, meta *domain.AuditMeta, in *domain.SetPasswordInput, tokenType enum.TokenType) error {
	return s.applyPasswordByToken(ctx, meta, in, tokenType)
}

func (s *AuthService) ResetPassword(ctx context.Context, meta *domain.AuditMeta, in *domain.SetPasswordInput, tokenType enum.TokenType) error {
	return s.applyPasswordByToken(ctx, meta, in, tokenType)
}

func (s *AuthService) applyPasswordByToken(ctx context.Context, meta *domain.AuditMeta, in *domain.SetPasswordInput, tokenType enum.TokenType) error {
	if in == nil {
		return ErrInvalidInput
	}

	if strings.TrimSpace(in.Token) == "" ||
		strings.TrimSpace(in.Password) == "" ||
		strings.TrimSpace(in.ConfirmPassword) == "" {
		return ErrInvalidInput
	}

	if in.Password != in.ConfirmPassword {
		return ErrPasswordConfirmationMismatch
	}

	tokenHash := token.HashToken(in.Token)

	found, err := s.tokenRepo.FindValidToken(ctx, tokenHash, tokenType)
	if err != nil {
		return ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, found.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Status == enum.StatusDeleted {
		return ErrUserAlreadyDeleted
	}

	if tokenType == enum.TokenTypePasswordSet {
		if user.PasswordHash != nil && *user.PasswordHash != "" {
			return ErrPasswordAlreadySet
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalServer
	}
	hashStr := string(hash)

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		switch tokenType {
		case enum.TokenTypePasswordSet:
			if err := s.userRepo.UpdatePassword(ctx, user.ID, hashStr); err != nil {
				return err
			}

			if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
				return err
			}

		case enum.TokenTypePasswordReset:
			if err := s.userRepo.UpdatePassword(ctx, user.ID, hashStr); err != nil {
				return err
			}

		default:
			return ErrInvalidInput
		}

		if err := s.tokenRepo.MarkUsed(ctx, found.ID); err != nil {
			return err
		}

		action := enum.ActionResetPassword
		if tokenType == enum.TokenTypePasswordSet {
			action = enum.ActionSetPassword
		}

		actor := domain.MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, action, actor, actor)
	})
	if err != nil {
		if errors.Is(err, ErrPasswordAlreadySet) || errors.Is(err, ErrInvalidInput) {
			return err
		}
		return ErrInternalServer
	}

	return nil
}

func (s *AuthService) Refresh(ctx context.Context, meta *domain.AuditMeta, refreshToken string) (*domain.RefreshResult, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidInput
	}
	refreshTokenHash := token.HashToken(refreshToken)
	session, err := s.sessionRepo.FindActiveByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if user.DeletedAt.Valid {
		return nil, ErrAccountDeleted
	}

	if user.EmailVerifiedAt == nil {
		return nil, ErrAccountNotVerified
	}

	now := time.Now()

	newRefreshTokenHash, newRefreshToken, err := generateHashToken()
	if err != nil {
		return nil, err
	}
	newRefreshExpiresAt := now.Add(s.refreshTTL)

	err = s.sessionRepo.RotateRefreshToken(ctx, domain.RefreshInput{
		SessionID:    session.ID,
		NewTokenHash: newRefreshTokenHash,
		NewExpiresAt: newRefreshExpiresAt,
		RotatedAt:    now,
	})
	if err != nil {
		return nil, ErrInternalServer
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(
		user.ID,
		session.ID,
		user.Role.ID,
		string(user.Role.Code),
		user.Email,
		user.Username,
	)
	if err != nil {
		return nil, ErrInternalServer
	}

	return &domain.RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.tokenManager.ExpiresInSeconds(),
	}, nil

}

func (s *AuthService) Logout(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, sessionID uuid.UUID) error {
	if actor == nil {
		return ErrUnauthorized
	}

	if sessionID == uuid.Nil {
		return ErrInvalidInput
	}

	if err := s.sessionRepo.RevokeByID(ctx, sessionID); err != nil {
		return ErrInternalServer
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser) error {
	if actor == nil {
		return ErrUnauthorized
	}

	if actor.ID == uuid.Nil {
		return ErrInvalidInput
	}

	if err := s.sessionRepo.RevokeAllByUserID(ctx, actor.ID); err != nil {
		return ErrInternalServer
	}

	return nil
}

func isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func generateHashToken() (string, string, error) {
	rawToken, err := token.GenerateSecureToken(32)
	if err != nil {
		return "", "", ErrInternalServer
	}
	tokenHash := token.HashToken(rawToken)
	return tokenHash, rawToken, nil
}
