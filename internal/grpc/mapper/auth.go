package mappers

import (
	authv1 "portal-system/gen/go/auth/v1"
	"portal-system/internal/domain"
)

func LoginResultToPB(result *domain.LoginResult) *authv1.LoginResponse {
	if result == nil {
		return nil
	}

	return &authv1.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(result.ExpiresIn),
		User:         UserModelToPB(result.User),
	}
}

func RefreshResultToPB(result *domain.RefreshResult) *authv1.RefreshResponse {
	if result == nil {
		return nil
	}

	return &authv1.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(result.ExpiresIn),
	}
}
