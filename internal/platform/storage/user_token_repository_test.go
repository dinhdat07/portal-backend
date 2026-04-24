package storage

import (
	"context"
	"testing"
	"time"

	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGormUserTokenRepository_Create(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormUserTokenRepository(testDB)
	role := activeUserRole(t, tx)
	user := mustCreateUser(t, tx, role.ID, "tokencreate@example.com", "tokencreate")

	token := &models.UserToken{
		UserID:    user.ID,
		TokenType: enum.TokenTypeEmailVerification,
		TokenHash: "create-test-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	err := repo.Create(ctx, token)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, token.ID)
}

func TestGormUserTokenRepository_CreateAndFindValid(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		token := mustCreateUserToken(t, tx, user.ID, enum.TokenTypeEmailVerification, "token-hash", time.Now().Add(time.Hour))

		found, err := repo.FindValidToken(ctx, token.TokenHash, token.TokenType)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, token.ID, found.ID)
		require.Equal(t, user.ID, found.User.ID)
	})

	t.Run("not found for used revoked or expired", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(t *testing.T, tx *gorm.DB, token *models.UserToken)
		}{
			{
				name: "used",
				mutate: func(t *testing.T, tx *gorm.DB, token *models.UserToken) {
					now := time.Now().UTC()
					require.NoError(t, tx.Model(token).Update("used_at", &now).Error)
				},
			},
			{
				name: "revoked",
				mutate: func(t *testing.T, tx *gorm.DB, token *models.UserToken) {
					now := time.Now().UTC()
					require.NoError(t, tx.Model(token).Update("revoked_at", &now).Error)
				},
			},
			{
				name: "expired",
				mutate: func(t *testing.T, tx *gorm.DB, token *models.UserToken) {
					require.NoError(t, tx.Model(token).Update("expires_at", time.Now().Add(-time.Hour)).Error)
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				ctx, tx := newTestTx(t)
				repo := NewGormUserTokenRepository(testDB)
				role := activeUserRole(t, tx)
				user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
				token := mustCreateUserToken(t, tx, user.ID, enum.TokenTypeEmailVerification, "token-hash-"+tc.name, time.Now().Add(time.Hour))
				tc.mutate(t, tx, token)

				found, err := repo.FindValidToken(ctx, token.TokenHash, token.TokenType)
				require.Nil(t, found)
				require.ErrorIs(t, err, repositories.ErrNotFound)
			})
		}
	})
}

func TestGormUserTokenRepository_MarkUsedAndRevoke(t *testing.T) {
	t.Run("mark used success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		token := mustCreateUserToken(t, tx, user.ID, enum.TokenTypeEmailVerification, "token-hash", time.Now().Add(time.Hour))

		require.NoError(t, repo.MarkUsed(ctx, token.ID))

		var updated models.UserToken
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", token.ID).Error)
		require.NotNil(t, updated.UsedAt)
	})

	t.Run("mark used not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserTokenRepository(testDB)
		require.ErrorIs(t, repo.MarkUsed(ctx, uuid.New()), repositories.ErrNotFound)
	})

	t.Run("revoke success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		token := mustCreateUserToken(t, tx, user.ID, enum.TokenTypeEmailVerification, "token-hash", time.Now().Add(time.Hour))

		require.NoError(t, repo.Revoke(ctx, token.ID))

		var updated models.UserToken
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", token.ID).Error)
		require.NotNil(t, updated.RevokedAt)
	})

	t.Run("revoke not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormUserTokenRepository(testDB)
		require.ErrorIs(t, repo.Revoke(ctx, uuid.New()), repositories.ErrNotFound)
	})

	t.Run("revoke by user and type success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormUserTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		matching := mustCreateUserToken(t, tx, user.ID, enum.TokenTypeEmailVerification, "token-hash", time.Now().Add(time.Hour))
		otherType := mustCreateUserToken(t, tx, user.ID, enum.TokenTypePasswordReset, "token-hash-2", time.Now().Add(time.Hour))

		require.NoError(t, repo.RevokeByUserAndType(ctx, user.ID, enum.TokenTypeEmailVerification))

		var updatedMatch, updatedOther models.UserToken
		require.NoError(t, tx.WithContext(context.Background()).First(&updatedMatch, "id = ?", matching.ID).Error)
		require.NoError(t, tx.WithContext(context.Background()).First(&updatedOther, "id = ?", otherType.ID).Error)
		require.NotNil(t, updatedMatch.RevokedAt)
		require.Nil(t, updatedOther.RevokedAt)
	})
}
