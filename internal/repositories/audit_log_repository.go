package repositories

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"

	"gorm.io/gorm"
)

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditLogRepository) List(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error) {
	var auditLog models.AuditLog

	db := r.db.WithContext(ctx).Model(&auditLog)

	if filter.Action != "" {
		db = db.Where("action = ?", filter.Action)
	}

	if filter.ActorUserID != nil {
		db = db.Where("actor_user_id = ?", *filter.ActorUserID)
	}

	if filter.TargetUserID != nil {
		db = db.Where("target_user_id = ?", *filter.TargetUserID)
	}

	if filter.From != nil {
		db = db.Where("created_at >= ?", *filter.From)
	}

	if filter.To != nil {
		db = db.Where("created_at <= ?", *filter.To)
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.AuditLog
	offset := (filter.Page - 1) * filter.PageSize

	err := db.
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditLogRepository) WithTx(tx *gorm.DB) *AuditLogRepository {
	return NewAuditLogRepository(tx)
}
