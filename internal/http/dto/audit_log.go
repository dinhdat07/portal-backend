package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListActionLogsQuery struct {
	Action       string     `form:"action"`
	ActorUserID  *uuid.UUID `form:"actor_user_id"`
	TargetUserID *uuid.UUID `form:"target_user_id"`
	From         *time.Time `form:"from" time_format:"2006-01-02T15:04:05Z07:00"`
	To           *time.Time `form:"to" time_format:"2006-01-02T15:04:05Z07:00"`
	Page         int        `form:"page" binding:"omitempty,min=1"`
	PageSize     int        `form:"page_size" binding:"omitempty,min=1,max=100"`
}
