package composer

import (
	"context"
	"portal-system/internal/app"
	"portal-system/internal/bootstrap"
	"portal-system/internal/config"
	redisx "portal-system/internal/platform/redis"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Composer() (*app.App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	smtpCfg, err := config.LoadSMTPConfig()
	if err != nil {
		return nil, err
	}

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	rdb := redisx.NewClient(redisCfg)
	if err := redisx.Ping(context.Background(), rdb); err != nil {
		return nil, err
	}

	if err := bootstrap.AutoMigrate(db); err != nil {
		return nil, err
	}

	if cfg.Env == "development" {
		if err := bootstrap.SeedPermissions(db); err != nil {
			return nil, err
		}
		if err := bootstrap.SeedRoles(db); err != nil {
			return nil, err
		}
		if err := bootstrap.SeedRolePermissions(db); err != nil {
			return nil, err
		}
		if err := bootstrap.SeedAdmin(db, cfg); err != nil {
			return nil, err
		}
	}

	infra := newInfra(cfg, smtpCfg, rdb)
	repos := newRepositories(db)
	svcs := newServices(cfg, infra, repos)
	grpcServers := newGRPCServers(svcs)

	return app.New(app.Deps{
		Config:        cfg,
		DB:            db,
		Authenticator: svcs.Authenticator,
		Authorizer:    svcs.Authorizer,
		AuthGRPC:      grpcServers.Auth,
		UserGRPC:      grpcServers.User,
		AdminGRPC:     grpcServers.Admin,
	}), nil
}
