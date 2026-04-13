package repositories

import (
	"context"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type UserTokenRepository interface {
	Create(ctx context.Context, token *models.UserToken) error
	FindValidToken(ctx context.Context, tokenHash string, tokenType enum.TokenType) (*models.UserToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByUserAndType(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error
}
