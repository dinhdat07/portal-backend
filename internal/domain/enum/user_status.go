package enum

type UserStatus string

const (
	StatusActive  UserStatus = "active"
	StatusPending UserStatus = "pending_verification"
	StatusDeleted UserStatus = "deleted"
)

func (r UserStatus) IsValid() bool {
	switch r {
	case StatusActive, StatusDeleted, StatusPending:
		return true
	default:
		return false
	}
}
