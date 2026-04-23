package auth

import (
	"context"
	"errors"
	"portal-system/internal/domain/constants"
	"portal-system/internal/models"
	"portal-system/internal/platform/token"
	"portal-system/internal/repositories"
	"portal-system/internal/services"
	"testing"
	"time"

	"github.com/google/uuid"
)

type roleRepoMock struct {
	getWithPermissionsFn func(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
}

func (m *roleRepoMock) FindByCode(ctx context.Context, code constants.RoleCode) (*models.Role, error) {
	return nil, nil
}

func (m *roleRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return nil, nil
}

func (m *roleRepoMock) List(ctx context.Context) ([]models.Role, error) {
	return nil, nil
}

func (m *roleRepoMock) GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	if m.getWithPermissionsFn != nil {
		return m.getWithPermissionsFn(ctx, roleID)
	}
	return nil, nil
}

func (m *roleRepoMock) AssignPermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return nil
}

func (m *roleRepoMock) RemovePermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return nil
}

var _ repositories.RoleRepository = (*roleRepoMock)(nil)

type sessionRepoMock struct {
	findActiveByIDFn func(ctx context.Context, id uuid.UUID) (*models.AuthSession, error)
}

func (m *sessionRepoMock) Create(ctx context.Context, session *models.AuthSession) error {
	return nil
}

func (m *sessionRepoMock) FindActiveByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
	if m.findActiveByIDFn != nil {
		return m.findActiveByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *sessionRepoMock) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (m *sessionRepoMock) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *sessionRepoMock) ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthSession, error) {
	return nil, nil
}

var _ repositories.AuthSessionRepository = (*sessionRepoMock)(nil)

type revocationStoreMock struct {
	isRevokedFn func(ctx context.Context, sessionID uuid.UUID) (bool, error)
}

func (m *revocationStoreMock) MarkRevoked(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	return nil
}

func (m *revocationStoreMock) IsRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	if m.isRevokedFn != nil {
		return m.isRevokedFn(ctx, sessionID)
	}
	return false, nil
}

var _ services.SessionRevocationStore = (*revocationStoreMock)(nil)

func TestAuthenticator_Authenticate_Table(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	roleID := uuid.New()
	manager := token.New("test-secret", 3600)

	role := &models.Role{
		ID:   roleID,
		Code: constants.RoleCodeUser,
		Permissions: []models.Permission{
			{ID: uuid.New(), Code: "user.read", Name: "User Read"},
			{ID: uuid.New(), Code: "user.write", Name: "User Write"},
		},
	}

	makeToken := func(t *testing.T) string {
		t.Helper()
		tokenString, err := manager.GenerateAccessToken(token.GenerateAccessTokenInput{
			UserID:    userID,
			SessionID: sessionID,
			RoleID:    roleID,
			RoleCode:  string(constants.RoleCodeUser),
			Email:     "john@example.com",
			Username:  "john",
		})
		if err != nil {
			t.Fatalf("GenerateAccessToken() error = %v", err)
		}
		return tokenString
	}

	tests := []struct {
		name        string
		tokenString string
		revoked     bool
		revokeErr   error
		session     *models.AuthSession
		sessionErr  error
		roleErr     error
		expectedErr string
	}{
		{name: "invalid token", tokenString: "bad-token", expectedErr: "invalid token"},
		{name: "revocation store error ignored", tokenString: makeToken(t), revokeErr: errors.New("redis down"), session: &models.AuthSession{ID: sessionID, UserID: userID}},
		{name: "revoked session", tokenString: makeToken(t), revoked: true, expectedErr: "session is already revoked"},
		{name: "session lookup error", tokenString: makeToken(t), sessionErr: errors.New("not found"), expectedErr: "not found"},
		{name: "session user mismatch", tokenString: makeToken(t), session: &models.AuthSession{ID: sessionID, UserID: uuid.New()}, expectedErr: "session does not belong to user"},
		{name: "role lookup error", tokenString: makeToken(t), session: &models.AuthSession{ID: sessionID, UserID: userID}, roleErr: errors.New("role missing"), expectedErr: "role missing"},
		{name: "success", tokenString: makeToken(t), session: &models.AuthSession{ID: sessionID, UserID: userID}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authenticator := NewAuthenticator(
				manager,
				&roleRepoMock{
					getWithPermissionsFn: func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
						if tc.roleErr != nil {
							return nil, tc.roleErr
						}
						return role, nil
					},
				},
				&sessionRepoMock{
					findActiveByIDFn: func(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
						if tc.sessionErr != nil {
							return nil, tc.sessionErr
						}
						return tc.session, nil
					},
				},
				&revocationStoreMock{
					isRevokedFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return tc.revoked, tc.revokeErr
					},
				},
			)

			principal, err := authenticator.Authenticate(context.Background(), tc.tokenString)
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Fatalf("expected error %q, got %v", tc.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if principal == nil {
				t.Fatal("expected principal, got nil")
			}
			if principal.UserID != userID || principal.SessionID != sessionID || principal.RoleID != roleID {
				t.Fatalf("unexpected principal identity: %#v", principal)
			}
			if len(principal.Permissions) != 2 || principal.Permissions[0] != "user.read" || principal.Permissions[1] != "user.write" {
				t.Fatalf("unexpected permissions: %#v", principal.Permissions)
			}
		})
	}
}
