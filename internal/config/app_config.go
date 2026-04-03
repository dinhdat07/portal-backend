package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl              string
	JWTSecret          string
	JWTAccessTTL       int
	Port               string
	Env                string
	AdminEmail         string
	AdminPassword      string
	ApiBaseUrl         string
	FrontEndUrl        string
	AuthCookieName     string
	AuthCookieDomain   string
	AuthCookieSecure   bool
	AuthCookieSameSite string
	CSRFCookieName     string
	CSRFHeaderName     string
}

func Load() (*Config, error) {
	// load .env into os env
	if err := godotenv.Load(); err != nil {
		return nil, errors.New("errors: cannot load env variables")
	}

	ttlStr := os.Getenv("JWT_ACCESS_TTL")
	ttl, _ := strconv.Atoi(ttlStr)

	cookieName := os.Getenv("AUTH_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "portal_access_token"
	}

	cookieSameSite := os.Getenv("AUTH_COOKIE_SAME_SITE")
	if cookieSameSite == "" {
		cookieSameSite = "lax"
	}

	cookieSecure, err := strconv.ParseBool(os.Getenv("AUTH_COOKIE_SECURE"))
	if err != nil {
		cookieSecure = os.Getenv("ENV") != "development"
	}

	csrfCookieName := os.Getenv("CSRF_COOKIE_NAME")
	if csrfCookieName == "" {
		csrfCookieName = "portal_csrf_token"
	}

	csrfHeaderName := os.Getenv("CSRF_HEADER_NAME")
	if csrfHeaderName == "" {
		csrfHeaderName = "X-CSRF-Token"
	}

	return &Config{
		DBUrl:              os.Getenv("DB_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTAccessTTL:       ttl,
		Port:               os.Getenv("PORT"),
		Env:                os.Getenv("ENV"),
		AdminEmail:         os.Getenv("ADMIN_EMAIL"),
		AdminPassword:      os.Getenv("ADMIN_PASSWORD"),
		ApiBaseUrl:         os.Getenv("API_BASE_URL"),
		FrontEndUrl:        os.Getenv("FRONTEND_BASE_URL"),
		AuthCookieName:     cookieName,
		AuthCookieDomain:   os.Getenv("AUTH_COOKIE_DOMAIN"),
		AuthCookieSecure:   cookieSecure,
		AuthCookieSameSite: cookieSameSite,
		CSRFCookieName:     csrfCookieName,
		CSRFHeaderName:     csrfHeaderName,
	}, nil
}
