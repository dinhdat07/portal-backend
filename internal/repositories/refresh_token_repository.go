package repositories

import (
	"context"

	"portal-system/internal/models"

	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	RevokeByFamilyID(ctx context.Context, familyID uuid.UUID) error
	MarkReplacement(ctx context.Context, id uuid.UUID, replacementID uuid.UUID) error
}
