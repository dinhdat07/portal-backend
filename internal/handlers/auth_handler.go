package handlers

import (
	"errors"
	"net/http"
	"portal-system/internal/dto"
	"portal-system/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	req := &dto.RegisterRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Register(c.Request.Context(), req.Email, req.Username, req.Password, req.FirstName, req.LastName, req.Dob.Time)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailExists),
			errors.Is(err, services.ErrUsernameExists),
			errors.Is(err, services.ErrEmailBlacklisted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, dto.AuthMessageResponse{
		Message: "registration successful",
	})

}

func (h *AuthHandler) LogIn(c *gin.Context) {
	req := &dto.LoginRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	result, err := h.service.LogIn(
		c.Request.Context(),
		req.Identifier,
		req.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
			return

		case errors.Is(err, services.ErrAccountNotVerified):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "account not verified",
			})
			return

		case errors.Is(err, services.ErrAccountDeleted):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "account deleted",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   result.ExpiresIn,
		User:        dto.ToUserResponse(result.User),
	})
}
