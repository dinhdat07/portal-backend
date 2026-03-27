package services

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService struct {
	userRepo *repositories.UserRepository
}

func NewAdminService(repo *repositories.UserRepository) *AdminService {
	return &AdminService{userRepo: repo}
}

func (svc *AdminService) ListUsers(ctx context.Context, in domain.ListUsersInput) (*domain.ListUsersResult, error) {
	if in.Role != "" && !in.Role.IsValid() {
		return nil, ErrInvalidInput
	}

	if in.Status != "" && !in.Status.IsValid() {
		return nil, ErrInvalidInput
	}

	users, total, err := svc.userRepo.ListUsers(ctx, in)
	if err != nil {
		return nil, ErrInternalServer
	}

	return &domain.ListUsersResult{
		Users:    users,
		Total:    total,
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil

}

func (svc *AdminService) CreateUser(ctx context.Context, in domain.CreateUserInput) (*models.User, error) {
	if in.Role != "" && !in.Role.IsValid() {
		return nil, ErrInvalidInput
	}

	user := &models.User{
		Email:        in.Email,
		Username:     in.Username,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		DOB:          in.DOB,
		PasswordHash: nil,
		Role:         "user",
		Status:       models.StatusPending,
	}

	err := svc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, ErrInternalServer
	}

	return user, nil
}

func (svc *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := svc.userRepo.Delete(ctx, userID, adminID); err != nil {
		return nil, ErrInternalServer
	}

	now := time.Now()
	user.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	user.DeletedBy = &adminID
	user.Status = models.StatusDeleted

	return user, nil
}

func (svc *AdminService) RestoreUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByIDUnscoped(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.DeletedAt.Time.IsZero() {
		return nil, ErrUserNotDeleted
	}

	if err := svc.userRepo.Restore(ctx, userID); err != nil {
		return nil, ErrInternalServer
	}

	user.DeletedAt = gorm.DeletedAt{}
	user.DeletedBy = nil
	user.Status = models.StatusActive

	return user, nil
}

func (svc *AdminService) UpdateRole(ctx context.Context, id uuid.UUID, role models.UserRole) (*models.User, error) {
	if !role.IsValid() {
		return nil, ErrInvalidInput
	}

	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.Role == role {
		return user, nil
	}

	if err := svc.userRepo.UpdateRole(ctx, id, role); err != nil {
		return nil, ErrInternalServer
	}

	user.Role = role

	return user, nil
}
