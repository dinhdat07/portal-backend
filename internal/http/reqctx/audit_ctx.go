package reqctx

import (
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"

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

	role := constants.RoleCode(principal.RoleCode)
	if !role.IsValid() {
		return nil, errors.New("invalid role in principal")
	}

	return &domain.AuditUser{
		ID:       principal.UserID,
		Username: principal.Username,
		Email:    principal.Email,
		RoleCode: role,
	}, nil
}
