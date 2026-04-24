package composer

import (
	"portal-system/internal/config"
	"portal-system/internal/platform/email"
	redisx "portal-system/internal/platform/redis"
	"portal-system/internal/platform/token"
	"portal-system/internal/services"

	"github.com/redis/go-redis/v9"
)

type Infra struct {
	EmailService    services.EmailSender
	TokenManager    services.TokenIssuer
	RevocationStore services.SessionRevocationStore
}

func newInfra(cfg *config.Config, smtpCfg *config.SMTPConfig, rdb *redis.Client) *Infra {
	return &Infra{
		EmailService:    email.NewSMTPEmailService(*smtpCfg),
		TokenManager:    token.New(cfg.JWTSecret, cfg.JWTAccessTTL),
		RevocationStore: redisx.NewRedisSessionRevocationStore(rdb),
	}
}
