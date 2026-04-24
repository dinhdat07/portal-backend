package storage

import (
	"context"
	"portal-system/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormRefreshTokenRepository struct {
	db *gorm.DB
}

func NewGormRefreshTokenRepository(db *gorm.DB) *GormRefreshTokenRepository {
	return &GormRefreshTokenRepository{db: db}
}

func (r *GormRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	return r.getDB(ctx).Create(token).Error
}

func (r *GormRefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken

	err := r.getDB(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *GormRefreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Update("revoked_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Update("revoked_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormRefreshTokenRepository) RevokeBySessionID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.RefreshToken{}).
		Where("session_id = ?", userID).
		Where("revoked_at IS NULL").
		Update("revoked_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormRefreshTokenRepository) MarkReplacement(ctx context.Context, id uuid.UUID, replacementID uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("replaced_by_token_id", replacementID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormRefreshTokenRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
