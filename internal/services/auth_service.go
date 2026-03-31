package services

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	"portal-system/internal/token"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db           *gorm.DB
	auditLogger  *AuditLogService
	userRepo     *repositories.UserRepository
	tokenManager *token.Manager
}

func NewAuthService(db *gorm.DB, userRepo *repositories.UserRepository, manager *token.Manager, logger *AuditLogService) *AuthService {
	return &AuthService{db: db, userRepo: userRepo, tokenManager: manager, auditLogger: logger}
}

func (s *AuthService) Register(ctx context.Context, meta *domain.AuditMeta, email, username, password, firstName, lastName string, dob time.Time) error {
	existing, _ := s.userRepo.FindByEmail(ctx, email)
	// later: check email in blacklist

	if existing != nil && existing.ID != uuid.Nil {
		return ErrEmailExists
	}

	existing, _ = s.userRepo.FindByUsername(ctx, username)
	if existing != nil && existing.ID != uuid.Nil {
		return ErrUsernameExists
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hash)

	// currently set auto active (no verify needed) for easy testing
	now := time.Now()
	user := &models.User{
		Email:           email,
		Username:        username,
		FirstName:       firstName,
		LastName:        lastName,
		DOB:             &dob,
		PasswordHash:    &hashStr,
		Role:            "user",
		Status:          "active",
		EmailVerifiedAt: &now,
	}

	// transaction, critical
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}
		return s.auditLogger.Log(ctx, meta, models.ActionRegister, nil, user)
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
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.DeletedAt.Valid {
		return nil, ErrAccountDeleted
	}

	if user.EmailVerifiedAt == nil {
		return nil, ErrAccountNotVerified
	}

	token, err := s.tokenManager.Generate(user.ID, user.Role, user.Email, user.Username)
	if err != nil {
		return nil, err
	}

	// best-effort
	s.auditLogger.Log(ctx, meta, models.ActionLogin, user, nil)

	return &domain.LoginResult{
		AccessToken: token,
		ExpiresIn:   s.tokenManager.ExpiresInSeconds(),
		User:        user,
	}, nil
}

func isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}
