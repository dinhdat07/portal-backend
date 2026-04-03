package app

import (
	"portal-system/internal/config"
	"portal-system/internal/domain/enum"
	"portal-system/internal/http/handlers"
	"portal-system/internal/http/middleware"
	"portal-system/internal/platform/token"

	"github.com/gin-gonic/gin"
)

func setupRouter(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, adminHandler *handlers.AdminHandler, tokenManager *token.Manager, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/v1")
	authMiddleware := middleware.JWTAuth(tokenManager, cfg.AuthCookieName)
	csrfMiddleware := middleware.CSRFProtect(cfg)

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.RegisterUser)
		auth.POST("/login", authHandler.LogIn)
		auth.POST("/logout", authMiddleware, csrfMiddleware, authHandler.LogOut)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/resend-verification", authHandler.ResendVerification)
		auth.POST("/set-password", authHandler.SetPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
	}

	protected := api.Group("/")
	protected.Use(authMiddleware, csrfMiddleware)
	{
		users := protected.Group("/users")
		{
			me := users.Group("/me")
			{
				me.GET("", userHandler.GetMyProfile)
				me.PUT("", userHandler.UpdateProfile)
				me.PUT("/change-password", userHandler.ChangeMyPassword)

			}
		}

		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole(enum.RoleAdmin))
		{
			users := admin.Group("/users")
			{
				users.GET("", adminHandler.ListUsers)
				users.POST("", adminHandler.CreateUser)

				user := users.Group("/:userId")
				{
					user.GET("", adminHandler.GetUserDetail)
					user.PUT("", adminHandler.UpdateUser)
					user.DELETE("/delete", adminHandler.DeleteUser)
					user.PUT("/restore", adminHandler.RestoreUser)
					user.PUT("/role", adminHandler.UpdateRole)
				}
			}

		}
	}

	return r
}
