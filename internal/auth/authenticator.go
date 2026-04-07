package auth

import (
	"context"
	"portal-system/internal/platform/token"
	"portal-system/internal/repositories"
)

type Authenticator struct {
	manager  *token.Manager
	roleRepo *repositories.RoleRepository
}

func NewAuthenticator(manager *token.Manager) *Authenticator {
	return &Authenticator{manager: manager}
}

func (a *Authenticator) Authenticate(ctx context.Context, tokenString string) (*Principal, error) {
	claims, err := a.manager.Parse(tokenString)
	if err != nil {
		return nil, err
	}

	role, err := a.roleRepo.GetWithPermissions(ctx, claims.RoleID)
	if err != nil {
		return nil, err
	}

	perms := make([]string, 0, len(role.Permissions))

	for _, p := range role.Permissions {
		perms = append(perms, p.Code)
	}

	principal := &Principal{
		UserID:      claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		RoleID:      claims.RoleID,
		Permissions: perms,
	}

	return principal, nil

}
