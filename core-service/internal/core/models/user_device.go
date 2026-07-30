package models

import (
	"time"

	"github.com/google/uuid"
)

type UserDevice struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	DeviceToken string    `gorm:"type:text;not null" json:"device_token"`
	Platform    string    `gorm:"size:20;not null;default:android" json:"platform"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
}

func (UserDevice) TableName() string {
	return "M_USER_DEVICE"
}
