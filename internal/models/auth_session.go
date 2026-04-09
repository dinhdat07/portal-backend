package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthSession struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;index;not null"`

	RefreshTokenHash string    `gorm:"size:255;uniqueIndex;not null"`
	ExpiresAt        time.Time `gorm:"not null"`

	UserAgent string `gorm:"type:text"`
	IPAddress string `gorm:"type:varchar(45)"`

	LastUsedAt *time.Time
	RevokedAt  *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a *AuthSession) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()
	return nil
}
