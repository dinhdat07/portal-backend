package enum

type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
	TokenTypePasswordSet       TokenType = "password_set" // user created by admin set password first time
)

func (t TokenType) IsValid() bool {
	switch t {
	case TokenTypePasswordReset,
		TokenTypeEmailVerification,
		TokenTypePasswordSet:
		return true
	default:
		return false
	}
}
