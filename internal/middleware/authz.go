package middleware

import (
	"net/http"
	"portal-system/internal/models"

	"github.com/gin-gonic/gin"
)

func RequireRole(role models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden",
			})
			return
		}
		c.Next()
	}
}
