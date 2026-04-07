package services

import (
	"context"
	"portal-system/internal/domain/constants"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleService struct {
	repo *repositories.RoleRepository
}

func NewRoleService(repo *repositories.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) ListRoles(ctx context.Context) ([]models.Role, error) {
	roles, err := s.repo.List(ctx)
	if err != nil {
		return nil, ErrInternalServer
	}

	return roles, nil
}

func (s *RoleService) GetRoleWithPermissions(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	role, err := s.repo.GetWithPermissions(ctx, roleID)
	if err != nil {
		return nil, ErrInternalServer
	}

	return role, nil
}

func (s *RoleService) AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := s.repo.AssignPermission(ctx, roleID, permissionID)
	if err != nil {
		return ErrInternalServer
	}
	return nil
}

func (s *RoleService) RemovePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := s.repo.RemovePermission(ctx, roleID, permissionID)
	if err != nil {
		return ErrInternalServer
	}
	return nil
}

func (s *RoleService) FindRoleByCode(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
	role, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, ErrInternalServer
	}

	return role, nil
}

func (s *RoleService) WithTx(tx *gorm.DB) *RoleService {
	return &RoleService{
		repo: s.repo.WithTx(tx),
	}
}
