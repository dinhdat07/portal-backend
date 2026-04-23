package token

import "github.com/google/uuid"

type GenerateAccessTokenInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	RoleID    uuid.UUID
	RoleCode  string
	Email     string
	Username  string
}
