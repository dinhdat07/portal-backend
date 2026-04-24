package constants

var RolePermissions = map[RoleCode][]PermissionCode{
	RoleCodeAdmin: {
		PermUserList,
		PermUserReadDetail,
		PermUserCreate,
		PermUserUpdate,
		PermUserDelete,
		PermUserRestore,
		PermUserRoleUpdate,
		PermProfileReadSelf,
		PermProfileUpdateSelf,
		PermProfileChangePassword,
	},
	RoleCodeUser: {
		PermProfileReadSelf,
		PermProfileUpdateSelf,
		PermProfileChangePassword,
	},
}
