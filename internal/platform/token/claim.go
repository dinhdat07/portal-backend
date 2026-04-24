package token

import (
	"portal-system/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type claims struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleCode  string    `json:"role_code"`
	jwt.RegisteredClaims
}

func toTokenClaims(c *claims) *services.TokenClaims {
	return &services.TokenClaims{
		UserID:    c.UserID,
		SessionID: c.SessionID,
		Username:  c.Username,
		Email:     c.Email,
		RoleID:    c.RoleID,
		RoleCode:  c.RoleCode,
	}
}
