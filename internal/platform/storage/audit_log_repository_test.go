package storage

import (
	"context"
	"testing"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGormAuditLogRepository_Create(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormAuditLogRepository(testDB)
	role := activeUserRole(t, tx)
	user := mustCreateUser(t, tx, role.ID, "audit-create@example.com", "auditcreate")

	log := &models.AuditLog{
		Action:      enum.ActionLogin,
		ActorUserID: &user.ID,
	}
	err := repo.Create(ctx, log)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, log.ID)
}

func TestGormAuditLogRepository_Create_NoTx(t *testing.T) {
	repo := NewGormAuditLogRepository(testDB)
	log := &models.AuditLog{
		Action: enum.ActionLogin,
	}
	err := repo.Create(context.Background(), log)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, log.ID)

	// cleanup
	testDB.Unscoped().Delete(log)
}

func TestGormAuditLogRepository_CreateAndList(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormAuditLogRepository(testDB)
	actorID := uuid.New()
	now := time.Now().UTC()

	first := mustCreateAuditLog(t, tx, enum.ActionLogin, &actorID, now.Add(-time.Minute))
	_ = first
	second := mustCreateAuditLog(t, tx, enum.ActionLogout, &actorID, now)

	logs, total, err := repo.List(ctx, domain.AuditLogFilter{
		ActorUserID: &actorID,
		Page:        1,
		PageSize:    10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, logs, 2)
	require.Equal(t, second.ID, logs[0].ID)
}

func TestGormAuditLogRepository_ListByAction(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormAuditLogRepository(testDB)
	actorID := uuid.New()

	mustCreateAuditLog(t, tx, enum.ActionLogin, &actorID, time.Now().Add(-time.Minute))
	match := mustCreateAuditLog(t, tx, enum.ActionLogout, &actorID, time.Now())

	logs, total, err := repo.List(ctx, domain.AuditLogFilter{
		Action:   string(enum.ActionLogout),
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
	require.Equal(t, match.ID, logs[0].ID)
}

func TestGormAuditLogRepository_ListByTargetUserID(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormAuditLogRepository(testDB)
	actorID := uuid.New()
	targetID := uuid.New()
	otherTargetID := uuid.New()

	// create log with target
	logWithTarget := &models.AuditLog{
		Action:       enum.ActionAdminViewUser,
		ActorUserID:  &actorID,
		TargetUserID: &targetID,
	}
	require.NoError(t, tx.Create(logWithTarget).Error)

	// create log with other target
	logOtherTarget := &models.AuditLog{
		Action:       enum.ActionAdminViewUser,
		ActorUserID:  &actorID,
		TargetUserID: &otherTargetID,
	}
	require.NoError(t, tx.Create(logOtherTarget).Error)

	logs, total, err := repo.List(ctx, domain.AuditLogFilter{
		TargetUserID: &targetID,
		Page:         1,
		PageSize:     10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
	require.Equal(t, logWithTarget.ID, logs[0].ID)
}

func TestGormAuditLogRepository_ListByDateRange(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormAuditLogRepository(testDB)
	actorID := uuid.New()
	now := time.Now().UTC()

	old := mustCreateAuditLog(t, tx, enum.ActionLogin, &actorID, now.Add(-48*time.Hour))
	recent := mustCreateAuditLog(t, tx, enum.ActionLogin, &actorID, now.Add(-1*time.Hour))

	from := now.Add(-24 * time.Hour)
	to := now

	logs, total, err := repo.List(ctx, domain.AuditLogFilter{
		ActorUserID: &actorID,
		From:        &from,
		To:          &to,
		Page:        1,
		PageSize:    10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
	require.Equal(t, recent.ID, logs[0].ID)
	require.NotEqual(t, old.ID, logs[0].ID)
}

func TestGormAuditLogRepository_ListDefaultPagination(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormAuditLogRepository(testDB)
	actorID := uuid.New()

	mustCreateAuditLog(t, tx, enum.ActionLogin, &actorID, time.Now())

	// Page <= 0 and PageSize <= 0 should default to 1 and 20
	logs, total, err := repo.List(ctx, domain.AuditLogFilter{
		ActorUserID: &actorID,
		Page:        0,
		PageSize:    0,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
}

func TestGormAuditLogRepository_ListMaxPageSize(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormAuditLogRepository(testDB)
	actorID := uuid.New()

	mustCreateAuditLog(t, tx, enum.ActionLogin, &actorID, time.Now())

	// PageSize > 100 should be capped at 100
	logs, total, err := repo.List(ctx, domain.AuditLogFilter{
		ActorUserID: &actorID,
		Page:        1,
		PageSize:    200,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
}
