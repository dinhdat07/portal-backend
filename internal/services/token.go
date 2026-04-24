package services

import "github.com/google/uuid"

type GenerateAccessTokenInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	RoleID    uuid.UUID
	RoleCode  string
	Email     string
	Username  string
}

type TokenClaims struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Username  string
	Email     string
	RoleID    uuid.UUID
	RoleCode  string
}

type TokenIssuer interface {
	GenerateAccessToken(input GenerateAccessTokenInput) (string, error)
	GenerateRefreshToken() (string, error)
	ExpiresInSeconds() int
	Parse(tokenString string) (*TokenClaims, error)
	HashToken(raw string) string
	GenerateHashToken() (string, string, error)
}
