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

	// Registration
	ErrEmailExists      = errors.New("email already exists")
	ErrUsernameExists   = errors.New("username already exists")
	ErrEmailBlacklisted = errors.New("email is blacklisted")

	// Password
	ErrIncorrectPassword            = errors.New("current password incorrect")
	ErrPasswordConfirmationMismatch = errors.New("password confirmation does not match")
	ErrNewPasswordMustBeDifferent   = errors.New("new password must be different from current password")
	ErrPasswordTooWeak              = errors.New("password is too weak")

	// Validation
	ErrInvalidInput = errors.New("invalid input")

	// System
	ErrInternalServer = errors.New("internal server error")
)
