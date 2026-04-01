package middleware

import (
	"net/http"
	"portal-system/internal/platform/token"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(manager *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")

		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// use manager parser to parse tokenstring for authorize
		tokenString := parts[1]
		claims, err := manager.Parse(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorize format",
			})
			return
		}

		// set necessary field in gin context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)

		c.Next()
	}
}
