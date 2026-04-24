package middleware

import (
	"errors"
	"net/http"
	"portal-system/internal/auth"
	"portal-system/internal/http/reqctx"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthenticationMiddleware(authenticator *auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerTokenFromGin(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorize format"})
			return
		}

		principal, err := authenticator.Authenticate(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorize format"})
			return
		}

		reqctx.SetPrincipal(c, principal)
		c.Next()
	}
}

func extractBearerTokenFromGin(c *gin.Context) (string, error) {
	auth := c.GetHeader("Authorization")

	if auth == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return "", errors.New("missing token")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return "", errors.New("invalid token")
	}

	tokenString := parts[1]

	return tokenString, nil
}
