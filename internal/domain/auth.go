package domain

import (
	"portal-system/internal/models"
	"time"

	"github.com/google/uuid"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         *models.User
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type SetPasswordInput struct {
	Token           string
	Password        string
	ConfirmPassword string
}

type RefreshInput struct {
	SessionID    uuid.UUID
	NewTokenHash string
	NewExpiresAt time.Time
	RotatedAt    time.Time
}
