package app

import (
	"context"
	"net/http"

	adminv1 "portal-system/gen/go/admin/v1"
	authv1 "portal-system/gen/go/auth/v1"
	userv1 "portal-system/gen/go/user/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func (a *App) NewGatewayMux(ctx context.Context, grpcAddr string) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := metadata.MD{}

			if auth := r.Header.Get("Authorization"); auth != "" {
				md.Set("authorization", auth)
			}
			if ua := r.UserAgent(); ua != "" {
				md.Set("user-agent", ua)
			}
			if ip := r.RemoteAddr; ip != "" {
				md.Set("x-forwarded-for", ip)
			}

			return md
		}),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return nil, err
	}

	if err := userv1.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return nil, err
	}

	if err := adminv1.RegisterAdminServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return nil, err
	}

	return mux, nil
}

func RunGatewayServer(addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return srv.ListenAndServe()
}
