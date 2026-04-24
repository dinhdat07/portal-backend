package services_test

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	repositoriesmocks "portal-system/internal/mocks/repositories"
	servicesmocks "portal-system/internal/mocks/services"
	"portal-system/internal/models"
	. "portal-system/internal/services"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAdminService_ListUsers_Table(t *testing.T) {
	actor := &domain.AuditUser{ID: uuid.New(), RoleCode: constants.RoleCodeAdmin}
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	roleID := uuid.New()
	roleCode := constants.RoleCodeUser
	tests := []struct {
		name         string
		filter       domain.UsersFilter
		roleErr      error
		listErr      error
		listUsers    []models.User
		total        int64
		expected     error
		expectAudit  bool
		expectRoleID bool
	}{
		{
			name:     "invalid role code",
			filter:   domain.UsersFilter{RoleCode: &roleCode},
			roleErr:  errors.New("role not found"),
			expected: ErrInvalidInput,
		},
		{
			name:     "invalid status",
			filter:   domain.UsersFilter{Status: enum.UserStatus("bad")},
			expected: ErrInvalidInput,
		},
		{
			name:     "repo error",
			filter:   domain.UsersFilter{Status: enum.StatusActive},
			listErr:  errors.New("list failed"),
			expected: ErrInternalServer,
		},
		{
			name:         "success",
			filter:       domain.UsersFilter{RoleCode: &roleCode, Status: enum.StatusActive, Page: 1, PageSize: 20},
			listUsers:    []models.User{{ID: uuid.New()}},
			total:        1,
			expectAudit:  true,
			expectRoleID: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger := servicesmocks.NewAuditLogger(t)
			auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, enum.ActionAdminSearchUser, actor, (*domain.AuditUser)(nil), mock.Anything).Return(nil).Maybe()
			roleRepo := repositoriesmocks.NewRoleRepository(t)
			roleRepo.EXPECT().FindByCode(mock.Anything, roleCode).RunAndReturn(func(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
				if tc.roleErr != nil {
					return nil, tc.roleErr
				}
				return &models.Role{ID: roleID, Code: code}, nil
			}).Maybe()

			var capturedFilter domain.UsersFilter
			userRepo := repositoriesmocks.NewUserRepository(t)
			userRepo.EXPECT().ListUsers(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, filter domain.UsersFilter) ([]models.User, int64, error) {
				capturedFilter = filter
				return tc.listUsers, tc.total, tc.listErr
			}).Maybe()

			tx := repositoriesmocks.NewTxManager(t)
			tokenRepo := repositoriesmocks.NewUserTokenRepository(t)
			email := servicesmocks.NewEmailSender(t)
			tokenMgr := servicesmocks.NewTokenIssuer(t)
			svc := newAdminServiceForTest(tx, auditLogger, userRepo, tokenRepo, roleRepo, tokenMgr, email)
			out, err := svc.ListUsers(context.Background(), meta, actor, tc.filter)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if out == nil || out.Total != tc.total || len(out.Users) != len(tc.listUsers) {
					t.Fatalf("unexpected output: %#v", out)
				}
				if tc.expectRoleID && (capturedFilter.RoleID == nil || *capturedFilter.RoleID != roleID) {
					t.Fatalf("expected role id to be resolved in filter")
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}

			if tc.expectAudit {
				auditLogger.AssertNumberOfCalls(t, "LogWithMetadata", 1)
			}
		})
	}
}

func TestAdminService_CreateUser_Table(t *testing.T) {
	actor := &domain.AuditUser{ID: uuid.New(), RoleCode: constants.RoleCodeAdmin}
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	role := &models.Role{ID: uuid.New(), Code: constants.RoleCodeUser}
	input := domain.CreateUserInput{
		Email:     "john@example.com",
		Username:  "john",
		FirstName: "John",
		LastName:  "Doe",
		RoleCode:  constants.RoleCodeUser,
	}

	tests := []struct {
		name            string
		in              domain.CreateUserInput
		findEmail       *models.User
		findEmailErr    error
		findUsername    *models.User
		findUsernameErr error
		roleErr         error
		roleNil         bool
		createErr       error
		revokeErr       error
		tokenCreateErr  error
		emailErr        error
		auditErr        error
		expected        error
		expectCreate    bool
		expectEmailSend bool
	}{
		{
			name:     "missing role code",
			in:       domain.CreateUserInput{Email: input.Email},
			expected: ErrInvalidInput,
		},
		{
			name:         "find by email error",
			in:           input,
			findEmailErr: errors.New("db read failed"),
			expected:     ErrInternalServer,
		},
		{
			name:      "email exists",
			in:        input,
			findEmail: &models.User{ID: uuid.New()},
			expected:  ErrEmailExists,
		},
		{
			name:            "find by username error",
			in:              input,
			findUsernameErr: errors.New("db read failed"),
			expected:        ErrInternalServer,
		},
		{
			name:         "username exists",
			in:           input,
			findUsername: &models.User{ID: uuid.New()},
			expected:     ErrUsernameExists,
		},
		{
			name:     "role repository error",
			in:       input,
			roleErr:  errors.New("role lookup failed"),
			expected: ErrInternalServer,
		},
		{
			name:     "role is nil",
			in:       input,
			roleNil:  true,
			expected: ErrInternalServer,
		},
		{
			name:         "create user error",
			in:           input,
			createErr:    errors.New("insert failed"),
			expected:     ErrInternalServer,
			expectCreate: true,
		},
		{
			name:         "revoke token error",
			in:           input,
			revokeErr:    errors.New("revoke failed"),
			expected:     ErrInternalServer,
			expectCreate: true,
		},
		{
			name:           "token create error",
			in:             input,
			tokenCreateErr: errors.New("token create failed"),
			expected:       ErrInternalServer,
			expectCreate:   true,
		},
		{
			name:            "email send error",
			in:              input,
			emailErr:        errors.New("smtp failed"),
			expected:        ErrSendSetPasswordEmail,
			expectCreate:    true,
			expectEmailSend: true,
		},
		{
			name:            "audit error",
			in:              input,
			auditErr:        errors.New("audit failed"),
			expected:        ErrAuditLogger,
			expectCreate:    true,
			expectEmailSend: true,
		},
		{
			name:            "success",
			in:              input,
			expectCreate:    true,
			expectEmailSend: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := repositoriesmocks.NewTxManager(t)
			tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}).Maybe()
			auditLogger := servicesmocks.NewAuditLogger(t)
			auditLogger.EXPECT().Log(mock.Anything, meta, enum.ActionAdminCreateUser, actor, mock.Anything).Return(tc.auditErr).Maybe()
			userRepo := repositoriesmocks.NewUserRepository(t)
			userRepo.EXPECT().FindByEmail(mock.Anything, input.Email).RunAndReturn(func(ctx context.Context, email string) (*models.User, error) {
				return tc.findEmail, tc.findEmailErr
			}).Maybe()
			userRepo.EXPECT().FindByUsername(mock.Anything, input.Username).RunAndReturn(func(ctx context.Context, username string) (*models.User, error) {
				return tc.findUsername, tc.findUsernameErr
			}).Maybe()
			userRepo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, user *models.User) error {
				if user.ID == uuid.Nil {
					user.ID = uuid.New()
				}
				return tc.createErr
			}).Maybe()
			tokenRepo := repositoriesmocks.NewUserTokenRepository(t)
			tokenRepo.EXPECT().RevokeByUserAndType(mock.Anything, mock.Anything, enum.TokenTypePasswordSet).RunAndReturn(func(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error {
				return tc.revokeErr
			}).Maybe()
			tokenRepo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, token *models.UserToken) error {
				return tc.tokenCreateErr
			}).Maybe()
			roleRepo := repositoriesmocks.NewRoleRepository(t)
			roleRepo.EXPECT().FindByCode(mock.Anything, input.RoleCode).RunAndReturn(func(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
				if tc.roleErr != nil {
					return nil, tc.roleErr
				}
				if tc.roleNil {
					return nil, nil
				}
				return role, nil
			}).Maybe()
			email := servicesmocks.NewEmailSender(t)
			email.EXPECT().SendSetPasswordEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.emailErr).Maybe()
			tokenMgr := servicesmocks.NewTokenIssuer(t)
			tokenMgr.EXPECT().GenerateHashToken().Return("token-hash", "raw-token", nil).Maybe()
			svc := newAdminServiceForTest(tx, auditLogger, userRepo, tokenRepo, roleRepo, tokenMgr, email)

			user, err := svc.CreateUser(context.Background(), meta, actor, tc.in)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if user == nil || user.Email != tc.in.Email || user.RoleID != role.ID {
					t.Fatalf("unexpected user result: %#v", user)
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}

			if tc.expectCreate {
				tx.AssertNumberOfCalls(t, "WithTx", 1)
			}
			if tc.expectEmailSend {
				email.AssertNumberOfCalls(t, "SendSetPasswordEmail", 1)
			}
		})
	}
}

func TestAdminService_DeleteUser_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	adminID := uuid.New()
	actor := &domain.AuditUser{ID: adminID, RoleCode: constants.RoleCodeAdmin}
	adminRole := &models.Role{ID: uuid.New(), Code: constants.RoleCodeAdmin}
	baseUser := &models.User{
		ID:       uuid.New(),
		RoleID:   uuid.New(),
		Role:     models.Role{Code: constants.RoleCodeUser},
		Email:    "u@example.com",
		Username: "u1",
		Status:   enum.StatusActive,
	}

	tests := []struct {
		name        string
		findErr     error
		user        *models.User
		roleErr     error
		deleteErr   error
		auditErr    error
		expected    error
		expectTx    bool
		expectAudit bool
	}{
		{name: "user not found", findErr: gorm.ErrRecordNotFound, expected: ErrUserNotFound},
		{name: "role repo error", user: cloneUser(baseUser), roleErr: errors.New("role failed"), expected: ErrInternalServer},
		{
			name:     "forbidden delete other admin",
			user:     &models.User{ID: uuid.New(), RoleID: adminRole.ID, Role: *adminRole},
			expected: ErrForbidden,
		},
		{
			name:      "delete repository error",
			user:      cloneUser(baseUser),
			deleteErr: errors.New("delete failed"),
			expected:  ErrInternalServer,
			expectTx:  true,
		},
		{
			name:        "audit error",
			user:        cloneUser(baseUser),
			auditErr:    errors.New("audit failed"),
			expected:    ErrAuditLogger,
			expectTx:    true,
			expectAudit: true,
		},
		{
			name:        "success",
			user:        cloneUser(baseUser),
			expectTx:    true,
			expectAudit: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger := servicesmocks.NewAuditLogger(t)
			auditLogger.EXPECT().Log(mock.Anything, meta, enum.ActionAdminDeleteUser, actor, mock.Anything).Return(tc.auditErr).Maybe()
			userRepo := repositoriesmocks.NewUserRepository(t)
			userRepo.EXPECT().FindByID(mock.Anything, baseUser.ID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				if tc.findErr != nil {
					return nil, tc.findErr
				}
				return cloneUser(tc.user), nil
			}).Maybe()
			userRepo.EXPECT().Delete(mock.Anything, baseUser.ID, actor.ID).RunAndReturn(func(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
				return tc.deleteErr
			}).Maybe()
			roleRepo := repositoriesmocks.NewRoleRepository(t)
			roleRepo.EXPECT().FindByCode(mock.Anything, constants.RoleCodeAdmin).RunAndReturn(func(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
				if tc.roleErr != nil {
					return nil, tc.roleErr
				}
				return adminRole, nil
			}).Maybe()
			tx := repositoriesmocks.NewTxManager(t)
			tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}).Maybe()
			svc := newAdminServiceForTest(tx, auditLogger, userRepo, nil, roleRepo, nil, nil)

			got, err := svc.DeleteUser(context.Background(), meta, actor, baseUser.ID)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if got == nil || !got.DeletedAt.Valid || got.Status != enum.StatusDeleted || got.DeletedBy == nil || *got.DeletedBy != actor.ID {
					t.Fatalf("unexpected deleted user result: %#v", got)
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}

			if tc.expectTx {
				tx.AssertNumberOfCalls(t, "WithTx", 1)
			}
			if tc.expectAudit {
				auditLogger.AssertNumberOfCalls(t, "Log", 1)
			}
		})
	}
}

func TestAdminService_RestoreUser_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &domain.AuditUser{ID: uuid.New(), RoleCode: constants.RoleCodeAdmin}
	userID := uuid.New()
	deletedAt := gorm.DeletedAt{Time: time.Now().Add(-time.Hour), Valid: true}

	tests := []struct {
		name        string
		findErr     error
		user        *models.User
		restoreErr  error
		auditErr    error
		expected    error
		expectTx    bool
		expectAudit bool
	}{
		{name: "user not found", findErr: gorm.ErrRecordNotFound, expected: ErrUserNotFound},
		{name: "user not deleted", user: &models.User{ID: userID, DeletedAt: gorm.DeletedAt{}}, expected: ErrUserNotDeleted},
		{
			name:       "restore repository error",
			user:       &models.User{ID: userID, DeletedAt: deletedAt},
			restoreErr: errors.New("restore failed"),
			expected:   ErrInternalServer,
			expectTx:   true,
		},
		{
			name:        "audit error",
			user:        &models.User{ID: userID, DeletedAt: deletedAt},
			auditErr:    errors.New("audit failed"),
			expected:    ErrAuditLogger,
			expectTx:    true,
			expectAudit: true,
		},
		{
			name:        "success",
			user:        &models.User{ID: userID, DeletedAt: deletedAt, Status: enum.StatusDeleted},
			expectTx:    true,
			expectAudit: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger := servicesmocks.NewAuditLogger(t)
			auditLogger.EXPECT().Log(mock.Anything, meta, enum.ActionAdminRestoreUser, actor, mock.Anything).Return(tc.auditErr).Maybe()
			userRepo := repositoriesmocks.NewUserRepository(t)
			userRepo.EXPECT().FindByIDUnscoped(mock.Anything, userID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				if tc.findErr != nil {
					return nil, tc.findErr
				}
				return cloneUser(tc.user), nil
			}).Maybe()
			userRepo.EXPECT().Restore(mock.Anything, userID).RunAndReturn(func(ctx context.Context, id uuid.UUID) error {
				return tc.restoreErr
			}).Maybe()
			tx := repositoriesmocks.NewTxManager(t)
			tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}).Maybe()
			svc := newAdminServiceForTest(tx, auditLogger, userRepo, nil, nil, nil, nil)

			got, err := svc.RestoreUser(context.Background(), meta, actor, userID)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if got == nil || got.DeletedAt.Valid || got.DeletedBy != nil || got.Status != enum.StatusActive {
					t.Fatalf("unexpected restored user result: %#v", got)
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}

			if tc.expectTx {
				tx.AssertNumberOfCalls(t, "WithTx", 1)
			}
			if tc.expectAudit {
				auditLogger.AssertNumberOfCalls(t, "Log", 1)
			}
		})
	}
}

func TestAdminService_UpdateRole_Table(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actorID := uuid.New()
	actor := &domain.AuditUser{ID: actorID, RoleCode: constants.RoleCodeAdmin}
	userID := uuid.New()
	adminRole := &models.Role{ID: uuid.New(), Code: constants.RoleCodeAdmin}
	userRole := &models.Role{ID: uuid.New(), Code: constants.RoleCodeUser}

	tests := []struct {
		name          string
		roleCode      constants.RoleCode
		findErr       error
		user          *models.User
		roleErr       error
		adminRoleErr  error
		updateRoleErr error
		auditErr      error
		expected      error
		expectTx      bool
		expectAudit   bool
		sameRole      bool
	}{
		{name: "empty role code", roleCode: "", expected: ErrInvalidInput},
		{name: "user not found", roleCode: constants.RoleCodeUser, findErr: gorm.ErrRecordNotFound, expected: ErrUserNotFound},
		{
			name:     "target role lookup error",
			roleCode: constants.RoleCodeUser,
			user:     &models.User{ID: userID, RoleID: userRole.ID, Role: *userRole},
			roleErr:  errors.New("role lookup failed"),
			expected: ErrInternalServer,
		},
		{
			name:         "admin role lookup error",
			roleCode:     constants.RoleCodeUser,
			user:         &models.User{ID: userID, RoleID: userRole.ID, Role: *userRole},
			adminRoleErr: errors.New("admin role failed"),
			expected:     ErrInternalServer,
		},
		{
			name:     "forbidden for other admin",
			roleCode: constants.RoleCodeUser,
			user:     &models.User{ID: uuid.New(), RoleID: adminRole.ID, Role: *adminRole},
			expected: ErrForbidden,
		},
		{
			name:     "same role no-op",
			roleCode: constants.RoleCodeUser,
			user:     &models.User{ID: userID, RoleID: userRole.ID, Role: *userRole},
			sameRole: true,
		},
		{
			name:          "update role repository error",
			roleCode:      constants.RoleCodeUser,
			user:          &models.User{ID: actorID, RoleID: adminRole.ID, Role: *adminRole},
			updateRoleErr: errors.New("update failed"),
			expected:      ErrInternalServer,
			expectTx:      true,
		},
		{
			name:        "audit error",
			roleCode:    constants.RoleCodeUser,
			user:        &models.User{ID: actorID, RoleID: adminRole.ID, Role: *adminRole, Email: "a@b.com", Username: "u"},
			auditErr:    errors.New("audit failed"),
			expected:    ErrAuditLogger,
			expectTx:    true,
			expectAudit: true,
		},
		{
			name:        "success",
			roleCode:    constants.RoleCodeUser,
			user:        &models.User{ID: actorID, RoleID: adminRole.ID, Role: *adminRole, Email: "a@b.com", Username: "u"},
			expectTx:    true,
			expectAudit: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			targetRole := userRole
			if tc.sameRole {
				targetRole = &models.Role{ID: tc.user.RoleID, Code: tc.user.Role.Code}
			}
			auditLogger := servicesmocks.NewAuditLogger(t)
			auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, enum.ActionAdminAssignRole, actor, mock.Anything, mock.Anything).Return(tc.auditErr).Maybe()
			userRepo := repositoriesmocks.NewUserRepository(t)
			userRepo.EXPECT().FindByID(mock.Anything, userID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				if tc.findErr != nil {
					return nil, tc.findErr
				}
				return cloneUser(tc.user), nil
			}).Maybe()
			userRepo.EXPECT().UpdateRole(mock.Anything, userID, mock.Anything).RunAndReturn(func(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error {
				return tc.updateRoleErr
			}).Maybe()
			roleRepo := repositoriesmocks.NewRoleRepository(t)
			roleRepo.EXPECT().FindByCode(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
				switch code {
				case constants.RoleCodeAdmin:
					if tc.adminRoleErr != nil {
						return nil, tc.adminRoleErr
					}
					return adminRole, nil
				default:
					if tc.roleErr != nil {
						return nil, tc.roleErr
					}
					return targetRole, nil
				}
			}).Maybe()
			tx := repositoriesmocks.NewTxManager(t)
			tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}).Maybe()
			svc := newAdminServiceForTest(tx, auditLogger, userRepo, nil, roleRepo, nil, nil)

			got, err := svc.UpdateRole(context.Background(), meta, actor, userID, tc.roleCode)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if got == nil {
					t.Fatal("expected user output")
				}
			} else if !errors.Is(err, tc.expected) {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}

			if tc.expectTx {
				tx.AssertNumberOfCalls(t, "WithTx", 1)
			}
			if tc.expectAudit {
				auditLogger.AssertNumberOfCalls(t, "LogWithMetadata", 1)
			}
		})
	}
}
