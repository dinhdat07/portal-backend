package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleCode string    `json:"role_code"`
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

func (m *Manager) Generate(userID uuid.UUID, roleId uuid.UUID, roleCode string, email string, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleId,
		RoleCode: roleCode,
		Email:    email,
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
