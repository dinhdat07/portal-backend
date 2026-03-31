package models

type ActionName string

const (
	ActionRegister    ActionName = "REGISTER"
	ActionVerifyEmail ActionName = "VERIFY_EMAIL"
	ActionLogin       ActionName = "LOGIN"
	ActionLogout      ActionName = "LOGOUT"

	ActionUpdateProfile  ActionName = "UPDATE_PROFILE"
	ActionChangePassword ActionName = "CHANGE_PASSWORD"

	ActionAdminSearchUser  ActionName = "ADMIN_SEARCH_USER"
	ActionAdminViewUser    ActionName = "ADMIN_VIEW_USER"
	ActionAdminCreateUser  ActionName = "ADMIN_CREATE_USER"
	ActionAdminUpdateUser  ActionName = "ADMIN_UPDATE_USER"
	ActionAdminDeleteUser  ActionName = "ADMIN_DELETE_USER"
	ActionAdminRestoreUser ActionName = "ADMIN_RESTORE_USER"
	ActionAdminAssignRole  ActionName = "ADMIN_ASSIGN_ROLE"
)

func (a ActionName) IsValid() bool {
	switch a {
	case ActionRegister,
		ActionVerifyEmail,
		ActionLogin,
		ActionLogout,
		ActionUpdateProfile,
		ActionChangePassword,
		ActionAdminSearchUser,
		ActionAdminViewUser,
		ActionAdminCreateUser,
		ActionAdminUpdateUser,
		ActionAdminDeleteUser,
		ActionAdminRestoreUser,
		ActionAdminAssignRole:
		return true
	default:
		return false
	}
}
