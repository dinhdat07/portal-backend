package app

import (
	"portal-system/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.AuditLog{},
		&models.UserToken{},
		&models.Role{},
		&models.Permission{},
		&models.AuthSession{},
		&models.RefreshToken{},
	)
}
