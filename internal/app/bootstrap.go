package app

import (
	"portal-system/internal/auth"
	"portal-system/internal/config"
	portalgrpc "portal-system/internal/grpc"

	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB

	Authenticator *auth.Authenticator
	Authorizer    *auth.Authorizer

	AuthGRPC  *portalgrpc.AuthServer
	UserGRPC  *portalgrpc.UserServer
	AdminGRPC *portalgrpc.AdminServer
}

type Deps struct {
	Config *config.Config
	DB     *gorm.DB

	Authenticator *auth.Authenticator
	Authorizer    *auth.Authorizer

	AuthGRPC  *portalgrpc.AuthServer
	UserGRPC  *portalgrpc.UserServer
	AdminGRPC *portalgrpc.AdminServer
}

func New(deps Deps) *App {
	return &App{
		Config:        deps.Config,
		DB:            deps.DB,
		Authenticator: deps.Authenticator,
		Authorizer:    deps.Authorizer,
		AuthGRPC:      deps.AuthGRPC,
		UserGRPC:      deps.UserGRPC,
		AdminGRPC:     deps.AdminGRPC,
	}
}
