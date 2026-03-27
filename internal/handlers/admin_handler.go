package handlers

import (
	"errors"
	"net/http"
	"portal-system/internal/domain"
	"portal-system/internal/dto"
	"portal-system/internal/models"
	"portal-system/internal/services"

	"github.com/gin-gonic/gin"
)

const (
	DEFAULT_PAGE      int = 1
	DEFAULT_PAGE_SIZE int = 20
)

type AdminHandler struct {
	service *services.AdminService
}

func NewAdminHandler(svc *services.AdminService) *AdminHandler {
	return &AdminHandler{service: svc}
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

	result, err := h.service.ListUsers(c, input)

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
