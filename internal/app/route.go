package app

import (
	"portal-system/internal/handlers"
	"portal-system/internal/middleware"
	"portal-system/internal/token"

	"github.com/gin-gonic/gin"
)

func setupRouter(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, tokenManager *token.Manager) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/v1")
	authMiddleware := middleware.JWTAuth(tokenManager)

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.RegisterUser)
		auth.POST("/login", authHandler.LogIn)
	}

	protected := api.Group("/")
	protected.Use(authMiddleware)
	{
		users := protected.Group("/users")
		{
			me := users.Group("/me")
			{
				me.GET("", userHandler.GetMyProfile)
				me.PUT("/change-password", userHandler.ChangeMyPassword)
			}
		}
	}

	return r
}
