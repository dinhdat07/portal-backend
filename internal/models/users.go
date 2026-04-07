package models

import (
	"portal-system/internal/domain/enum"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email           string     `gorm:"size:255;uniqueIndex;not null"`
	Username        string     `gorm:"size:50;uniqueIndex;not null"`
	FirstName       string     `gorm:"size:100;not null"`
	LastName        string     `gorm:"size:100;not null"`
	DOB             *time.Time `gorm:"type:date"`
	PasswordHash    *string    `gorm:"size:255"`
	RoleID          uuid.UUID
	Role            Role            `gorm:"foreignKey:RoleID;references:ID"`
	Status          enum.UserStatus `gorm:"size:30;not null"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
	DeletedBy       *uuid.UUID     `gorm:"type:uuid"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
