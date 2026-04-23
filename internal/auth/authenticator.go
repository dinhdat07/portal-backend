package auth

import (
	"context"
	"errors"
	appLogger "log"
	"portal-system/internal/platform/token"
	"portal-system/internal/repositories"
	"portal-system/internal/services"
)

type Authenticator struct {
	manager     *token.Manager
	roleRepo    repositories.RoleRepository
	sessionRepo repositories.AuthSessionRepository
	revoStore   services.SessionRevocationStore
}

func NewAuthenticator(manager *token.Manager, roleRepo repositories.RoleRepository, sessionRepo repositories.AuthSessionRepository, revoStore services.SessionRevocationStore) *Authenticator {
	return &Authenticator{manager: manager, roleRepo: roleRepo, sessionRepo: sessionRepo, revoStore: revoStore}
}

func (a *Authenticator) Authenticate(ctx context.Context, tokenString string) (*Principal, error) {
	claims, err := a.manager.Parse(tokenString)
	if err != nil {
		return nil, err
	}

	revoked, err := a.revoStore.IsRevoked(ctx, claims.SessionID)
	if err != nil {
		appLogger.Println(err)
	}

	if revoked {
		return nil, errors.New("session is already revoked")
	}

	session, err := a.sessionRepo.FindActiveByID(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != claims.UserID {
		// SHOULD ADD SECURITY LOG ?
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
