package reqctx

import (
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAuditMetaFromGin(c *gin.Context) *domain.AuditMeta {
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	return &domain.AuditMeta{
		IPAddress: ip,
		UserAgent: userAgent,
	}
}

func GetActorFromGin(c *gin.Context) (*models.User, error) {
	userIDValue, ok := c.Get("user_id")
	if !ok {
		return nil, errors.New("missing user_id in context")
	}

	usernameValue, ok := c.Get("username")
	if !ok {
		return nil, errors.New("missing username in context")
	}

	emailValue, ok := c.Get("email")
	if !ok {
		return nil, errors.New("missing email in context")
	}

	roleValue, ok := c.Get("role")
	if !ok {
		return nil, errors.New("missing role in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, errors.New("invalid user_id type")
	}

	username, ok := usernameValue.(string)
	if !ok {
		return nil, errors.New("invalid username type")
	}

	email, ok := emailValue.(string)
	if !ok {
		return nil, errors.New("invalid email type")
	}

	role, ok := roleValue.(enum.UserRole)
	if !ok {
		return nil, errors.New("invalid role type")
	}

	return &models.User{
		ID:       userID,
		Username: username,
		Email:    email,
		Role:     role,
	}, nil
}
