package composer

import portalgrpc "portal-system/internal/grpc"

type GRPCServers struct {
	Auth  *portalgrpc.AuthServer
	User  *portalgrpc.UserServer
	Admin *portalgrpc.AdminServer
}

func newGRPCServers(svcs *Services) *GRPCServers {
	return &GRPCServers{
		Auth:  portalgrpc.NewAuthServer(svcs.Auth),
		User:  portalgrpc.NewUserServer(svcs.User),
		Admin: portalgrpc.NewAdminServer(svcs.Admin, svcs.User),
	}
}
