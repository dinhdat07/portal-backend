package repositories

import (
	"context"
	"time"

	"portal-system/internal/domain"
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
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Unscoped().First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error
	return &user, err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
			"status":     models.StatusDeleted,
		}).Error
}

func (r *UserRepository) ListUsers(ctx context.Context, in domain.ListUsersInput) ([]models.User, int64, error) {
	var user models.User

	db := r.db.WithContext(ctx).Model(&user)

	// build dynamic query
	if in.IncludeDeleted {
		db = db.Unscoped()
	}

	if in.Username != "" {
		db = db.Where("username ILIKE ?", "%"+in.Username+"%")
	}

	if in.Email != "" {
		db = db.Where("email ILIKE ?", "%"+in.Email+"%")
	}

	if in.FullName != "" {
		db = db.Where(
			"CONCAT(first_name, ' ', last_name) ILIKE ?",
			"%"+in.FullName+"%",
		)
	}

	if in.Dob != nil {
		db = db.Where("dob = ?", *in.Dob)
	}

	if in.Role != "" {
		db = db.Where("role = ?", in.Role)
	}

	if in.Status != "" {
		db = db.Where("status = ?", in.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (in.Page - 1) * in.PageSize
	var users []models.User

	if err := db.Offset(offset).Limit(in.PageSize).Find(&users).Error; err != nil {
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
			"status":     models.StatusActive,
		}).Error
}

func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role models.UserRole) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("role", role).Error
}
