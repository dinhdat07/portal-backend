package composer

import (
	"portal-system/internal/platform/storage"

	"gorm.io/gorm"
)

type Repositories struct {
	UserRepo    *storage.GormUserRepository
	AuditLog    *storage.GormAuditLogRepository
	TokenRepo   *storage.GormUserTokenRepository
	RoleRepo    *storage.GormRoleRepository
	SessionRepo *storage.GormAuthSessionRepository
	RefreshRepo *storage.GormRefreshTokenRepository
	TxManager   *storage.GormTxManager
}

func newRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserRepo:    storage.NewGormUserRepository(db),
		AuditLog:    storage.NewGormAuditLogRepository(db),
		TokenRepo:   storage.NewGormUserTokenRepository(db),
		RoleRepo:    storage.NewGormRoleRepository(db),
		SessionRepo: storage.NewGormAuthSessionRepository(db),
		RefreshRepo: storage.NewGormRefreshTokenRepository(db),
		TxManager:   storage.NewGormTxManager(db),
	}
}
