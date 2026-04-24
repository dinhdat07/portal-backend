package mappers

import (
	adminv1 "portal-system/gen/go/admin/v1"
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/domain"
)

func ListUsersResultToPB(result *domain.ListUsersResult) *adminv1.ListUsersResponse {
	if result == nil {
		return nil
	}

	data := make([]*commonv1.User, 0, len(result.Users))
	for i := range result.Users {
		data = append(data, UserModelToPB(&result.Users[i]))
	}

	return &adminv1.ListUsersResponse{
		Data: data,
		Meta: &commonv1.PaginationMeta{
			Page:     int32(result.Page),
			PageSize: int32(result.PageSize),
			Total:    result.Total,
		},
	}
}
