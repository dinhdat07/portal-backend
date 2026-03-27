package handlers

import (
	"errors"
	"net/http"
	"portal-system/internal/domain"
	"portal-system/internal/dto"
	"portal-system/internal/models"
	"portal-system/internal/services"

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
	}

	if query.Page == 0 {
		query.Page = DEFAULT_PAGE
	}

	if query.PageSize == 0 {
		query.PageSize = DEFAULT_PAGE_SIZE
	}

	input := domain.ListUsersInput{
		Page:           query.Page,
		PageSize:       query.PageSize,
		Username:       query.Username,
		Email:          query.Email,
		FullName:       query.Email,
		Dob:            &query.Dob.Time,
		Role:           models.UserRole(query.Role),
		Status:         models.UserStatus(query.Status),
		IncludeDeleted: *query.IncludeDeleted,
	}

	result, err := h.adminSvc.ListUsers(c.Request.Context(), input)

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
	}

	data := make([]dto.UserResponse, 0, len(result.Users))
	for _, u := range result.Users {
		data = append(data, dto.ToUserResponse(&u))
	}

	meta := dto.PaginationMeta{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}

	c.JSON(http.StatusOK, dto.PaginatedUsersResponse{
		Data: data,
		Meta: meta,
	})

}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	req := &dto.CreateUserRequest{}

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
	}

	input := domain.CreateUserInput{
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		DOB:       &req.DOB.Time,
		Role:      models.UserRole(req.Role),
	}

	user, err := h.adminSvc.CreateUser(c.Request.Context(), input)
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
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
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

	user, err := h.userSvc.GetProfile(c.Request.Context(), userID)

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
	}

	input := domain.UpdateUserInput{
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		DOB:       &req.DOB.Time,
	}

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), userID, input)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
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

	adminIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	adminID, ok := adminIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid admin id in token",
		})
		return
	}

	user, err := h.adminSvc.DeleteUser(c.Request.Context(), userID, adminID)

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

	user, err := h.adminSvc.RestoreUser(c.Request.Context(), userID)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
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
	}

	user, err := h.adminSvc.UpdateRole(c.Request.Context(), userID, req.Role)

	if err != nil {
		switch {
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
