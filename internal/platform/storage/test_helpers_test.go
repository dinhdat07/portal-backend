package storage

import (
	"testing"
	"time"

	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func mustCreatePermission(t *testing.T, tx *gorm.DB, code, name string) *models.Permission {
	t.Helper()

	perm := &models.Permission{
		ID:   uuid.New(),
		Code: code,
		Name: name,
	}
	require.NoError(t, tx.Create(perm).Error)
	return perm
}

func mustCreateAuthSession(t *testing.T, tx *gorm.DB, userID uuid.UUID, expiresAt time.Time) *models.AuthSession {
	t.Helper()

	session := &models.AuthSession{
		UserID:    userID,
		ExpiresAt: expiresAt,
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
	}
	require.NoError(t, tx.Create(session).Error)
	return session
}

func mustCreateRefreshToken(t *testing.T, tx *gorm.DB, sessionID, userID uuid.UUID, tokenHash string, expiresAt time.Time) *models.RefreshToken {
	t.Helper()

	token := &models.RefreshToken{
		SessionID: sessionID,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	require.NoError(t, tx.Create(token).Error)
	return token
}

func mustCreateUserToken(t *testing.T, tx *gorm.DB, userID uuid.UUID, tokenType enum.TokenType, tokenHash string, expiresAt time.Time) *models.UserToken {
	t.Helper()

	token := &models.UserToken{
		UserID:    userID,
		TokenType: tokenType,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	require.NoError(t, tx.Create(token).Error)
	return token
}

func mustCreateAuditLog(t *testing.T, tx *gorm.DB, action enum.ActionName, actorUserID *uuid.UUID, createdAt time.Time) *models.AuditLog {
	t.Helper()

	log := &models.AuditLog{
		Action:      action,
		ActorUserID: actorUserID,
		CreatedAt:   createdAt,
	}
	require.NoError(t, tx.Create(log).Error)
	return log
}

func activeUserRole(t *testing.T, tx *gorm.DB) *models.Role {
	t.Helper()
	return mustCreateRole(t, tx, constants.RoleCodeUser)
}
