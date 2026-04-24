package models

import (
	"portal-system/internal/domain/constants"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID          `gorm:"type:uuid;primaryKey"`
	Code        constants.RoleCode `gorm:"size:50;uniqueIndex;not null"`
	Name        string             `gorm:"size:100;not null"`
	Permissions []Permission       `gorm:"many2many:role_permissions;"`
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}
