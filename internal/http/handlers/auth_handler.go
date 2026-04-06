package handlers

import (
	"errors"
	"net/http"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/dto"
	"portal-system/internal/http/reqctx"
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

	meta := reqctx.GetAuditMetaFromGin(c)

	err := h.service.Register(c.Request.Context(), meta, req.Email, req.Username, req.Password, req.FirstName, req.LastName, req.Dob.Time)

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
		Message: "Registration successful",
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

	meta := reqctx.GetAuditMetaFromGin(c)

	result, err := h.service.LogIn(
		c.Request.Context(),
		meta,
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

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	req := &dto.VerifyEmailRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)
	err := h.service.VerifyEmail(c.Request.Context(), meta, req.Token, enum.TokenTypeEmailVerification)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid or expired token",
			})
		case errors.Is(err, services.ErrUserAlreadyDeleted), errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusConflict, gin.H{
				"message": "user already deleted",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AuthMessageResponse{
		Message: "Email verification successful",
	})

}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	req := &dto.ResendVerificationRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)

	if err := h.service.ResendVerification(c.Request.Context(), meta, req.Email, enum.TokenTypeEmailVerification); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid or expired token",
			})
		case errors.Is(err, services.ErrUserAlreadyDeleted), errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusConflict, gin.H{
				"message": "user already deleted",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AuthMessageResponse{
		Message: "Resend verification successfully",
	})

}

func (h *AuthHandler) SetPassword(c *gin.Context) {
	req := &dto.SetPasswordRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)

	input := &domain.SetPasswordInput{
		Token:           req.Token,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}
	err := h.service.SetPassword(c.Request.Context(), meta, input, enum.TokenTypePasswordSet)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken), errors.Is(err, services.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid or expired token",
			})
		case errors.Is(err, services.ErrUserAlreadyDeleted), errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusConflict, gin.H{
				"message": "user not found or already deleted",
			})

		case errors.Is(err, services.ErrPasswordConfirmationMismatch),
			errors.Is(err, services.ErrPasswordAlreadySet):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot set password",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AuthMessageResponse{
		Message: "Email verification and password set successful",
	})

}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	req := &dto.SetPasswordRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)

	input := &domain.SetPasswordInput{
		Token:           req.Token,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}

	err := h.service.ResetPassword(
		c.Request.Context(),
		meta,
		input,
		enum.TokenTypePasswordReset,
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken),
			errors.Is(err, services.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid or expired token",
			})

		case errors.Is(err, services.ErrUserAlreadyDeleted),
			errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusConflict, gin.H{
				"message": "user not found or already deleted",
			})

		case errors.Is(err, services.ErrPasswordConfirmationMismatch):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot reset password",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthMessageResponse{
		Message: "Password reset successful",
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	req := &dto.ForgotPasswordRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)

	err := h.service.ForgotPassword(c.Request.Context(), meta, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cannot process forgot password",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AuthMessageResponse{
		Message: "If the account exists, a password reset email has been sent",
	})
}
