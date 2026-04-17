package portalgrpc

import (
	"context"
	adminv1 "portal-system/gen/go/admin/v1"
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/domain"
	mappers "portal-system/internal/grpc/mapper"
	"portal-system/internal/services"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
	dateLayout      = "2006-01-02"
)

type AdminServer struct {
	adminv1.UnimplementedAdminServiceServer
	adminService *services.AdminService
	userService  *services.UserService
}

func NewAdminServer(adminService *services.AdminService, userService *services.UserService) *AdminServer {
	return &AdminServer{
		adminService: adminService,
		userService:  userService,
	}
}

func (s *AdminServer) ListUsers(ctx context.Context, req *adminv1.ListUsersRequest) (*adminv1.ListUsersResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page == 0 {
		page = defaultPage
	}

	pageSize := int(req.GetPageSize())
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	if page < 1 || pageSize < 1 || pageSize > maxPageSize {
		return nil, gstatus.Error(codes.InvalidArgument, "invalid pagination")
	}

	filter := domain.UsersFilter{
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: req.GetIncludeDeleted(),
	}

	if req.Username != nil {
		filter.Username = req.GetUsername()
	}
	if req.Email != nil {
		filter.Email = req.GetEmail()
	}
	if req.FullName != nil {
		filter.FullName = req.GetFullName()
	}
	if req.Dob != nil {
		dob, err := time.Parse(dateLayout, req.GetDob())
		if err != nil {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
		}
		filter.Dob = &dob
	}
	if req.Role != nil {
		roleCode, ok := mappers.RoleCodeFromPB(req.GetRole())
		if !ok {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid role code")
		}
		filter.RoleCode = &roleCode
	}
	if req.Status != nil {
		status, ok := mappers.UserStatusFromPB(req.GetStatus())
		if !ok {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid status")
		}
		filter.Status = status
	}

	meta := getAuditFromCtx(ctx)
	result, err := s.adminService.ListUsers(ctx, meta, actor, filter)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.ListUsersResultToPB(result), nil
}

func (s *AdminServer) CreateUser(ctx context.Context, req *adminv1.CreateUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	roleCode, ok := mappers.RoleCodeFromPB(req.GetRole())
	if !ok {

		return nil, gstatus.Error(codes.InvalidArgument, "invalid role code")
	}

	dob, err := time.Parse(dateLayout, req.GetDob())
	if err != nil {
		return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
	}

	input := domain.CreateUserInput{
		Email:     req.GetEmail(),
		Username:  req.GetUsername(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		RoleCode:  roleCode,
		DOB:       &dob,
	}

	user, err := s.adminService.CreateUser(ctx, meta, actor, input)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) GetUserDetail(ctx context.Context, req *adminv1.GetUserDetailRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.userService.GetProfile(ctx, meta, actor, userID)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) UpdateUser(ctx context.Context, req *adminv1.UpdateUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

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
		dob, err := time.Parse(dateLayout, req.GetDob())
		if err != nil {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
		}
		input.DOB = &dob
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.userService.UpdateProfile(ctx, meta, actor, userID, input)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) DeleteUser(ctx context.Context, req *adminv1.DeleteUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.adminService.DeleteUser(ctx, meta, actor, userID)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) RestoreUser(ctx context.Context, req *adminv1.RestoreUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.adminService.RestoreUser(ctx, meta, actor, userID)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) UpdateUserRole(ctx context.Context, req *adminv1.UpdateUserRoleRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	roleCode, ok := mappers.RoleCodeFromPB(req.GetRole())
	if !ok {
		return nil, gstatus.Error(codes.InvalidArgument, "invalid role code")
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.adminService.UpdateRole(ctx, meta, actor, userID, roleCode)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func parseUserID(id string) (uuid.UUID, error) {
	if id == "" {
		return uuid.Nil, gstatus.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, gstatus.Error(codes.InvalidArgument, "invalid user_id")
	}

	return userID, nil
}
