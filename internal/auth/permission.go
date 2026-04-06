package auth

type Permission string

const (
	PermProfileReadSelf       Permission = "profile.read_self"
	PermProfileUpdateSelf     Permission = "profile.update_self"
	PermProfileChangePassword Permission = "profile.change_password"

	PermAuthLogout Permission = "auth.logout"

	PermUserList       Permission = "user.list"
	PermUserReadDetail Permission = "user.read_detail"
	PermUserCreate     Permission = "user.create"
	PermUserUpdate     Permission = "user.update"
	PermUserDelete     Permission = "user.delete"
	PermUserRestore    Permission = "user.restore"
	PermUserRoleUpdate Permission = "user.role.update"
)
