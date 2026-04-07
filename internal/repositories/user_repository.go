package repositories

import (
	"context"
	"errors"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	var user models.User
	return r.db.WithContext(ctx).
		Model(&user).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).
		Model(user).
		Updates(map[string]interface{}{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"dob":        user.DOB,
			"username":   user.Username,
		}).Error
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Role").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Role").Unscoped().First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Role").
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

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Role").
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

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
			"status":     enum.StatusDeleted,
		}).Error
}

func (r *UserRepository) ListUsers(ctx context.Context, filter domain.UsersFilter) ([]models.User, int64, error) {
	var user models.User

	db := r.db.WithContext(ctx).Model(&user)

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

	if err := db.Offset(offset).Limit(filter.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Unscoped().
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"deleted_by": nil,
			"status":     enum.StatusActive,
		}).Error
}

func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("role_id", roleID).Error
}

func (r *UserRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"email_verified_at": &now,
			"status":            enum.StatusActive,
		}).Error
}

func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	return NewUserRepository(tx)
}
