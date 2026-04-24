package constants

type RoleCode string

const (
	RoleCodeUser  RoleCode = "ROLE_CODE_USER"
	RoleCodeAdmin RoleCode = "ROLE_CODE_ADMIN"
)

func (r RoleCode) IsValid() bool {
	switch r {
	case RoleCodeUser, RoleCodeAdmin:
		return true
	default:
		return false
	}
}
