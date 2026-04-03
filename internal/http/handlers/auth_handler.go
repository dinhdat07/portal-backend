package handlers

import (
	"errors"
	"net/http"
	"portal-system/internal/config"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/dto"
	"portal-system/internal/platform/token"
	"portal-system/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *services.AuthService
	config  *config.Config
}

func NewAuthHandler(service *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{service: service, config: cfg}
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func (h *AuthHandler) setAuthCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(parseSameSite(h.config.AuthCookieSameSite))
	c.SetCookie(
		h.config.AuthCookieName,
		token,
		maxAge,
		"/",
		h.config.AuthCookieDomain,
		h.config.AuthCookieSecure,
		true,
	)
}

func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	c.SetSameSite(parseSameSite(h.config.AuthCookieSameSite))
	c.SetCookie(
		h.config.AuthCookieName,
		"",
		-1,
		"/",
		h.config.AuthCookieDomain,
		h.config.AuthCookieSecure,
		true,
	)
}

func (h *AuthHandler) setCSRFCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(parseSameSite(h.config.AuthCookieSameSite))
	c.SetCookie(
		h.config.CSRFCookieName,
		value,
		maxAge,
		"/",
		h.config.AuthCookieDomain,
		h.config.AuthCookieSecure,
		false,
	)
}

func (h *AuthHandler) clearCSRFCookie(c *gin.Context) {
	c.SetSameSite(parseSameSite(h.config.AuthCookieSameSite))
	c.SetCookie(
		h.config.CSRFCookieName,
		"",
		-1,
		"/",
		h.config.AuthCookieDomain,
		h.config.AuthCookieSecure,
		false,
	)
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	req := &dto.RegisterRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	meta := getAuditMetaFromGin(c)

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

	meta := getAuditMetaFromGin(c)

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

	csrfToken, err := token.GenerateSecureToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	h.setAuthCookie(c, result.AccessToken, result.ExpiresIn)
	h.setCSRFCookie(c, csrfToken, result.ExpiresIn)

	c.JSON(http.StatusOK, dto.LoginResponse{
		ExpiresIn: result.ExpiresIn,
		User:      dto.ToUserResponse(result.User),
	})
}

func (h *AuthHandler) LogOut(c *gin.Context) {
	h.clearAuthCookie(c)
	h.clearCSRFCookie(c)
	c.JSON(http.StatusOK, dto.AuthMessageResponse{Message: "Logged out"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	req := &dto.VerifyEmailRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	meta := getAuditMetaFromGin(c)
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

	meta := getAuditMetaFromGin(c)

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

	meta := getAuditMetaFromGin(c)

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

	meta := getAuditMetaFromGin(c)

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

	meta := getAuditMetaFromGin(c)

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
