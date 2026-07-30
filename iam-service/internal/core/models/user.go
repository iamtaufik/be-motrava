package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AuthProviderLocal  = "LOCAL"
	AuthProviderGoogle = "GOOGLE"
)

// User maps to public."M_USER".
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email        string         `gorm:"size:255;not null" json:"email"`
	FullName     *string        `gorm:"size:255" json:"full_name,omitempty"`
	Password     *string        `gorm:"type:text" json:"-"`
	AuthProvider string         `gorm:"size:20;not null;default:LOCAL" json:"auth_provider"`
	GoogleSub    *string        `gorm:"size:255" json:"google_sub,omitempty"`
	AvatarURL    *string        `gorm:"type:text" json:"avatar_url,omitempty"`
	PhoneNumber  *string        `gorm:"size:30" json:"phone_number,omitempty"`
	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
	IsVerified   bool           `gorm:"not null;default:false" json:"is_verified"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "M_USER"
}
