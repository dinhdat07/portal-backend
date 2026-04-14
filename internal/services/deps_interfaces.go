package services

import (
	"context"

	"github.com/google/uuid"
)

type emailSender interface {
	SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error
	SendResetPasswordEmail(ctx context.Context, to, name, resetURL string) error
	SendSetPasswordEmail(ctx context.Context, to, name, setPasswordURL string) error
}

type tokenIssuer interface {
	GenerateAccessToken(userID uuid.UUID, sessionID uuid.UUID, roleID uuid.UUID, roleCode string, email string, username string) (string, error)
	GenerateRefreshToken() (string, error)
	ExpiresInSeconds() int
}
