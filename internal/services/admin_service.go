package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/platform/email"
	"portal-system/internal/repositories"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService struct {
	db          *gorm.DB
	auditLogger *AuditLogService
	userRepo    *repositories.UserRepository
	tokenRepo   *repositories.UserTokenRepository
	roleRepo    *repositories.RoleRepository
	emailSvc    *email.SMTPEmailService
	frontendURL string
}

func NewAdminService(
	db *gorm.DB,
	repo *repositories.UserRepository,
	tokenRepo *repositories.UserTokenRepository,
	roleRepo *repositories.RoleRepository,
	logger *AuditLogService,
	emailSvc *email.SMTPEmailService,
	frontendURL string,
) *AdminService {
	return &AdminService{
		db:          db,
		userRepo:    repo,
		tokenRepo:   tokenRepo,
		roleRepo:    roleRepo,
		auditLogger: logger,
		emailSvc:    emailSvc,
		frontendURL: frontendURL,
	}
}

func (svc *AdminService) ListUsers(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, in domain.UsersFilter) (*domain.ListUsersResult, error) {
	if in.RoleCode != nil {
		role, err := svc.roleRepo.FindByCode(ctx, *in.RoleCode)
		if err != nil {
			return nil, ErrInvalidInput
		}
		in.RoleID = &role.ID
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
			"role_code":       in.RoleCode,
			"role_id":         in.RoleID,
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
		enum.ActionAdminSearchUser,
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

func (svc *AdminService) CreateUser(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, in domain.CreateUserInput) (*models.User, error) {
	if in.RoleCode == "" {
		return nil, ErrInvalidInput
	}

	existingByEmail, _ := svc.userRepo.FindByEmail(ctx, in.Email)
	if existingByEmail != nil && existingByEmail.ID != uuid.Nil {
		return nil, ErrEmailExists
	}

	existingByUsername, err := svc.userRepo.FindByUsername(ctx, in.Username)
	if err != nil {
		return nil, ErrInternalServer
	}
	if existingByUsername != nil && existingByUsername.ID != uuid.Nil {
		return nil, ErrUsernameExists
	}

	tokenHash, rawToken, err := generateHashToken()
	if err != nil {
		return nil, err
	}

	role, err := svc.roleRepo.FindByCode(ctx, in.RoleCode)
	if role == nil || err != nil {
		return nil, ErrInternalServer
	}

	user := &models.User{
		Email:     in.Email,
		Username:  in.Username,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		DOB:       in.DOB,
		RoleID:    role.ID,
		Role:      *role,
		Status:    enum.StatusPending,
	}

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.WithTx(tx).Create(ctx, user); err != nil {
			return ErrInternalServer
		}

		if err := svc.tokenRepo.WithTx(tx).
			RevokeByUserAndType(ctx, user.ID, enum.TokenTypePasswordSet); err != nil {
			return ErrInternalServer
		}

		setPasswordToken := &models.UserToken{
			UserID:    user.ID,
			TokenType: enum.TokenTypePasswordSet,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}

		if err := svc.tokenRepo.WithTx(tx).Create(ctx, setPasswordToken); err != nil {
			return ErrInternalServer
		}

		setPasswordURL := fmt.Sprintf("%s/set-password?token=%s", svc.frontendURL, url.QueryEscape(rawToken))

		if err := svc.emailSvc.SendSetPasswordEmail(ctx, user.Email, user.FirstName, setPasswordURL); err != nil {
			return ErrSendSetPasswordEmail
		}

		target := domain.MapUserToAuditUser(user)
		if err := svc.auditLogger.WithTx(tx).Log(ctx, meta, enum.ActionAdminCreateUser, actor, target); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *AdminService) DeleteUser(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, userID uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	roleAdmin, err := svc.roleRepo.FindByCode(ctx, constants.RoleCodeAdmin)
	if err != nil {
		return nil, ErrInternalServer
	}

	if actor.ID != user.ID && user.RoleID == roleAdmin.ID {
		return nil, ErrForbidden
	}

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.WithTx(tx).Delete(ctx, userID, actor.ID); err != nil {
			return ErrInternalServer
		}

		now := time.Now()
		user.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
		user.DeletedBy = &actor.ID
		user.Status = enum.StatusDeleted

		target := domain.MapUserToAuditUser(user)
		if err := svc.auditLogger.WithTx(tx).Log(ctx, meta, enum.ActionAdminDeleteUser, actor, target); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *AdminService) RestoreUser(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, userID uuid.UUID) (*models.User, error) {
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
		if err := svc.userRepo.WithTx(tx).Restore(ctx, userID); err != nil {
			return ErrInternalServer
		}

		user.DeletedAt = gorm.DeletedAt{}
		user.DeletedBy = nil
		user.Status = enum.StatusActive

		target := domain.MapUserToAuditUser(user)
		if err := svc.auditLogger.WithTx(tx).Log(ctx, meta, enum.ActionAdminRestoreUser, actor, target); err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *AdminService) UpdateRole(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, id uuid.UUID, roleCode constants.RoleCode) (*models.User, error) {
	if roleCode == "" {
		return nil, ErrInvalidInput
	}

	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	role, err := svc.roleRepo.FindByCode(ctx, roleCode)
	if err != nil {
		return nil, ErrInternalServer
	}

	roleAdmin, err := svc.roleRepo.FindByCode(ctx, constants.RoleCodeAdmin)
	if err != nil {
		return nil, ErrInternalServer
	}

	if actor.ID != user.ID && user.RoleID == roleAdmin.ID {
		return nil, ErrForbidden
	}

	if user.RoleID == role.ID {
		return user, nil
	}

	changes := map[string]any{}
	changes["role"] = map[string]any{
		"old": user.Role,
		"new": role,
	}

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.WithTx(tx).UpdateRole(ctx, id, role.ID); err != nil {
			return ErrInternalServer
		}
		user.Role = *role

		target := domain.MapUserToAuditUser(user)

		if err := svc.auditLogger.WithTx(tx).LogWithMetadata(ctx, meta, enum.ActionAdminAssignRole, actor, target, changes); err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}
