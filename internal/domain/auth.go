package domain

import (
	"portal-system/internal/models"
)

type LoginResult struct {
	AccessToken string
	ExpiresIn   int
	User        *models.User
}
