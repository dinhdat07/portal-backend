package middleware

import (
	"net/http"
	"portal-system/internal/auth"
	"portal-system/internal/domain/constants"
	"portal-system/internal/http/reqctx"

	"github.com/gin-gonic/gin"
)

func RequirePermission(authorizer *auth.Authorizer, perm constants.PermissionCode) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authorizer == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		principal, ok := reqctx.GetPrincipal(c)
		if !ok || principal == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		allowed := authorizer.HasPermission(c.Request.Context(), principal, perm)

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}
