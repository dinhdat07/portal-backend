package storage

import (
	"context"
	"portal-system/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormAuthSessionRepository struct {
	db *gorm.DB
}

func NewGormAuthSessionRepository(db *gorm.DB) *GormAuthSessionRepository {
	return &GormAuthSessionRepository{db: db}
}

func (r *GormAuthSessionRepository) Create(ctx context.Context, session *models.AuthSession) error {
	return r.getDB(ctx).Create(session).Error
}

func (r *GormAuthSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, hashToken string) (*models.AuthSession, error) {
	var AuthSession models.AuthSession

	err := r.getDB(ctx).
		Where("refresh_token_hash = ?", hashToken).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&AuthSession).Error

	if err != nil {
		return nil, err
	}

	return &AuthSession, nil
}

func (r *GormAuthSessionRepository) FindActiveByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
	var AuthSession models.AuthSession

	err := r.getDB(ctx).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&AuthSession).Error

	if err != nil {
		return nil, err
	}

	return &AuthSession, nil
}

func (r *GormAuthSessionRepository) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.AuthSession{}).
		Where("id = ?", sessionID).
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

func (r *GormAuthSessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.AuthSession{}).
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

func (r *GormAuthSessionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
