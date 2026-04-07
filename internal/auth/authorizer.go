package auth

import (
	"context"
	"portal-system/internal/domain/constants"
)

type Authorizer struct {
}

func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

func (a *Authorizer) HasPermission(ctx context.Context, principal *Principal, permission constants.PermissionCode) bool {
	for _, p := range principal.Permissions {
		if p == string(permission) {
			return true
		}
	}
	return false
}
