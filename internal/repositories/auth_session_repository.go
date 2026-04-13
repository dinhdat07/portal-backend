package repositories

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type AuthSessionRepository interface {
	Create(ctx context.Context, session *models.AuthSession) error
	FindActiveByRefreshTokenHash(ctx context.Context, hashToken string) (*models.AuthSession, error)
	FindActiveByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error)
	RotateRefreshToken(ctx context.Context, in domain.RefreshInput) error
	RevokeByID(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}
