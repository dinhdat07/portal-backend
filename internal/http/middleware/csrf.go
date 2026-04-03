package middleware

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"portal-system/internal/config"
	"strings"

	"github.com/gin-gonic/gin"
)

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func sameOrigin(origin string, allowed map[string]struct{}) bool {
	normalized := strings.TrimSpace(strings.ToLower(origin))
	if normalized == "" {
		return false
	}
	_, ok := allowed[normalized]
	return ok
}

func extractOrigin(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return strings.ToLower(u.Scheme + "://" + u.Host)
}

func CSRFProtect(cfg *config.Config) gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{}
	for _, raw := range []string{cfg.FrontEndUrl, cfg.ApiBaseUrl} {
		origin := extractOrigin(raw)
		if origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin != "" && !sameOrigin(origin, allowedOrigins) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid origin"})
			return
		}

		cookieToken, err := c.Cookie(cfg.CSRFCookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing csrf token"})
			return
		}

		headerToken := c.GetHeader(cfg.CSRFHeaderName)
		if headerToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing csrf token"})
			return
		}

		if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid csrf token"})
			return
		}

		c.Next()
	}
}
