package repositories

import (
	"context"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type AuthSessionRepository interface {
	Create(ctx context.Context, session *models.AuthSession) error
	FindActiveByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error)
	RevokeByID(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}
