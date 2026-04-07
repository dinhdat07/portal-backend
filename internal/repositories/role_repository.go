package repositories

import (
	"context"
	"portal-system/internal/domain/constants"
	"portal-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByCode(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
	var role models.Role

	err := r.db.WithContext(ctx).
		Where("code = ?", string(code)).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role

	if err := r.db.WithContext(ctx).Where("id =", id).First(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepository) List(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role

	err := r.db.WithContext(ctx).Preload("Permissions").Order("name ASC").Find(roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *RoleRepository) GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	var role models.Role

	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepository) AssignPermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	rp := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permID,
	}

	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		FirstOrCreate(rp).Error
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		Delete(&models.RolePermission{}).Error
}

func (r *RoleRepository) WithTx(tx *gorm.DB) *RoleRepository {
	return NewRoleRepository(tx)
}
