package repositories

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthSessionRepository struct {
	db *gorm.DB
}

func NewAuthSessionRepository(db *gorm.DB) *AuthSessionRepository {
	return &AuthSessionRepository{db: db}
}

func (r *AuthSessionRepository) Create(ctx context.Context, session *models.AuthSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *AuthSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, hashToken string) (*models.AuthSession, error) {
	var AuthSession models.AuthSession

	err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", hashToken).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&AuthSession).Error

	if err != nil {
		return nil, err
	}

	return &AuthSession, nil
}

func (r *AuthSessionRepository) FindActiveByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
	var AuthSession models.AuthSession

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&AuthSession).Error

	if err != nil {
		return nil, err
	}

	return &AuthSession, nil
}

func (r *AuthSessionRepository) RotateRefreshToken(ctx context.Context, in domain.RefreshInput) error {
	return r.db.WithContext(ctx).
		Model(&models.AuthSession{}).
		Where("id = ?", in.SessionID).
		Updates(map[string]any{
			"refresh_token_hash": in.NewTokenHash,
			"expires_at":         in.NewExpiresAt,
			"last_used_at":       in.RotatedAt,
		}).Error
}

func (r *AuthSessionRepository) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
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

func (r *AuthSessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
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

func (r *AuthSessionRepository) WithTx(tx *gorm.DB) *AuthSessionRepository {
	return NewAuthSessionRepository(tx)
}
