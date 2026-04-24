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

func TestGormAuthSessionRepository_CreateAndFind(t *testing.T) {
	t.Run("create and find active", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")

		session := &models.AuthSession{
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour),
			UserAgent: "test-agent",
			IPAddress: "127.0.0.1",
		}
		require.NoError(t, repo.Create(ctx, session))

		found, err := repo.FindActiveByID(ctx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, session.ID, found.ID)
	})

	t.Run("not found for revoked", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		now := time.Now().UTC()
		require.NoError(t, tx.Model(session).Update("revoked_at", &now).Error)

		found, err := repo.FindActiveByID(ctx, session.ID)
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})

	t.Run("not found for expired", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(-time.Hour))

		found, err := repo.FindActiveByID(ctx, session.ID)
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormAuthSessionRepository_RevokeAndList(t *testing.T) {
	t.Run("revoke by id success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		session := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))

		require.NoError(t, repo.RevokeByID(ctx, session.ID))

		var updated models.AuthSession
		require.NoError(t, tx.WithContext(context.Background()).First(&updated, "id = ?", session.ID).Error)
		require.NotNil(t, updated.RevokedAt)
	})

	t.Run("revoke by id not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		require.ErrorIs(t, repo.RevokeByID(ctx, uuid.New()), repositories.ErrNotFound)
	})

	t.Run("revoke all by user id success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		s1 := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		s2 := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(2*time.Hour))

		require.NoError(t, repo.RevokeAllByUserID(ctx, user.ID))

		var sessions []models.AuthSession
		require.NoError(t, tx.WithContext(context.Background()).Find(&sessions, "id IN ?", []uuid.UUID{s1.ID, s2.ID}).Error)
		require.Len(t, sessions, 2)
		require.NotNil(t, sessions[0].RevokedAt)
		require.NotNil(t, sessions[1].RevokedAt)
	})

	t.Run("revoke all by user id not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		require.ErrorIs(t, repo.RevokeAllByUserID(ctx, uuid.New()), repositories.ErrNotFound)
	})

	t.Run("list active by user id filters revoked and expired", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormAuthSessionRepository(testDB)
		role := activeUserRole(t, tx)
		user := mustCreateUser(t, tx, role.ID, "john@example.com", "john")
		active := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(time.Hour))
		revoked := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(2*time.Hour))
		expired := mustCreateAuthSession(t, tx, user.ID, time.Now().Add(-time.Hour))
		now := time.Now().UTC()
		require.NoError(t, tx.Model(revoked).Update("revoked_at", &now).Error)

		got, err := repo.ListActiveByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, active.ID, got[0].ID)
		require.NotEqual(t, expired.ID, got[0].ID)
	})
}
