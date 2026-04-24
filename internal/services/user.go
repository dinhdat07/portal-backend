package services

import (
	"context"
	"errors"
	appLogger "log"
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

type UserService interface {
	GetProfile(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, id uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, id uuid.UUID, input domain.UpdateUserInput) (*models.User, error)
	ChangePassword(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, current, newPassword, confirm string) error
}

type userService struct {
	txManager   repositories.TxManager
	auditLogger AuditLogger
	roleRepo    repositories.RoleRepository
	userRepo    repositories.UserRepository
}

type UserServiceDeps struct {
	TxManager   repositories.TxManager
	AuditLogger AuditLogger
	RoleRepo    repositories.RoleRepository
	UserRepo    repositories.UserRepository
}

func NewUserService(deps UserServiceDeps) *userService {
	return &userService{
		txManager:   deps.TxManager,
		userRepo:    deps.UserRepo,
		roleRepo:    deps.RoleRepo,
		auditLogger: deps.AuditLogger,
	}
}

func (svc *userService) GetProfile(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, id uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if actor.RoleCode == constants.RoleCodeAdmin {
		target := domain.MapUserToAuditUser(user)
		if err := svc.auditLogger.Log(ctx, meta, enum.ActionAdminViewUser, actor, target); err != nil {
			appLogger.Println("failed to log admin view user action", "error", err)
		}
	}

	return user, nil
}

func (svc *userService) ChangePassword(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, current, newPassword, confirm string) error {
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

	// check nil before compare to avoid panic
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

func (svc *userService) UpdateProfile(ctx context.Context, meta *domain.AuditMeta, actor *domain.AuditUser, id uuid.UUID, input domain.UpdateUserInput) (*models.User, error) {
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
