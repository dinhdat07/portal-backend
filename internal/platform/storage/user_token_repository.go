package storage

import (
	"context"
	"errors"
	"time"

	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormUserTokenRepository struct {
	db *gorm.DB
}

func NewGormUserTokenRepository(db *gorm.DB) *GormUserTokenRepository {
	return &GormUserTokenRepository{db: db}
}

func (r *GormUserTokenRepository) Create(ctx context.Context, token *models.UserToken) error {
	return r.getDB(ctx).Create(token).Error
}

func (r *GormUserTokenRepository) FindValidToken(ctx context.Context, tokenHash string, tokenType enum.TokenType) (*models.UserToken, error) {
	var token models.UserToken

	err := r.getDB(ctx).
		Preload("User").
		Where("token_hash = ?", tokenHash).
		Where("token_type = ?", tokenType).
		Where("used_at IS NULL").
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	return &token, nil
}

func (r *GormUserTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.UserToken{}).
		Where("id = ?", id).
		Where("used_at IS NULL").
		Update("used_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

func (r *GormUserTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.UserToken{}).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Update("revoked_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

func (r *GormUserTokenRepository) RevokeByUserAndType(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error {
	now := time.Now().UTC()

	return r.getDB(ctx).
		Model(&models.UserToken{}).
		Where("user_id = ?", userID).
		Where("token_type = ?", tokenType).
		Where("used_at IS NULL").
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		Update("revoked_at", &now).Error
}

func (r *GormUserTokenRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
