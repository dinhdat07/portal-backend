package auth

import (
	"github.com/google/uuid"
)

type Principal struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	RoleID      uuid.UUID `json:"role_id"`
	RoleCode    string
	Permissions []string
}
