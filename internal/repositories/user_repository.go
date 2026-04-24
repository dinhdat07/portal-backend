package repositories

import (
	"context"

	"portal-system/internal/domain"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateRole(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, filter domain.UsersFilter) ([]models.User, int64, error)
}
