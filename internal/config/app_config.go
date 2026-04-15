package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCPort string
	HTTPPort string

	DBUrl         string
	JWTSecret     string
	JWTAccessTTL  int
	RefreshTTL    int
	Port          string
	Env           string
	AdminEmail    string
	AdminPassword string
	ApiBaseUrl    string
	FrontEndUrl   string
}

func Load() (*Config, error) {
	// load .env into os env
	if err := godotenv.Load(); err != nil {
		return nil, errors.New("errors: cannot load env variables")
	}

	accessTTL, err := strconv.Atoi(os.Getenv("JWT_ACCESS_TTL"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	refreshTTL, err := strconv.Atoi(os.Getenv("JWT_REFRESH_TTL"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	return &Config{
		DBUrl:         os.Getenv("DB_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTAccessTTL:  accessTTL,
		RefreshTTL:    refreshTTL,
		Port:          os.Getenv("PORT"),
		Env:           os.Getenv("ENV"),
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		ApiBaseUrl:    os.Getenv("API_BASE_URL"),
		FrontEndUrl:   os.Getenv("FRONTEND_BASE_URL"),
	}, nil
}
