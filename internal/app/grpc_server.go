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
		"/auth.v1.AuthService/RegisterUser":       true,
		"/auth.v1.AuthService/LogIn":              true,
		"/auth.v1.AuthService/VerifyEmail":        true,
		"/auth.v1.AuthService/ResendVerification": true,
		"/auth.v1.AuthService/SetPassword":        true,
		"/auth.v1.AuthService/ResetPassword":      true,
		"/auth.v1.AuthService/ForgotPassword":     true,
		"/auth.v1.AuthService/Refresh":            true,
	}
}

// auth-only:
//   logout, logout-all

func buildGRPCMethodPermissions() map[string]constants.PermissionCode {
	return map[string]constants.PermissionCode{
		// user self-service
		"/user.v1.UserService/GetMyProfile":     constants.PermProfileReadSelf,
		"/user.v1.UserService/UpdateProfile":    constants.PermProfileUpdateSelf,
		"/user.v1.UserService/ChangeMyPassword": constants.PermProfileChangePassword,

		// admin user management
		"/admin.v1.AdminService/ListUsers":     constants.PermUserList,
		"/admin.v1.AdminService/CreateUser":    constants.PermUserCreate,
		"/admin.v1.AdminService/GetUserDetail": constants.PermUserReadDetail,
		"/admin.v1.AdminService/UpdateUser":    constants.PermUserUpdate,
		"/admin.v1.AdminService/DeleteUser":    constants.PermUserDelete,
		"/admin.v1.AdminService/RestoreUser":   constants.PermUserRestore,
		"/admin.v1.AdminService/UpdateRole":    constants.PermUserRoleUpdate,
	}
}
