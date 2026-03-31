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
	db          *gorm.DB
	auditLogger *AuditLogService
	userRepo    *repositories.UserRepository
}

func NewAdminService(db *gorm.DB, repo *repositories.UserRepository, logger *AuditLogService) *AdminService {
	return &AdminService{db: db, userRepo: repo, auditLogger: logger}
}

func (svc *AdminService) ListUsers(ctx context.Context, meta *domain.AuditMeta, actor *models.User, in domain.UsersFilter) (*domain.ListUsersResult, error) {
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

	logMeta := map[string]any{
		"filters": map[string]any{
			"username":        in.Username,
			"email":           in.Email,
			"full_name":       in.FullName,
			"dob":             in.Dob,
			"role":            in.Role,
			"status":          in.Status,
			"include_deleted": in.IncludeDeleted,
		},
		"pagination": map[string]any{
			"page":      in.Page,
			"page_size": in.PageSize,
		},
		"result_count": len(users),
		"total":        total,
	}

	svc.auditLogger.LogWithMetadata(
		ctx,
		meta,
		models.ActionAdminSearchUser,
		actor,
		nil,
		logMeta,
	)

	return &domain.ListUsersResult{
		Users:    users,
		Total:    total,
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil

}

func (svc *AdminService) CreateUser(ctx context.Context, meta *domain.AuditMeta, actor *models.User, in domain.CreateUserInput) (*models.User, error) {
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

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := svc.userRepo.Create(ctx, user)
		if err != nil {
			return ErrInternalServer
		}

		if err := svc.auditLogger.Log(ctx, meta, models.ActionAdminCreateUser, actor, user); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	return user, nil
}

func (svc *AdminService) DeleteUser(ctx context.Context, meta *domain.AuditMeta, actor *models.User, userID uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.Delete(ctx, userID, actor.ID); err != nil {
			return ErrInternalServer
		}

		now := time.Now()
		user.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
		user.DeletedBy = &actor.ID
		user.Status = models.StatusDeleted

		if err := svc.auditLogger.Log(ctx, meta, models.ActionAdminDeleteUser, actor, user); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *AdminService) RestoreUser(ctx context.Context, meta *domain.AuditMeta, actor *models.User, userID uuid.UUID) (*models.User, error) {
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

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.Restore(ctx, userID); err != nil {
			return ErrInternalServer
		}

		user.DeletedAt = gorm.DeletedAt{}
		user.DeletedBy = nil
		user.Status = models.StatusActive

		if err := svc.auditLogger.Log(ctx, meta, models.ActionAdminRestoreUser, actor, user); err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *AdminService) UpdateRole(ctx context.Context, meta *domain.AuditMeta, actor *models.User, id uuid.UUID, role models.UserRole) (*models.User, error) {
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

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.UpdateRole(ctx, id, role); err != nil {
			return ErrInternalServer
		}
		user.Role = role

		if err := svc.auditLogger.Log(ctx, meta, models.ActionAdminAssignRole, actor, user); err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}
