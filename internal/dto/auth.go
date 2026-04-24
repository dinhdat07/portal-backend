package dto

import (
	"portal-system/internal/types"
)

type RegisterRequest struct {
	Email     string         `json:"email" binding:"required,email"`
	Username  string         `json:"username" binding:"required,min=3,max=50"`
	FirstName string         `json:"first_name" binding:"required,max=100"`
	LastName  string         `json:"last_name" binding:"required,max=100"`
	Password  string         `json:"password" binding:"required,min=8,max=255"`
	Dob       types.DateOnly `json:"dob" binding:"required"`
}

type AuthMessageResponse struct {
	Message string `json:"message"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         UserResponse `json:"user"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type VerifyEmailResponse struct {
	Message            string `json:"message"`
	RequirePasswordSet bool   `json:"require_password_set"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type SetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"password" binding:"required,min=8,max=255"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8,max=255"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}
