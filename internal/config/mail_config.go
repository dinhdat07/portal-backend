package config

import "errors"

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
	UseAuth  bool
	UseTLS   bool
}

func LoadSMTPConfig() (*SMTPConfig, error) {
	loadEnv()

	host := getEnvDefault("SMTP_HOST", "")
	port := getEnvDefault("SMTP_PORT", "")
	if host == "" || port == "" {
		return nil, errors.New("missing SMTP_HOST or SMTP_PORT")
	}

	useAuth, err := getEnvBool("SMTP_USE_AUTH", false)
	if err != nil {
		return nil, err
	}

	useTLS, err := getEnvBool("SMTP_USE_TLS", false)
	if err != nil {
		return nil, err
	}

	return &SMTPConfig{
		Host:     host,
		Port:     port,
		Username: getEnvDefault("SMTP_USERNAME", ""),
		Password: getEnvDefault("SMTP_PASSWORD", ""),
		From:     getEnvDefault("SMTP_FROM", "noreply@portal.local"),
		FromName: getEnvDefault("SMTP_FROM_NAME", "Portal System"),
		UseAuth:  useAuth,
		UseTLS:   useTLS,
	}, nil
}
