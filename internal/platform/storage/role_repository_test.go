package storage

import (
	"testing"

	"portal-system/internal/domain/constants"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGormRoleRepository_Finders(t *testing.T) {
	t.Run("find by code success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)

		found, err := repo.FindByCode(ctx, role.Code)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, role.ID, found.ID)
	})

	t.Run("find by code not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		found, err := repo.FindByCode(ctx, constants.RoleCode("missing"))
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})

	t.Run("find by id success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)

		found, err := repo.FindByID(ctx, role.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, role.Code, found.Code)
	})

	t.Run("find by id not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		found, err := repo.FindByID(ctx, uuid.New())
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormRoleRepository_List(t *testing.T) {
	t.Run("list returns roles with permissions", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		perm := mustCreatePermission(t, tx, "role.list.read", "Role List Read")
		require.NoError(t, tx.Model(role).Association("Permissions").Append(perm))

		roles, err := repo.List(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, roles)

		var found *models.Role
		for i := range roles {
			if roles[i].ID == role.ID {
				found = &roles[i]
				break
			}
		}
		require.NotNil(t, found)
		require.Len(t, found.Permissions, 1)
		require.Equal(t, perm.ID, found.Permissions[0].ID)
	})
}

func TestGormRoleRepository_Permissions(t *testing.T) {
	t.Run("get with permissions success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		perm := mustCreatePermission(t, tx, "user.read", "User Read")
		require.NoError(t, tx.Model(role).Association("Permissions").Append(perm))

		found, err := repo.GetWithPermissions(ctx, role.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Len(t, found.Permissions, 1)
		require.Equal(t, perm.ID, found.Permissions[0].ID)
	})

	t.Run("get with permissions not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		found, err := repo.GetWithPermissions(ctx, uuid.New())
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})

	t.Run("assign and remove permission", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRoleRepository(testDB)
		role := mustCreateRole(t, tx, constants.RoleCodeUser)
		perm := mustCreatePermission(t, tx, "user.write", "User Write")

		require.NoError(t, repo.AssignPermission(ctx, role.ID, perm.ID))

		var afterAssign models.Role
		require.NoError(t, tx.Preload("Permissions").First(&afterAssign, "id = ?", role.ID).Error)
		require.Len(t, afterAssign.Permissions, 1)

		require.NoError(t, repo.RemovePermission(ctx, role.ID, perm.ID))

		var afterRemove models.Role
		require.NoError(t, tx.Preload("Permissions").First(&afterRemove, "id = ?", role.ID).Error)
		require.Len(t, afterRemove.Permissions, 0)
	})
}
