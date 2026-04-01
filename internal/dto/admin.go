package dto

import (
	"portal-system/internal/domain/enum"
	"portal-system/internal/types"
)

type ListUsersQuery struct {
	Page           int             `form:"page" binding:"omitempty,min=1"`
	PageSize       int             `form:"page_size" binding:"omitempty,min=1,max=100"`
	Username       string          `form:"username"`
	Email          string          `form:"email"`
	FullName       string          `form:"full_name"`
	Dob            *types.DateOnly `form:"dob"`
	Role           string          `form:"role"`
	Status         string          `form:"status"`
	IncludeDeleted *bool           `form:"include_deleted"`
}

type CreateUserRequest struct {
	Email     string          `json:"email" binding:"required" format:"email"`
	Username  string          `json:"username" binding:"required,min=3,max=50"`
	FirstName string          `json:"first_name" binding:"required,max=100"`
	LastName  string          `json:"last_name" binding:"required,max=100"`
	DOB       *types.DateOnly `json:"dob,omitempty" binding:"required"`
	Role      enum.UserRole   `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	Username  *string         `json:"username,omitempty" binding:"min=3,max=50"`
	FirstName *string         `json:"first_name,omitempty" binding:"max=100"`
	LastName  *string         `json:"last_name,omitempty" binding:"max=100"`
	DOB       *types.DateOnly `json:"dob,omitempty"`
}

type UpdateRoleRequest struct {
	Role enum.UserRole `json:"role" binding:"required"`
}
