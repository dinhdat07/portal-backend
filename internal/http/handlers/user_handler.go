package handlers

import (
	"errors"
	"net/http"
	"portal-system/internal/domain"
	"portal-system/internal/dto"
	"portal-system/internal/http/reqctx"
	"portal-system/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

func (h *UserHandler) GetMyProfile(c *gin.Context) {

	meta := reqctx.GetAuditMetaFromGin(c)
	actor, err := reqctx.GetActorFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), meta, actor, actor.ID)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot load user info",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))

}

func (h *UserHandler) ChangeMyPassword(c *gin.Context) {
	req := &dto.ChangePasswordRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	actor, err := reqctx.GetActorFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)

	if err := h.userService.ChangePassword(c.Request.Context(), meta, actor, req.CurrentPassword, req.NewPassword, req.ConfirmPassword); err != nil {
		switch {
		case errors.Is(err, services.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrIncorrectPassword):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrPasswordConfirmationMismatch),
			errors.Is(err, services.ErrNewPasswordMustBeDifferent),
			errors.Is(err, services.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot change password",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthMessageResponse{Message: "Password changed successfully"})

}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	req := &dto.UpdateUserRequest{}

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	var dob *time.Time
	if req.DOB != nil {
		dob = &req.DOB.Time
	}

	input := domain.UpdateUserInput{
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		DOB:       dob,
	}

	meta := reqctx.GetAuditMetaFromGin(c)
	actor, err := reqctx.GetActorFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	user, err := h.userService.UpdateProfile(c.Request.Context(), meta, actor, actor.ID, input)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrUsernameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot update user info",
			})
		}
		return
	}
	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}
