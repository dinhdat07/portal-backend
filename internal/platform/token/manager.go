package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"portal-system/internal/auth"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type manager struct {
	secret   []byte // HMAC requires []byte
	tokenTTL time.Duration
}

func New(secret string, ttlSec int) auth.TokenIssuer {
	tokenTTL := time.Duration(ttlSec) * time.Second
	return &manager{[]byte(secret), tokenTTL}
}

func (m *manager) GenerateAccessToken(input auth.GenerateAccessTokenInput) (string, error) {
	claims := auth.Claims{
		UserID:    input.UserID,
		SessionID: input.SessionID,
		Username:  input.Username,
		RoleID:    input.RoleID,
		RoleCode:  input.RoleCode,
		Email:     input.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "access-token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *manager) GenerateRefreshToken() (string, error) {
	return m.generateSecureToken(32)
}

func (m *manager) Parse(tokenString string) (*auth.Claims, error) {
	// parsing token string need a pointer claim to receive value, and a key function to supply key for verify
	token, err := jwt.ParseWithClaims(
		tokenString,
		&auth.Claims{},
		func(t *jwt.Token) (interface{}, error) {
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*auth.Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func (m *manager) HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (m *manager) ExpiresInSeconds() int {
	return int(m.tokenTTL.Seconds())
}

func (m *manager) GenerateHashToken() (string, string, error) {
	rawToken, err := m.generateSecureToken(32)
	if err != nil {
		return "", "", err
	}
	tokenHash := m.HashToken(rawToken)
	return tokenHash, rawToken, nil
}

func (m *manager) generateSecureToken(length int) (string, error) {
	b := make([]byte, length)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
