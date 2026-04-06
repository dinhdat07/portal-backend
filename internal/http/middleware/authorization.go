package middleware

import (
	"net/http"
	"portal-system/internal/auth"
	"portal-system/internal/http/reqctx"

	"github.com/gin-gonic/gin"
)

func RequirePermission(authorizer *auth.Authorizer, perm auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := reqctx.GetPrincipal(c)

		if !ok || principal == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if !authorizer.HasPermission(principal, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}
