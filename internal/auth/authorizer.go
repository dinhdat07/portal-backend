package auth

type Authorizer struct {
}

func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

func (a *Authorizer) HasPermission(principal *Principal, permission Permission) bool {
	if principal == nil {
		return false
	}

	role := principal.Role
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}

	_, ok = perms[permission]
	return ok
}
