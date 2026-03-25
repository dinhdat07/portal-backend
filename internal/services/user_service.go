package services

import (
	"context"
	"errors"
	"time"

	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, email, username, password string) (*models.User, error) {
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil && existing.ID != uuid.Nil {
		return nil, errors.New("email already exists")
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hash)

	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: &hashStr,
		Role:         "user",
		Status:       "active",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	user.LastLoginAt = &now

	return user, nil
}
