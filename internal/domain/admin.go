package domain

import (
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"time"

	"github.com/google/uuid"
)

type UsersFilter struct {
	Page     int
	PageSize int
	Username string
	Email    string
	FullName string
	Dob      *time.Time

	// input API
	RoleCode *constants.RoleCode

	// internal (service fill)
	RoleID *uuid.UUID

	Status         enum.UserStatus
	IncludeDeleted bool
}

type ListUsersResult struct {
	Users    []models.User
	Total    int64
	Page     int
	PageSize int
}

type CreateUserInput struct {
	Email     string
	Username  string
	FirstName string
	LastName  string
	DOB       *time.Time
	RoleCode  constants.RoleCode
}

type UpdateUserInput struct {
	Username  *string
	FirstName *string
	LastName  *string
	DOB       *time.Time
}
