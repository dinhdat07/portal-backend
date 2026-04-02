package services

import "errors"

var (
	// Auth
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")

	// Account state
	ErrAccountNotVerified = errors.New("account not verified")
	ErrAccountDeleted     = errors.New("account deleted")
	ErrUserInactive       = errors.New("user is inactive")

	// User
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyDeleted = errors.New("user already deleted")
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrUserNotDeleted     = errors.New("user is not deleted")

	// Registration
	ErrEmailExists      = errors.New("email already exists")
	ErrUsernameExists   = errors.New("username already exists")
	ErrEmailBlacklisted = errors.New("email is blacklisted")

	// Password
	ErrIncorrectPassword            = errors.New("current password incorrect")
	ErrPasswordConfirmationMismatch = errors.New("password confirmation does not match")
	ErrNewPasswordMustBeDifferent   = errors.New("new password must be different from current password")
	ErrPasswordAlreadySet           = errors.New("password already set")

	// Validation
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidAction    = errors.New("invalid action for audit log")
	ErrInvalidTimeRange = errors.New("invalid time range for audit log")

	// System
	ErrInternalServer         = errors.New("internal server error")
	ErrInvalidToken           = errors.New("token is not valid")
	ErrAuditLogger            = errors.New("cannot log this action")
	ErrSendVerificationEmail  = errors.New("cannot send verification email")
	ErrSendResetPasswordEmail = errors.New("cannot send reset password email")
	ErrSendSetPasswordEmail   = errors.New("cannot send set password email")
)
