package app

import (
	"portal-system/internal/handlers"

	"github.com/gin-gonic/gin"
)

func setupRouter(authHandler *handlers.AuthHandler) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.LogIn)
		}
	}

	return r
}
