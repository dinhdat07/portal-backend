package constants

type RoleCode string

const (
	RoleCodeUser  RoleCode = "user"
	RoleCodeAdmin RoleCode = "admin"
)

func (r RoleCode) IsValid() bool {
	switch r {
	case RoleCodeUser, RoleCodeAdmin:
		return true
	default:
		return false
	}
}
