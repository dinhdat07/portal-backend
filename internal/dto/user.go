package dto

import (
	"portal-system/internal/models"
	"time"
)

type UserResponse struct {
	ID              string            `json:"id"`
	Email           string            `json:"email"`
	Username        string            `json:"username"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	DOB             *time.Time        `json:"dob,omitempty"`
	Role            models.UserRole   `json:"role"`
	Status          models.UserStatus `json:"status"`
	EmailVerifiedAt *time.Time        `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time        `json:"last_login_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       *time.Time        `json:"deleted_at,omitempty"`
	DeletedBy       *string           `json:"deleted_by,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=255"`
	ConfirmPassword string `json:"confirm_new_password" binding:"required,min=8,max=255"`
}

func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
