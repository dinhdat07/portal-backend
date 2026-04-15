package interceptor

import (
	"context"
	"portal-system/internal/auth"
	"portal-system/internal/domain/constants"
	portalgrpc "portal-system/internal/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PermissionInterceptor(authorizer *auth.Authorizer, methodPermissions map[string]constants.PermissionCode) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authorizer == nil {
			return nil, status.Error(codes.Internal, "internal server error")
		}

		requiredPerm, hasRule := methodPermissions[info.FullMethod]
		if !hasRule {
			return handler(ctx, req)
		}

		principal, ok := portalgrpc.GetPrincipal(ctx)
		if !ok || principal == nil {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}

		allowed := authorizer.HasPermission(ctx, principal, requiredPerm)
		if !allowed {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		ctx = portalgrpc.SetPrincipal(ctx, principal)
		return handler(ctx, req)

	}
}
