package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl        string
	JWTSecret    string
	JWTAccessTTL int
	Port         string
}

func Load() (*Config, error) {
	// load .env into os env
	if err := godotenv.Load(); err != nil {
		return nil, errors.New("errors: cannot load env variables")
	}

	ttlStr := os.Getenv("JWT_ACCESS_TTL")
	ttl, _ := strconv.Atoi(ttlStr)

	return &Config{
		DBUrl:        os.Getenv("DB_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTAccessTTL: ttl,
		Port:         os.Getenv("PORT"),
	}, nil
}
