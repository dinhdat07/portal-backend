package app

import (
	"portal-system/internal/auth"
	"portal-system/internal/domain/constants"
	"portal-system/internal/http/handlers"
	"portal-system/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func setupRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	adminHandler *handlers.AdminHandler,
	authenticator *auth.Authenticator,
	authorizer *auth.Authorizer) *gin.Engine {

	r := gin.Default()
	api := r.Group("/api/v1")

	authn := middleware.AuthenticationMiddleware(authenticator)

	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.RegisterUser)
		authRoutes.POST("/login", authHandler.LogIn)
		authRoutes.POST("/verify-email", authHandler.VerifyEmail)
		authRoutes.POST("/resend-verification", authHandler.ResendVerification)
		authRoutes.POST("/set-password", authHandler.SetPassword)
		authRoutes.POST("/reset-password", authHandler.ResetPassword)
		authRoutes.POST("/forgot-password", authHandler.ForgotPassword)

		authRoutes.POST("/refresh", authHandler.Refresh)
	}

	protected := api.Group("/")

	// authentication phase
	protected.Use(authn)
	{
		// protected auth routes
		protected.POST("/auth/logout", authHandler.Logout)
		protected.POST("/auth/logout-all", authHandler.LogoutAll)

		users := protected.Group("/users")
		{
			me := users.Group("/me")
			{
				// authorization phase by permission
				me.GET("", middleware.RequirePermission(authorizer, constants.PermProfileReadSelf), userHandler.GetMyProfile)
				me.PUT("", middleware.RequirePermission(authorizer, constants.PermProfileUpdateSelf), userHandler.UpdateProfile)
				me.PUT("/change-password", middleware.RequirePermission(authorizer, constants.PermProfileChangePassword), userHandler.ChangeMyPassword)
			}
		}

		admin := protected.Group("/admin")
		{
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", middleware.RequirePermission(authorizer, constants.PermUserList), adminHandler.ListUsers)
				adminUsers.POST("", middleware.RequirePermission(authorizer, constants.PermUserCreate), adminHandler.CreateUser)

				adminUser := adminUsers.Group("/:userId")
				{
					adminUser.GET("", middleware.RequirePermission(authorizer, constants.PermUserReadDetail), adminHandler.GetUserDetail)
					adminUser.PUT("", middleware.RequirePermission(authorizer, constants.PermUserUpdate), adminHandler.UpdateUser)
					adminUser.DELETE("/delete", middleware.RequirePermission(authorizer, constants.PermUserDelete), adminHandler.DeleteUser)
					adminUser.PUT("/restore", middleware.RequirePermission(authorizer, constants.PermUserRestore), adminHandler.RestoreUser)
					adminUser.PUT("/role", middleware.RequirePermission(authorizer, constants.PermUserRoleUpdate), adminHandler.UpdateRole)
				}
			}
		}
	}

	return r
}
