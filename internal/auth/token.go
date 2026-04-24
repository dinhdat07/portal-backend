package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleCode  string    `json:"role_code"`
	jwt.RegisteredClaims
}

type GenerateAccessTokenInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	RoleID    uuid.UUID
	RoleCode  string
	Email     string
	Username  string
}

type TokenIssuer interface {
	GenerateAccessToken(input GenerateAccessTokenInput) (string, error)
	GenerateRefreshToken() (string, error)
	ExpiresInSeconds() int
	Parse(tokenString string) (*Claims, error)
	HashToken(raw string) string
	GenerateHashToken() (string, string, error)
}
