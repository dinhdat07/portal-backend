package dto

type PaginationMeta struct {
	Page     int   `json:"page" binding:"required"`
	PageSize int   `json:"page_size" binding:"required"`
	Total    int64 `json:"total" binding:"required"`
}

type PaginatedUsersResponse struct {
	Data []UserResponse `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
