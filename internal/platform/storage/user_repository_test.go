package storage

import (
	"context"
	"testing"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGormUserRepository_Create(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormUserRepository(testDB)
	role := mustCreateRole(t, tx, constants.RoleCodeUser)

	user := &models.User{
		Email:     "john@example.com",
		Username:  "john",
		FirstName: "John",
		LastName:  "Doe",
		RoleID:    role.ID,
		Status:    enum.StatusPending,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user.ID)

	var stored models.User
	err = tx.WithContext(context.Background()).First(&stored, "id = ?", user.ID).Error
	require.NoError(t, err)
	require.Equal(t, user.Email, stored.Email)
	require.Equal(t, user.Username, stored.Username)
}

func TestGormUserRepository_Create_DuplicateEmail(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormUserRepository(testDB)
	role := mustCreateRole(t, tx, constants.RoleCodeUser)

	existing := &models.User{
		Email:     "john@example.com",
		Username:  "john",
		FirstName: "John",
		LastName:  "Doe",
		RoleID:    role.ID,
		Status:    enum.StatusPending,
	}
	require.NoError(t, repo.Create(ctx, existing))

	duplicate := &models.User{
		Email:     existing.Email,
		Username:  "john-2",
		FirstName: "John",
		LastName:  "Other",
		RoleID:    role.ID,
		Status:    enum.StatusPending,
	}

	err := repo.Create(ctx, duplicate)
	require.Error(t, err)
}

func TestGormUserRepository_FindByEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")

		found, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, user.ID, found.ID)
		require.Equal(t, role.ID, found.Role.ID)
		require.Equal(t, role.Code, found.Role.Code)
	})

	t.Run("not found returns nil nil", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		found, err := repo.FindByEmail(ctx, "missing@example.com")
		require.NoError(t, err)
		require.Nil(t, found)
	})
}

func TestGormUserRepository_FindByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "findid@example.com", "findid")

		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, user.ID, found.ID)
		require.Equal(t, role.ID, found.Role.ID)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		found, err := repo.FindByID(ctx, uuid.New())
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormUserRepository_FindByIDUnscoped(t *testing.T) {
	t.Run("finds soft-deleted user", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "unscoped@example.com", "unscoped")

		// soft-delete the user
		deletedBy := uuid.New()
		require.NoError(t, tx.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
			"status":     enum.StatusDeleted,
		}).Error)

		// FindByID should NOT find it
		notFound, err := repo.FindByID(ctx, user.ID)
		require.Nil(t, notFound)
		require.ErrorIs(t, err, repositories.ErrNotFound)

		// FindByIDUnscoped SHOULD find it
		found, err := repo.FindByIDUnscoped(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, user.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		found, err := repo.FindByIDUnscoped(ctx, uuid.New())
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormUserRepository_FindByUsername(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "byuser@example.com", "findbyuname")

		found, err := repo.FindByUsername(ctx, user.Username)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, user.ID, found.ID)
		require.Equal(t, role.ID, found.Role.ID)
	})

	t.Run("not found returns nil nil", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		found, err := repo.FindByUsername(ctx, "nonexistent")
		require.NoError(t, err)
		require.Nil(t, found)
	})
}

func TestGormUserRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "update@example.com", "updateuser")

		newDob := time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC)
		user.FirstName = "Updated"
		user.LastName = "Name"
		user.Username = "updateduser"
		user.DOB = &newDob

		err := repo.Update(ctx, user)
		require.NoError(t, err)

		var updated models.User
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", user.ID).Error)
		require.Equal(t, "Updated", updated.FirstName)
		require.Equal(t, "Name", updated.LastName)
		require.Equal(t, "updateduser", updated.Username)
		require.NotNil(t, updated.DOB)
		require.Equal(t, newDob, *updated.DOB)
	})
}

func TestGormUserRepository_UpdatePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")

		err := repo.UpdatePassword(ctx, user.ID, "hashed-password")
		require.NoError(t, err)

		var updated models.User
		err = tx.WithContext(context.Background()).First(&updated, "id = ?", user.ID).Error
		require.NoError(t, err)
		require.NotNil(t, updated.PasswordHash)
		require.Equal(t, "hashed-password", *updated.PasswordHash)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		err := repo.UpdatePassword(ctx, uuid.New(), "hashed-password")
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormUserRepository_UpdateRole(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		newRole := mustCreateRole(t, tx, constants.RoleCodeAdmin)
		user := mustCreateUser(t, tx, role.ID, "uprole@example.com", "uprole")

		err := repo.UpdateRole(ctx, user.ID, newRole.ID)
		require.NoError(t, err)

		var updated models.User
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", user.ID).Error)
		require.Equal(t, newRole.ID, updated.RoleID)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		err := repo.UpdateRole(ctx, uuid.New(), uuid.New())
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormUserRepository_MarkEmailVerified(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "verify@example.com", "verifyuser")

		// user starts as active (from mustCreateUser), set to pending first
		require.NoError(t, tx.Model(&models.User{}).Where("id = ?", user.ID).Update("status", enum.StatusPending).Error)

		err := repo.MarkEmailVerified(ctx, user.ID)
		require.NoError(t, err)

		var updated models.User
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", user.ID).Error)
		require.NotNil(t, updated.EmailVerifiedAt)
		require.Equal(t, enum.StatusActive, updated.Status)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		err := repo.MarkEmailVerified(ctx, uuid.New())
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormUserRepository_DeleteAndRestore(t *testing.T) {
	t.Run("delete success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "del@example.com", "deluser")
		deletedBy := uuid.New()

		err := repo.Delete(ctx, user.ID, deletedBy)
		require.NoError(t, err)

		// soft-deleted: not found via scoped query
		var scoped models.User
		scopedErr := tx.WithContext(context.Background()).First(&scoped, "id = ?", user.ID).Error
		require.Error(t, scopedErr)

		// but found via unscoped
		var unscoped models.User
		require.NoError(t, tx.WithContext(context.Background()).Unscoped().First(&unscoped, "id = ?", user.ID).Error)
		require.Equal(t, enum.StatusDeleted, unscoped.Status)
		require.NotNil(t, unscoped.DeletedBy)
		require.Equal(t, deletedBy, *unscoped.DeletedBy)
	})

	t.Run("delete not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		err := repo.Delete(ctx, uuid.New(), uuid.New())
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})

	t.Run("restore success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "restore@example.com", "restoreuser")
		deletedBy := uuid.New()

		// delete first
		require.NoError(t, repo.Delete(ctx, user.ID, deletedBy))

		// restore
		err := repo.Restore(ctx, user.ID)
		require.NoError(t, err)

		// should be findable again via scoped query
		var restored models.User
		require.NoError(t, tx.WithContext(context.Background()).First(&restored, "id = ?", user.ID).Error)
		require.Equal(t, enum.StatusActive, restored.Status)
		require.Nil(t, restored.DeletedBy)
	})

	t.Run("restore not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserRepository(testDB)

		err := repo.Restore(ctx, uuid.New())
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormUserRepository_ListUsers(t *testing.T) {
	t.Run("basic list with pagination", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		u1 := mustCreateUser(t, tx, role.ID, "list1@example.com", "listuser1")
		u2 := mustCreateUser(t, tx, role.ID, "list2@example.com", "listuser2")
		u3 := mustCreateUser(t, tx, role.ID, "list3@example.com", "listuser3")

		users, total, err := repo.ListUsers(ctx, domain.UsersFilter{
			Page:     1,
			PageSize: 2,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(3))
		require.Len(t, users, 2)

		// second page
		users2, _, err := repo.ListUsers(ctx, domain.UsersFilter{
			Page:     2,
			PageSize: 2,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(users2), 1)

		_ = u1
		_ = u2
		_ = u3
	})

	t.Run("filter by username", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		mustCreateUser(t, tx, role.ID, "ufilter1@example.com", "alphauser")
		mustCreateUser(t, tx, role.ID, "ufilter2@example.com", "betauser")

		users, total, err := repo.ListUsers(ctx, domain.UsersFilter{
			Username: "alpha",
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, users, 1)
		require.Equal(t, "alphauser", users[0].Username)
	})

	t.Run("filter by email", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		mustCreateUser(t, tx, role.ID, "emailsearch@special.com", "emailsrch")
		mustCreateUser(t, tx, role.ID, "other@example.com", "otheruser")

		users, total, err := repo.ListUsers(ctx, domain.UsersFilter{
			Email:    "emailsearch@special",
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, users, 1)
		require.Equal(t, "emailsearch@special.com", users[0].Email)
	})

	t.Run("filter by full name", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		mustCreateUser(t, tx, role.ID, "fname@example.com", "fnameuser")

		users, total, err := repo.ListUsers(ctx, domain.UsersFilter{
			FullName: "John Doe",
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		require.NotEmpty(t, users)
	})

	t.Run("filter by dob", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		mustCreateUser(t, tx, role.ID, "dob@example.com", "dobuser")

		dob := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		users, total, err := repo.ListUsers(ctx, domain.UsersFilter{
			Dob:      &dob,
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		require.NotEmpty(t, users)
	})

	t.Run("filter by role id", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role1 := mustCreateRole(t, tx, constants.RoleCodeUser)
		role2 := mustCreateRole(t, tx, constants.RoleCodeAdmin)
		mustCreateUser(t, tx, role1.ID, "rolefilt1@example.com", "rolefilt1")
		mustCreateUser(t, tx, role2.ID, "rolefilt2@example.com", "rolefilt2")

		users, total, err := repo.ListUsers(ctx, domain.UsersFilter{
			RoleID:   &role1.ID,
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, users, 1)
		require.Equal(t, role1.ID, users[0].RoleID)
	})

	t.Run("filter by status", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "statusfilt@example.com", "statusfilt")

		// set user to pending
		require.NoError(t, tx.Model(&models.User{}).Where("id = ?", user.ID).Update("status", enum.StatusPending).Error)

		users, total, err := repo.ListUsers(ctx, domain.UsersFilter{
			Status:   enum.StatusPending,
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, users, 1)
		require.Equal(t, user.ID, users[0].ID)
	})

	t.Run("include deleted", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		user := mustCreateUser(t, tx, role.ID, "incldel@example.com", "incldel")
		deletedBy := uuid.New()

		// soft-delete
		require.NoError(t, tx.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
			"status":     enum.StatusDeleted,
		}).Error)

		// without include deleted
		users, _, err := repo.ListUsers(ctx, domain.UsersFilter{
			Email:    "incldel@example.com",
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		require.Empty(t, users)

		// with include deleted
		users2, total2, err := repo.ListUsers(ctx, domain.UsersFilter{
			Email:          "incldel@example.com",
			IncludeDeleted: true,
			Page:           1,
			PageSize:       10,
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), total2)
		require.Len(t, users2, 1)
		require.Equal(t, user.ID, users2[0].ID)
	})
}

func mustCreateRole(t *testing.T, tx *gorm.DB, code constants.RoleCode) *models.Role {
	t.Helper()

	role := &models.Role{
		ID:   uuid.New(),
		Code: code,
		Name: string(code),
	}
	require.NoError(t, tx.Create(role).Error)

	return role
}

func mustCreateUser(t *testing.T, tx *gorm.DB, roleID uuid.UUID, email, username string) *models.User {
	t.Helper()

	dob := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	user := &models.User{
		Email:     email,
		Username:  username,
		FirstName: "John",
		LastName:  "Doe",
		DOB:       &dob,
		RoleID:    roleID,
		Status:    enum.StatusActive,
	}
	require.NoError(t, tx.Create(user).Error)

	return user
}
