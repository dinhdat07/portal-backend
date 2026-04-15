package storage

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"

	"gorm.io/gorm"
)

type GormAuditLogRepository struct {
	db *gorm.DB
}

func NewGormAuditLogRepository(db *gorm.DB) *GormAuditLogRepository {
	return &GormAuditLogRepository{db: db}
}

func (r *GormAuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.getDB(ctx).Create(log).Error
}

func (r *GormAuditLogRepository) List(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error) {
	db := r.getDB(ctx).Model(&models.AuditLog{})

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
	if filter.PageSize > 100 {
		filter.PageSize = 100
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

func (r *GormAuditLogRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
