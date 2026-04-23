package services

import (
	"context"
	"portal-system/internal/platform/token"
)

type emailSender interface {
	SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error
	SendResetPasswordEmail(ctx context.Context, to, name, resetURL string) error
	SendSetPasswordEmail(ctx context.Context, to, name, setPasswordURL string) error
}

type tokenIssuer interface {
	GenerateAccessToken(input token.GenerateAccessTokenInput) (string, error)
	GenerateRefreshToken() (string, error)
	ExpiresInSeconds() int
}
