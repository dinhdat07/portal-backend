package auth

import (
	"portal-system/internal/domain/enum"

	"github.com/google/uuid"
)

type Principal struct {
	UserID   uuid.UUID     `json:"user_id"`
	Username string        `json:"username"`
	Email    string        `json:"email"`
	Role     enum.UserRole `json:"role"`
}
