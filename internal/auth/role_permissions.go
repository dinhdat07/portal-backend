package auth

import "portal-system/internal/domain/enum"

var rolePermissions = map[enum.UserRole]map[Permission]struct{}{
	enum.RoleAdmin: {
		PermUserList:              {},
		PermUserReadDetail:        {},
		PermUserCreate:            {},
		PermUserUpdate:            {},
		PermUserDelete:            {},
		PermUserRestore:           {},
		PermUserRoleUpdate:        {},
		PermProfileReadSelf:       {},
		PermProfileUpdateSelf:     {},
		PermProfileChangePassword: {},
	},
	enum.RoleUser: {
		PermProfileReadSelf:       {},
		PermProfileUpdateSelf:     {},
		PermProfileChangePassword: {},
	},
}
