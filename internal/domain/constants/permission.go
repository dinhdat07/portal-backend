package constants

type PermissionCode string

const (
	// profile
	PermProfileReadSelf       PermissionCode = "profile:read_self"
	PermProfileUpdateSelf     PermissionCode = "profile:update_self"
	PermProfileChangePassword PermissionCode = "profile:change_password"

	// user management
	PermUserList       PermissionCode = "users:list"
	PermUserReadDetail PermissionCode = "users:read_detail"
	PermUserCreate     PermissionCode = "users:create"
	PermUserUpdate     PermissionCode = "users:update"
	PermUserDelete     PermissionCode = "users:delete"
	PermUserRestore    PermissionCode = "users:restore"
	PermUserRoleUpdate PermissionCode = "users:update_role"
)

var AllPermissions = []PermissionCode{
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
}
