package storage

import (
	"context"
	"testing"
	"time"

	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGormRefreshTokenRepository_Create(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormRefreshTokenRepository(testDB)
	role := activeUserRole(t, tx)
	user := mustCreateUser(t, tx, role.ID, "rtcreate@example.com", "rtcreate")
	session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))

	token := &models.RefreshToken{
		SessionID: session.ID,
		UserID:    user.ID,
		TokenHash: "create-test-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	err := repo.Create(ctx, token)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, token.ID)
}

func TestGormRefreshTokenRepository_FindByTokenHash(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		token := mustCreateRefreshToken(t, tx, session.ID, user.ID, "refresh-hash", time.Now().Add(time.Hour))

		found, err := repo.FindByTokenHash(ctx, token.TokenHash)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, token.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)

		found, err := repo.FindByTokenHash(ctx, "missing")
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormRefreshTokenRepository_RevokeAndReplacement(t *testing.T) {
	t.Run("revoke by id success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		token := mustCreateRefreshToken(t, tx, session.ID, user.ID, "refresh-hash", time.Now().Add(time.Hour))

		require.NoError(t, repo.RevokeByID(ctx, token.ID))

		var updated struct{ RevokedAt *time.Time }
		require.NoError(t, tx.WithContext(context.Background()).Model(token).Select("revoked_at").First(&updated, "id = ?", token.ID).Error)
		require.NotNil(t, updated.RevokedAt)
	})

	t.Run("revoke by id not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		require.ErrorIs(t, repo.RevokeByID(ctx, uuid.New()), repositories.ErrNotFound)
	})

	t.Run("revoke by user id success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		token := mustCreateRefreshToken(t, tx, session.ID, user.ID, "refresh-hash", time.Now().Add(time.Hour))

		require.NoError(t, repo.RevokeByUserID(ctx, user.ID))

		var updated models.RefreshToken
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", token.ID).Error)
		require.NotNil(t, updated.RevokedAt)
	})

	t.Run("revoke by user id not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		require.ErrorIs(t, repo.RevokeByUserID(ctx, uuid.New()), repositories.ErrNotFound)
	})

	t.Run("revoke by session id success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		token := mustCreateRefreshToken(t, tx, session.ID, user.ID, "refresh-hash", time.Now().Add(time.Hour))

		require.NoError(t, repo.RevokeBySessionID(ctx, session.ID))

		var updated models.RefreshToken
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", token.ID).Error)
		require.NotNil(t, updated.RevokedAt)
	})

	t.Run("revoke by session id not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		require.ErrorIs(t, repo.RevokeBySessionID(ctx, uuid.New()), repositories.ErrNotFound)
	})

	t.Run("mark replacement success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		original := mustCreateRefreshToken(t, tx, session.ID, user.ID, "refresh-hash", time.Now().Add(time.Hour))
		replacement := mustCreateRefreshToken(t, tx, session.ID, user.ID, "refresh-hash-2", time.Now().Add(2*time.Hour))

		require.NoError(t, repo.MarkReplacement(ctx, original.ID, replacement.ID))

		var updated models.RefreshToken
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", original.ID).Error)
		require.NotNil(t, updated.ReplacedByTokenID)
		require.Equal(t, replacement.ID, *updated.ReplacedByTokenID)
	})

	t.Run("mark replacement not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormRefreshTokenRepository(testDB)
		require.ErrorIs(t, repo.MarkReplacement(ctx, uuid.New(), uuid.New()), repositories.ErrNotFound)
	})
}
