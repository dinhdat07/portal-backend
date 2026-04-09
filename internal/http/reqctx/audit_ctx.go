package reqctx

import (
	"errors"
	"portal-system/internal/domain"

	"github.com/gin-gonic/gin"
)

func GetAuditMetaFromGin(c *gin.Context) *domain.AuditMeta {
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	return &domain.AuditMeta{
		IPAddress: ip,
		UserAgent: userAgent,
	}
}

func GetActorFromGin(c *gin.Context) (*domain.AuditUser, error) {
	principal, exists := GetPrincipal(c)
	if principal == nil || !exists {
		return nil, errors.New("missing principal in context")
	}

	return &domain.AuditUser{
		ID:       principal.UserID,
		Username: principal.Username,
		Email:    principal.Email,
		RoleCode: principal.RoleCode,
	}, nil
}
