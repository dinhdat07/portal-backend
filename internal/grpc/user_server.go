package portalgrpc

import (
	"context"
	commonv1 "portal-system/gen/go/common/v1"
	userv1 "portal-system/gen/go/user/v1"
	"portal-system/internal/domain"
	mappers "portal-system/internal/grpc/mapper"
	"portal-system/internal/services"
	"time"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	userService services.UserService
}

func NewUserServer(userService services.UserService) *UserServer {
	return &UserServer{userService: userService}
}

func (s *UserServer) GetMyProfile(ctx context.Context, req *userv1.GetMyProfileRequest) (*commonv1.User, error) {
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	user, err := s.userService.GetProfile(ctx, meta, actor, actor.ID)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *UserServer) UpdateMyProfile(ctx context.Context, req *userv1.UpdateMyProfileRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	input := domain.UpdateUserInput{}

	if req.FirstName != nil {
		v := req.GetFirstName()
		input.FirstName = &v
	}
	if req.LastName != nil {
		v := req.GetLastName()
		input.LastName = &v
	}
	if req.Username != nil {
		v := req.GetUsername()
		input.Username = &v
	}
	if req.Dob != nil {
		dob, err := time.Parse("2006-01-02", req.GetDob())
		if err != nil {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
		}
		input.DOB = &dob
	}

	user, err := s.userService.UpdateProfile(ctx, meta, actor, actor.ID, input)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *UserServer) ChangeMyPassword(ctx context.Context, req *userv1.ChangeMyPasswordRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	if err = s.userService.ChangePassword(ctx, meta, actor,
		req.GetCurrentPassword(),
		req.GetNewPassword(),
		req.GetConfirmNewPassword()); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "password changed successfully",
	}, nil
}
