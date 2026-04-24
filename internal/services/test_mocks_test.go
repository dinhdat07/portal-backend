package services

import (
	"context"
	"time"

	"portal-system/internal/platform/token"
)

func cloneUser[T any](v *T) *T {
	if v == nil {
		return nil
	}
	cpy := *v
	return &cpy
}

func ptrString(s string) *string {
	return &s
}

type emailSenderMock struct {
	sendVerificationFn func(ctx context.Context, to, name, verifyURL string) error
	sendResetFn        func(ctx context.Context, to, name, resetURL string) error
	sendSetFn          func(ctx context.Context, to, name, setPasswordURL string) error
	verificationCalled int
	resetCalled        int
	setCalled          int
}

func (m *emailSenderMock) SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error {
	m.verificationCalled++
	if m.sendVerificationFn != nil {
		return m.sendVerificationFn(ctx, to, name, verifyURL)
	}
	return nil
}

func (m *emailSenderMock) SendResetPasswordEmail(ctx context.Context, to, name, resetURL string) error {
	m.resetCalled++
	if m.sendResetFn != nil {
		return m.sendResetFn(ctx, to, name, resetURL)
	}
	return nil
}

func (m *emailSenderMock) SendSetPasswordEmail(ctx context.Context, to, name, setPasswordURL string) error {
	m.setCalled++
	if m.sendSetFn != nil {
		return m.sendSetFn(ctx, to, name, setPasswordURL)
	}
	return nil
}

var _ emailSender = (*emailSenderMock)(nil)

type tokenIssuerMock struct {
	generateAccessTokenFn func(input token.GenerateAccessTokenInput) (string, error)
	generateRefreshFn     func() (string, error)
	expiresInSecondsFn    func() int
	accessCalled          int
	refreshCalled         int
}

func (m *tokenIssuerMock) GenerateAccessToken(input token.GenerateAccessTokenInput) (string, error) {
	m.accessCalled++
	if m.generateAccessTokenFn != nil {
		return m.generateAccessTokenFn(input)
	}
	return "access-token", nil
}

func (m *tokenIssuerMock) GenerateRefreshToken() (string, error) {
	m.refreshCalled++
	if m.generateRefreshFn != nil {
		return m.generateRefreshFn()
	}
	return "refresh-token", nil
}

func (m *tokenIssuerMock) ExpiresInSeconds() int {
	if m.expiresInSecondsFn != nil {
		return m.expiresInSecondsFn()
	}
	return 3600
}

var _ tokenIssuer = (*tokenIssuerMock)(nil)

func ptrTime(v time.Time) *time.Time {
	return &v
}
