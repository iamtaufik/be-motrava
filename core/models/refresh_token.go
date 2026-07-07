package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken stores server-side refresh token state for rotation and revocation.
type RefreshToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string         `gorm:"size:128;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time     `json:"revoked_at,omitempty"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
}

func (RefreshToken) TableName() string {
	return "M_REFRESH_TOKEN"
}
