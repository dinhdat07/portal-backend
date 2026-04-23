package services

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
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

type auditLogger interface {
	Log(ctx context.Context, meta *domain.AuditMeta, action enum.ActionName, actor *domain.AuditUser, target *domain.AuditUser) error
	LogWithMetadata(ctx context.Context, meta *domain.AuditMeta, action enum.ActionName, actor *domain.AuditUser, target *domain.AuditUser, data map[string]any) error
}
