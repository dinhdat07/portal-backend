package storage

import (
	"context"
	"portal-system/internal/domain/constants"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormRoleRepository struct {
	db *gorm.DB
}

func NewGormRoleRepository(db *gorm.DB) repositories.RoleRepository {
	return &GormRoleRepository{db: db}
}

func (r *GormRoleRepository) FindByCode(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
	var role models.Role

	err := r.getDB(ctx).
		Where("code = ?", string(code)).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *GormRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role

	if err := r.getDB(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *GormRoleRepository) List(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role

	err := r.getDB(ctx).Preload("Permissions").Order("name ASC").Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *GormRoleRepository) GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	var role models.Role

	err := r.getDB(ctx).
		Preload("Permissions").
		First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *GormRoleRepository) AssignPermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	rp := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permID,
	}

	return r.getDB(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		FirstOrCreate(rp).Error
}

func (r *GormRoleRepository) RemovePermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return r.getDB(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		Delete(&models.RolePermission{}).Error
}

func (r *GormRoleRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
