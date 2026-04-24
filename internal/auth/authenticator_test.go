package auth_test

import (
	"context"
	"errors"
	"portal-system/internal/auth"
	"portal-system/internal/domain/constants"
	authmocks "portal-system/internal/mocks/auth"
	repositoriesmocks "portal-system/internal/mocks/repositories"
	"portal-system/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestAuthenticator_Authenticate_Table(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	roleID := uuid.New()

	role := &models.Role{
		ID:   roleID,
		Code: constants.RoleCodeUser,
		Permissions: []models.Permission{
			{ID: uuid.New(), Code: "user.read", Name: "User Read"},
			{ID: uuid.New(), Code: "user.write", Name: "User Write"},
		},
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
		{name: "revocation store error ignored", tokenString: "valid-token", revokeErr: errors.New("redis down"), session: &models.AuthSession{ID: sessionID, UserID: userID}},
		{name: "revoked session", tokenString: "valid-token", revoked: true, expectedErr: "session is already revoked"},
		{name: "session lookup error", tokenString: "valid-token", sessionErr: errors.New("not found"), expectedErr: "not found"},
		{name: "session user mismatch", tokenString: "valid-token", session: &models.AuthSession{ID: sessionID, UserID: uuid.New()}, expectedErr: "session does not belong to user"},
		{name: "role lookup error", tokenString: "valid-token", session: &models.AuthSession{ID: sessionID, UserID: userID}, roleErr: errors.New("role missing"), expectedErr: "role missing"},
		{name: "success", tokenString: "valid-token", session: &models.AuthSession{ID: sessionID, UserID: userID}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := authmocks.NewTokenIssuer(t)
			manager.EXPECT().Parse(tc.tokenString).RunAndReturn(func(tokenString string) (*auth.Claims, error) {
				if tokenString == "bad-token" {
					return nil, errors.New("invalid token")
				}
				return &auth.Claims{
					UserID:    userID,
					SessionID: sessionID,
					Username:  "john",
					Email:     "john@example.com",
					RoleID:    roleID,
					RoleCode:  string(constants.RoleCodeUser),
				}, nil
			}).Maybe()

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

			revoStore := authmocks.NewSessionRevocationStore(t)
			revoStore.EXPECT().IsRevoked(mock.Anything, sessionID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (bool, error) {
				return tc.revoked, tc.revokeErr
			}).Maybe()

			authenticator := auth.NewAuthenticator(manager, roleRepo, sessionRepo, revoStore)

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
