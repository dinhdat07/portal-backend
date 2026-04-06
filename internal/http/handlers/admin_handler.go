package handlers

import (
	"errors"
	"net/http"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/dto"
	"portal-system/internal/http/reqctx"
	"portal-system/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	DEFAULT_PAGE      int = 1
	DEFAULT_PAGE_SIZE int = 20
)

type AdminHandler struct {
	adminSvc *services.AdminService
	userSvc  *services.UserService
}

func NewAdminHandler(adminSvc *services.AdminService, userSvc *services.UserService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc, userSvc: userSvc}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	query := &dto.ListUsersQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid query",
		})
		return
	}

	if query.Page == 0 {
		query.Page = DEFAULT_PAGE
	}

	if query.PageSize == 0 {
		query.PageSize = DEFAULT_PAGE_SIZE
	}

	var dob *time.Time
	if query.Dob != nil {
		dob = &query.Dob.Time
	}

	includeDeleted := false
	if query.IncludeDeleted != nil {
		includeDeleted = *query.IncludeDeleted
	}

	input := domain.UsersFilter{
		Page:           query.Page,
		PageSize:       query.PageSize,
		Username:       query.Username,
		Email:          query.Email,
		FullName:       query.FullName,
		Dob:            dob,
		Role:           enum.UserRole(query.Role),
		Status:         enum.UserStatus(query.Status),
		IncludeDeleted: includeDeleted,
	}

	meta := reqctx.GetAuditMetaFromGin(c)
	actor, err := reqctx.GetActorFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	result, err := h.adminSvc.ListUsers(c.Request.Context(), meta, actor, input)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid query",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot list users",
			})
		}
		return
	}

	data := make([]dto.UserResponse, 0, len(result.Users))
	for _, u := range result.Users {
		data = append(data, dto.ToUserResponse(&u))
	}

	pageMeta := dto.PaginationMeta{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}

	c.JSON(http.StatusOK, dto.PaginatedUsersResponse{
		Data: data,
		Meta: pageMeta,
	})

}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	req := &dto.CreateUserRequest{}
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

	input := domain.CreateUserInput{
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}
	if req.DOB != nil {
		input.DOB = &req.DOB.Time
	}

	user, err := h.adminSvc.CreateUser(c.Request.Context(), meta, actor, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid input",
			})
		case errors.Is(err, services.ErrEmailExists),
			errors.Is(err, services.ErrUsernameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot create user",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ToUserResponse(user))
}
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userIDValue := c.Param("userId")
	if userIDValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is invalid",
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

	user, err := h.userSvc.GetProfile(c.Request.Context(), meta, actor, userID)

	if err != nil {
		switch {
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

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userIDValue := c.Param("userId")
	if userIDValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is invalid",
		})
		return
	}

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

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), meta, actor, userID, input)

	if err != nil {
		switch {
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

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDValue := c.Param("userId")
	if userIDValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is invalid",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)
	actor, err := reqctx.GetActorFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.adminSvc.DeleteUser(c.Request.Context(), meta, actor, userID)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot delete user info",
			})
		}
		return
	}
	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

func (h *AdminHandler) RestoreUser(c *gin.Context) {
	userIDValue := c.Param("userId")
	if userIDValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is invalid",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)
	actor, err := reqctx.GetActorFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.adminSvc.RestoreUser(c.Request.Context(), meta, actor, userID)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrUserNotDeleted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot restore user",
			})
		}
		return
	}
	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

func (h *AdminHandler) UpdateRole(c *gin.Context) {
	userIDValue := c.Param("userId")
	if userIDValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is invalid",
		})
		return
	}

	req := &dto.UpdateRoleRequest{}

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	meta := reqctx.GetAuditMetaFromGin(c)
	actor, err := reqctx.GetActorFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.adminSvc.UpdateRole(c.Request.Context(), meta, actor, userID, req.Role)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot update user role",
			})
		}
		return
	}
	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}
