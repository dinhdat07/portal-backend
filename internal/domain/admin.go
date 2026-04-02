package domain

import (
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"time"
)

type UsersFilter struct {
	Page           int
	PageSize       int
	Username       string
	Email          string
	FullName       string
	Dob            *time.Time
	Role           enum.UserRole
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
	Password  string
	DOB       *time.Time
	Role      enum.UserRole
}

type UpdateUserInput struct {
	Username  *string
	FirstName *string
	LastName  *string
	DOB       *time.Time
}
