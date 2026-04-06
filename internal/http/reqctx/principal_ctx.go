package reqctx

import (
	"portal-system/internal/auth"

	"github.com/gin-gonic/gin"
)

type contextKey string

const principalKey contextKey = "principal"

func SetPrincipal(c *gin.Context, p *auth.Principal) {
	c.Set(string(principalKey), p)
}

func GetPrincipal(c *gin.Context) (*auth.Principal, bool) {
	val, exists := c.Get(string(principalKey))
	if !exists {
		return nil, false
	}

	p, ok := val.(*auth.Principal)
	return p, ok
}
