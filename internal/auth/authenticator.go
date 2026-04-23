package auth

import (
	"context"
	"errors"
	"portal-system/internal/platform/token"
	"portal-system/internal/repositories"
)

type Authenticator struct {
	manager     *token.Manager
	roleRepo    repositories.RoleRepository
	sessionRepo repositories.AuthSessionRepository
}

func NewAuthenticator(manager *token.Manager, roleRepo repositories.RoleRepository, sessionRepo repositories.AuthSessionRepository) *Authenticator {
	return &Authenticator{manager: manager, roleRepo: roleRepo, sessionRepo: sessionRepo}
}

func (a *Authenticator) Authenticate(ctx context.Context, tokenString string) (*Principal, error) {
	claims, err := a.manager.Parse(tokenString)
	if err != nil {
		return nil, err
	}

	session, err := a.sessionRepo.FindActiveByID(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != claims.UserID {

		// SECURITY FLAG
		return nil, errors.New("session does not belong to user")
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
		RoleCode:    role.Code,
		SessionID:   claims.SessionID,
		Permissions: perms,
	}

	return principal, nil

}
