package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	SessionID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`

	TokenHash string `gorm:"type:varchar(255);not null;uniqueIndex"`

	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time

	ReplacedByTokenID *uuid.UUID `gorm:"type:uuid;index"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	r.ID = uuid.New()
	return nil
}
