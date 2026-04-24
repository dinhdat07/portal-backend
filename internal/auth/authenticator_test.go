package auth

import (
	"context"
	"errors"
	"portal-system/internal/domain/constants"
	repositoriesmocks "portal-system/internal/mocks/repositories"
	servicesmocks "portal-system/internal/mocks/services"
	"portal-system/internal/models"
	"portal-system/internal/platform/token"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

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
			roleRepo := repositoriesmocks.NewRoleRepository(t)
			roleRepo.EXPECT().GetWithPermissions(mock.Anything, roleID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
				if tc.roleErr != nil {
					return nil, tc.roleErr
				}
				return role, nil
			}).Maybe()

			sessionRepo := repositoriesmocks.NewAuthSessionRepository(t)
			sessionRepo.EXPECT().FindActiveByID(mock.Anything, sessionID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
				if tc.sessionErr != nil {
					return nil, tc.sessionErr
				}
				return tc.session, nil
			}).Maybe()

			revoStore := servicesmocks.NewSessionRevocationStore(t)
			revoStore.EXPECT().IsRevoked(mock.Anything, sessionID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (bool, error) {
				return tc.revoked, tc.revokeErr
			}).Maybe()

			authenticator := NewAuthenticator(manager, roleRepo, sessionRepo, revoStore)

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
