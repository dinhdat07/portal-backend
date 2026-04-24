package repositories

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	List(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error)
}
