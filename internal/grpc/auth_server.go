package portalgrpc

import (
	"context"
	authv1 "portal-system/gen/go/auth/v1"
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	mappers "portal-system/internal/grpc/mapper"
	"portal-system/internal/services"
	"time"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	authService services.AuthService
}

func NewAuthServer(authService services.AuthService) *AuthServer {
	return &AuthServer{authService: authService}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	dob, err := time.Parse("2006-01-02", req.GetDob())
	if err != nil {
		return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
	}

	meta := getAuditFromCtx(ctx)

	err = s.authService.Register(
		ctx,
		meta,
		req.GetEmail(),
		req.GetUsername(),
		req.GetPassword(),
		req.GetFirstName(),
		req.GetLastName(),
		dob,
	)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "Registration successful",
	}, nil
}

func (s *AuthServer) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	meta := getAuditFromCtx(ctx)

	if err := s.authService.VerifyEmail(ctx, meta, req.GetToken(), enum.TokenTypeEmailVerification); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "Email verification successful",
	}, nil
}

func (s *AuthServer) ResendVerification(ctx context.Context, req *authv1.ResendVerificationRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	meta := getAuditFromCtx(ctx)

	if err := s.authService.ResendVerification(ctx, meta, req.GetEmail(), enum.TokenTypeEmailVerification); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "Resend verification successfully",
	}, nil
}

func (s *AuthServer) SetPassword(ctx context.Context, req *authv1.SetPasswordRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	meta := getAuditFromCtx(ctx)

	input := &domain.SetPasswordInput{
		Token:           req.GetToken(),
		Password:        req.GetPassword(),
		ConfirmPassword: req.GetConfirmPassword(),
	}

	if err := s.authService.SetPassword(ctx, meta, input, enum.TokenTypePasswordSet); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "Email verification and password set successful",
	}, nil
}

func (s *AuthServer) ResetPassword(ctx context.Context, req *authv1.SetPasswordRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	meta := getAuditFromCtx(ctx)

	input := &domain.SetPasswordInput{
		Token:           req.GetToken(),
		Password:        req.GetPassword(),
		ConfirmPassword: req.GetConfirmPassword(),
	}

	if err := s.authService.ResetPassword(ctx, meta, input, enum.TokenTypePasswordReset); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "Password reset successful",
	}, nil
}

func (s *AuthServer) ForgotPassword(ctx context.Context, req *authv1.ForgotPasswordRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	meta := getAuditFromCtx(ctx)

	if err := s.authService.ForgotPassword(ctx, meta, req.GetEmail()); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "If the account exists, a password reset email has been sent",
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	meta := getAuditFromCtx(ctx)

	result, err := s.authService.LogIn(ctx, meta, req.GetIdentifier(), req.GetPassword())
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.LoginResultToPB(result), nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	meta := getAuditFromCtx(ctx)

	result, err := s.authService.Refresh(ctx, meta, req.GetRefreshToken())
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.RefreshResultToPB(result), nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*commonv1.MessageResponse, error) {
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	sessionID, err := getSessionIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	if err := s.authService.Logout(ctx, meta, actor, sessionID); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "Logout successful",
	}, nil
}

func (s *AuthServer) LogoutAll(ctx context.Context, req *authv1.LogoutAllRequest) (*commonv1.MessageResponse, error) {
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	if err := s.authService.LogoutAll(ctx, meta, actor); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "All devices has been logged out",
	}, nil
}
