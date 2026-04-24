package app

import (
	"net"

	adminv1 "portal-system/gen/go/admin/v1"
	authv1 "portal-system/gen/go/auth/v1"
	userv1 "portal-system/gen/go/user/v1"
	"portal-system/internal/domain/constants"
	"portal-system/internal/grpc/interceptor"

	"google.golang.org/grpc"
)

func (a *App) NewGRPCServer() *grpc.Server {
	publicMethods := buildGRPCPublicMethods()
	methodPermissions := buildGRPCMethodPermissions()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.AuthenticationInterceptor(a.Authenticator, publicMethods),
			interceptor.PermissionInterceptor(a.Authorizer, methodPermissions),
		),
	)

	authv1.RegisterAuthServiceServer(s, a.AuthGRPC)
	adminv1.RegisterAdminServiceServer(s, a.AdminGRPC)
	userv1.RegisterUserServiceServer(s, a.UserGRPC)

	return s
}

func RunGRPCServer(addr string, server *grpc.Server) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return server.Serve(lis)
}

func buildGRPCPublicMethods() map[string]bool {
	return map[string]bool{
		// auth public
		authv1.AuthService_Register_FullMethodName:           true,
		authv1.AuthService_Login_FullMethodName:              true,
		authv1.AuthService_VerifyEmail_FullMethodName:        true,
		authv1.AuthService_ResendVerification_FullMethodName: true,
		authv1.AuthService_SetPassword_FullMethodName:        true,
		authv1.AuthService_ResetPassword_FullMethodName:      true,
		authv1.AuthService_ForgotPassword_FullMethodName:     true,
		authv1.AuthService_RefreshToken_FullMethodName:       true,
	}
}

// auth-only:
//   logout, logout-all

func buildGRPCMethodPermissions() map[string]constants.PermissionCode {
	return map[string]constants.PermissionCode{
		// user self-service
		userv1.UserService_GetMyProfile_FullMethodName:     constants.PermProfileReadSelf,
		userv1.UserService_UpdateMyProfile_FullMethodName:  constants.PermProfileUpdateSelf,
		userv1.UserService_ChangeMyPassword_FullMethodName: constants.PermProfileChangePassword,

		// admin user management
		adminv1.AdminService_ListUsers_FullMethodName:      constants.PermUserList,
		adminv1.AdminService_CreateUser_FullMethodName:     constants.PermUserCreate,
		adminv1.AdminService_GetUserDetail_FullMethodName:  constants.PermUserReadDetail,
		adminv1.AdminService_UpdateUser_FullMethodName:     constants.PermUserUpdate,
		adminv1.AdminService_DeleteUser_FullMethodName:     constants.PermUserDelete,
		adminv1.AdminService_RestoreUser_FullMethodName:    constants.PermUserRestore,
		adminv1.AdminService_UpdateUserRole_FullMethodName: constants.PermUserRoleUpdate,
	}
}
