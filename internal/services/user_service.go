package services

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	txManager   repositories.TxManager
	auditLogger *AuditLogService
	roleRepo    repositories.RoleRepository
	userRepo    repositories.UserRepository
}

func NewUserService(txManager repositories.TxManager, repo repositories.UserRepository, roleRepo repositories.RoleRepository, logger *AuditLogService) *UserService {
	return &UserService{txManager: txManager, userRepo: repo, roleRepo: roleRepo, auditLogger: logger}
}

func (svc *UserService) GetProfile(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, id uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if actor.RoleCode == constants.RoleCodeAdmin {
		target := domain.MapUserToAuditUser(user)
		svc.auditLogger.Log(ctx, meta, enum.ActionAdminViewUser, actor, target)
	}

	return user, nil
}

func (svc *UserService) ChangePassword(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, current, newPassword, confirm string) error {
	user, err := svc.userRepo.FindByID(ctx, actor.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnauthorized
		}
		return err
	}

	if strings.TrimSpace(newPassword) == "" ||
		strings.TrimSpace(confirm) == "" {
		return ErrInvalidInput
	}

	//check nil before compare to avoid panic
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return ErrUnauthorized
	}

	if newPassword != confirm {
		return ErrPasswordConfirmationMismatch
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(current)); err != nil {
		return ErrIncorrectPassword
	}

	if current == newPassword {
		return ErrNewPasswordMustBeDifferent
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.UpdatePassword(ctx, actor.ID, string(hashed)); err != nil {
			return ErrInternalServer
		}

		if err := svc.auditLogger.Log(ctx, meta, enum.ActionChangePassword, actor, actor); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	return err

}

func (svc *UserService) UpdateProfile(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, id uuid.UUID, input domain.UpdateUserInput) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
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

	changes := map[string]any{}

	// update allowed fields
	if input.FirstName != nil {
		changes["first_name"] = map[string]any{
			"old": user.FirstName,
			"new": *input.FirstName,
		}
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		changes["last_name"] = map[string]any{
			"old": user.LastName,
			"new": *input.LastName,
		}
		user.LastName = *input.LastName
	}
	if input.DOB != nil {
		changes["dob"] = map[string]any{
			"old": user.DOB,
			"new": input.DOB,
		}
		user.DOB = input.DOB
	}

	// check duplicate username
	if input.Username != nil && *input.Username != user.Username {
		existing, err := svc.userRepo.FindByUsername(ctx, *input.Username)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrUsernameExists
		}
		changes["username"] = map[string]any{
			"old": user.Username,
			"new": *input.Username,
		}
		user.Username = *input.Username
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.Update(ctx, user); err != nil {
			return ErrInternalServer
		}

		action := enum.ActionUpdateProfile
		if actor.RoleCode == constants.RoleCodeAdmin {
			action = enum.ActionAdminUpdateUser
		}

		target := domain.MapUserToAuditUser(user)
		err := svc.auditLogger.LogWithMetadata(ctx, meta, action, actor, target, map[string]any{
			"changes": changes,
		})
		if err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}
