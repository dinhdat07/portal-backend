package auth

import (
	"portal-system/internal/domain/constants"

	"github.com/google/uuid"
)

type Principal struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	RoleID      uuid.UUID `json:"role_id"`
	RoleCode    constants.RoleCode
	Permissions []string
}
