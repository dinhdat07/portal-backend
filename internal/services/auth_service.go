package services

import (
	"context"
	"strings"
	"time"

	"net/mail"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	"portal-system/internal/token"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     *repositories.UserRepository
	tokenManager *token.Manager
}

func NewAuthService(userRepo *repositories.UserRepository, manager *token.Manager) *AuthService {
	return &AuthService{userRepo: userRepo, tokenManager: manager}
}

func (s *AuthService) Register(ctx context.Context, email, username, password, firstName, lastName string, dob time.Time) error {
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

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) LogIn(ctx context.Context, identifier, password string) (*LoginResult, error) {
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

	token, err := s.tokenManager.Generate(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken: token,
		ExpiresIn:   s.tokenManager.ExpiresInSeconds(),
		User:        user,
	}, nil
}

type LoginResult struct {
	AccessToken string
	ExpiresIn   int
	User        *models.User
}

func isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}
