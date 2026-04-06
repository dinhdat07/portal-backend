package auth

import (
	"portal-system/internal/platform/token"
)

type Authenticator struct {
	manager *token.Manager
}

func NewAuthenticator(manager *token.Manager) *Authenticator {
	return &Authenticator{manager: manager}
}

func (a *Authenticator) Authenticate(tokenString string) (*Principal, error) {
	claims, err := a.manager.Parse(tokenString)
	if err != nil {
		return nil, err
	}

	principal := &Principal{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     claims.Role,
	}

	return principal, nil
}
