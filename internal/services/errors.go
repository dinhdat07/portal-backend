package services

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrInvalidCredentials = &AppError{
		Code:    "invalid_credentials",
		Message: "Email/username or password is incorrect",
	}

	ErrUnauthorized = &AppError{
		Code:    "unauthorized",
		Message: "You are not authenticated",
	}

	ErrForbidden = &AppError{
		Code:    "forbidden",
		Message: "You do not have permission to perform this action",
	}
)

var (
	ErrAccountNotVerified = &AppError{
		Code:    "account_not_verified",
		Message: "Your account is not verified. Please check your email",
	}

	ErrAccountDeleted = &AppError{
		Code:    "account_deleted",
		Message: "This account has been deleted",
	}

	ErrUserInactive = &AppError{
		Code:    "user_inactive",
		Message: "Your account is inactive",
	}

	ErrInvalidRefreshToken = &AppError{
		Code:    "invalid_refresh_token",
		Message: "Your session is expired",
	}
)

var (
	ErrUserNotFound = &AppError{
		Code:    "user_not_found",
		Message: "User not found",
	}

	ErrUserAlreadyDeleted = &AppError{
		Code:    "user_already_deleted",
		Message: "User is already deleted",
	}

	ErrInvalidUserID = &AppError{
		Code:    "invalid_user_id",
		Message: "Invalid user ID",
	}

	ErrUserNotDeleted = &AppError{
		Code:    "user_not_deleted",
		Message: "User is not deleted",
	}
)

var (
	ErrEmailExists = &AppError{
		Code:    "email_already_exists",
		Message: "Email is already in use",
	}

	ErrUsernameExists = &AppError{
		Code:    "username_already_exists",
		Message: "Username is already taken",
	}

	ErrEmailBlacklisted = &AppError{
		Code:    "email_blacklisted",
		Message: "This email cannot be used",
	}
)

var (
	ErrIncorrectPassword = &AppError{
		Code:    "incorrect_password",
		Message: "Current password is incorrect",
	}

	ErrPasswordConfirmationMismatch = &AppError{
		Code:    "password_mismatch",
		Message: "Password confirmation does not match",
	}

	ErrNewPasswordMustBeDifferent = &AppError{
		Code:    "password_not_changed",
		Message: "New password must be different from the current one",
	}

	ErrPasswordAlreadySet = &AppError{
		Code:    "password_already_set",
		Message: "Password has already been set",
	}
)

var (
	ErrInvalidInput = &AppError{
		Code:    "invalid_input",
		Message: "Invalid input data",
	}

	ErrInvalidAction = &AppError{
		Code:    "invalid_action",
		Message: "Invalid action",
	}

	ErrInvalidTimeRange = &AppError{
		Code:    "invalid_time_range",
		Message: "Invalid time range",
	}
)

var (
	ErrInternalServer = &AppError{
		Code:    "internal_error",
		Message: "Something went wrong. Please try again later",
	}

	ErrInvalidToken = &AppError{
		Code:    "invalid_token",
		Message: "Invalid or expired token",
	}

	ErrAuditLogger = &AppError{
		Code:    "audit_log_failed",
		Message: "Failed to record activity",
	}

	ErrSendVerificationEmail = &AppError{
		Code:    "send_verification_failed",
		Message: "Failed to send verification email",
	}

	ErrSendResetPasswordEmail = &AppError{
		Code:    "send_reset_password_failed",
		Message: "Failed to send reset password email",
	}

	ErrSendSetPasswordEmail = &AppError{
		Code:    "send_set_password_failed",
		Message: "Failed to send set password email",
	}
)
