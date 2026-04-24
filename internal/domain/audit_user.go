package domain

import (
	"portal-system/internal/domain/constants"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type AuditUser struct {
	ID       uuid.UUID
	Username string
	Email    string
	RoleCode constants.RoleCode
}

func MapUserToAuditUser(u *models.User) *AuditUser {
	if u == nil {
		return nil
	}

	return &AuditUser{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		RoleCode: u.Role.Code,
	}
}
