package services

import "context"

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error
	SendResetPasswordEmail(ctx context.Context, to, name, resetURL string) error
	SendSetPasswordEmail(ctx context.Context, to, name, setPasswordURL string) error
}
