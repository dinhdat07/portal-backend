package dto

import "portal-system/internal/types"

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
