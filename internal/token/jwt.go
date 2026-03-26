package token

import (
	"errors"
	"portal-system/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string          `json:"user_id"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret   []byte //HMAC requires []byte
	tokenTTL time.Duration
}

func New(secret string, ttlSec int) *Manager {
	tokenTTL := time.Duration(ttlSec) * time.Second
	return &Manager{[]byte(secret), tokenTTL}
}

func (m *Manager) Generate(userID string, role models.UserRole) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "access-token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	// parsing token string need a pointer claim to receive value, and a key function to supply key for verify
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func (m *Manager) ExpiresInSeconds() int {
	return int(m.tokenTTL.Seconds())
}
