package domain

import (
	"portal-system/internal/models"
	"time"
)

type ListUsersInput struct {
	Page           int
	PageSize       int
	Username       string
	Email          string
	FullName       string
	Dob            *time.Time
	Role           models.UserRole
	Status         models.UserStatus
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
	Role      models.UserRole
}

type UpdateUserInput struct {
	Username  *string
	FirstName *string
	LastName  *string
	DOB       *time.Time
}
