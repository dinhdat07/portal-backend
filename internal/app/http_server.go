package app

import (
	"portal-system/internal/domain/constants"
	"portal-system/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func (a *App) NewHTTPServer() *gin.Engine {

	r := gin.Default()

	r.Use(middleware.RecoveryMiddleware())
	api := r.Group("/api/v1")

	authn := middleware.AuthenticationMiddleware(a.Authenticator)

	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/register", a.AuthHandler.RegisterUser)
		authRoutes.POST("/login", a.AuthHandler.LogIn)
		authRoutes.POST("/verify-email", a.AuthHandler.VerifyEmail)
		authRoutes.POST("/resend-verification", a.AuthHandler.ResendVerification)
		authRoutes.POST("/set-password", a.AuthHandler.SetPassword)
		authRoutes.POST("/reset-password", a.AuthHandler.ResetPassword)
		authRoutes.POST("/forgot-password", a.AuthHandler.ForgotPassword)

		authRoutes.POST("/refresh", a.AuthHandler.Refresh)
	}

	protected := api.Group("/")

	// authentication phase
	protected.Use(authn)
	{
		// protected auth routes
		protected.POST("/auth/logout", a.AuthHandler.Logout)
		protected.POST("/auth/logout-all", a.AuthHandler.LogoutAll)

		users := protected.Group("/users")
		{
			me := users.Group("/me")
			{
				// authorization phase by permission
				me.GET("", middleware.RequirePermission(a.Authorizer, constants.PermProfileReadSelf), a.UserHandler.GetMyProfile)
				me.PUT("", middleware.RequirePermission(a.Authorizer, constants.PermProfileUpdateSelf), a.UserHandler.UpdateProfile)
				me.PUT("/change-password", middleware.RequirePermission(a.Authorizer, constants.PermProfileChangePassword), a.UserHandler.ChangeMyPassword)
			}
		}

		admin := protected.Group("/admin")
		{
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", middleware.RequirePermission(a.Authorizer, constants.PermUserList), a.AdminHandler.ListUsers)
				adminUsers.POST("", middleware.RequirePermission(a.Authorizer, constants.PermUserCreate), a.AdminHandler.CreateUser)

				adminUser := adminUsers.Group("/:userId")
				{
					adminUser.GET("", middleware.RequirePermission(a.Authorizer, constants.PermUserReadDetail), a.AdminHandler.GetUserDetail)
					adminUser.PUT("", middleware.RequirePermission(a.Authorizer, constants.PermUserUpdate), a.AdminHandler.UpdateUser)
					adminUser.DELETE("/delete", middleware.RequirePermission(a.Authorizer, constants.PermUserDelete), a.AdminHandler.DeleteUser)
					adminUser.PUT("/restore", middleware.RequirePermission(a.Authorizer, constants.PermUserRestore), a.AdminHandler.RestoreUser)
					adminUser.PUT("/role", middleware.RequirePermission(a.Authorizer, constants.PermUserRoleUpdate), a.AdminHandler.UpdateRole)
				}
			}
		}
	}

	return r
}

func RunHTTP(addr string, httpServer *gin.Engine) error {
	return httpServer.Run(":" + addr)
}
