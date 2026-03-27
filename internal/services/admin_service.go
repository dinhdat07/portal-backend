package services

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/repositories"
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
