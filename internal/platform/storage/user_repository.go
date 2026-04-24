package storage

import (
	"context"
	"errors"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *models.User) error {
	return r.getDB(ctx).Create(user).Error
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.getDB(ctx).Preload("Role").First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.getDB(ctx).Preload("Role").Unscoped().First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.getDB(ctx).Preload("Role").
		Where("username = ?", username).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.getDB(ctx).Preload("Role").
		Where("email = ?", email).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (r *GormUserRepository) Update(ctx context.Context, user *models.User) error {
	return r.getDB(ctx).
		Model(user).
		Updates(map[string]interface{}{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"dob":        user.DOB,
			"username":   user.Username,
		}).Error
}

func (r *GormUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	var user models.User
	result := r.getDB(ctx).
		Model(&user).
		Where("id = ?", id).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) UpdateRole(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("role_id", roleID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"email_verified_at": &now,
			"status":            enum.StatusActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
			"status":     enum.StatusDeleted,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&models.User{}).
		Unscoped().
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"deleted_by": nil,
			"status":     enum.StatusActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) ListUsers(ctx context.Context, filter domain.UsersFilter) ([]models.User, int64, error) {
	var user models.User

	db := r.getDB(ctx).Model(&user)

	// build dynamic query
	if filter.IncludeDeleted {
		db = db.Unscoped()
	}

	if filter.Username != "" {
		db = db.Where("username ILIKE ?", "%"+filter.Username+"%")
	}

	if filter.Email != "" {
		db = db.Where("email ILIKE ?", "%"+filter.Email+"%")
	}

	if filter.FullName != "" {
		db = db.Where(
			"CONCAT(first_name, ' ', last_name) ILIKE ?",
			"%"+filter.FullName+"%",
		)
	}

	if filter.Dob != nil {
		db = db.Where("dob = ?", *filter.Dob)
	}

	if filter.RoleID != nil {
		db = db.Where("role_id = ?", *filter.RoleID)
	}

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	var users []models.User

	if err := db.Preload("Role").Offset(offset).Limit(filter.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *GormUserRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
