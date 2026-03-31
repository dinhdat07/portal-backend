package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLogFilter struct {
	Action       string
	ActorUserID  *uuid.UUID
	TargetUserID *uuid.UUID
	From         *time.Time
	To           *time.Time
	Page         int
	PageSize     int
}

type AuditMeta struct {
	IPAddress string
	UserAgent string
}
