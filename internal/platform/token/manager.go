package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"portal-system/internal/services"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret   []byte // HMAC requires []byte
	tokenTTL time.Duration
}

func New(secret string, ttlSec int) *Manager {
	tokenTTL := time.Duration(ttlSec) * time.Second
	return &Manager{[]byte(secret), tokenTTL}
}

func (m *Manager) GenerateAccessToken(input services.GenerateAccessTokenInput) (string, error) {
	now := time.Now()

	claims := claims{
		UserID:    input.UserID,
		SessionID: input.SessionID,
		Username:  input.Username,
		RoleID:    input.RoleID,
		RoleCode:  input.RoleCode,
		Email:     input.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "access-token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) GenerateRefreshToken() (string, error) {
	return m.generateSecureToken(32)
}

func (m *Manager) Parse(tokenString string) (*services.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims{},
		func(t *jwt.Token) (interface{}, error) {
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	parsedClaims, ok := token.Claims.(*claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return toTokenClaims(parsedClaims), nil
}

func (m *Manager) HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (m *Manager) ExpiresInSeconds() int {
	return int(m.tokenTTL.Seconds())
}

func (m *Manager) GenerateHashToken() (string, string, error) {
	rawToken, err := m.generateSecureToken(32)
	if err != nil {
		return "", "", err
	}
	tokenHash := m.HashToken(rawToken)
	return tokenHash, rawToken, nil
}

func (m *Manager) generateSecureToken(length int) (string, error) {
	b := make([]byte, length)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
