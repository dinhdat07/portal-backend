package services

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

func (svc *UserService) GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (svc *UserService) ChangePassword(ctx context.Context, id uuid.UUID, current, newPassword, confirm string) error {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnauthorized
		}
		return err
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

	if err := svc.userRepo.UpdatePassword(ctx, id, string(hashed)); err != nil {
		return err
	}

	return nil
}

func (svc *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, input domain.UpdateUserInput) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// update allowed fields
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.DOB != nil {
		user.DOB = input.DOB
	}

	// check duplicate username
	if input.Username != nil && *input.Username != user.Username {
		existing, err := svc.userRepo.FindByUsername(ctx, *input.Username)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, ErrUsernameExists
		}
		user.Username = *input.Username
	}

	if err := svc.userRepo.Update(ctx, user); err != nil {
		return nil, ErrInternalServer
	}

	return user, nil
}
