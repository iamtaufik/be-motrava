package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceReminder struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	VehicleID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	ServiceName     string         `gorm:"size:255;not null" json:"service_name"`
	IntervalKM      float64        `gorm:"not null" json:"interval_km"`
	AccumulatedKM   float64        `gorm:"not null;default:0" json:"accumulated_km"`
	NotifiedAt      *time.Time     `json:"notified_at,omitempty"`
	LastServiceAt   *time.Time     `json:"last_service_at,omitempty"`
	IsActive        bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt       time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Vehicle         Vehicle        `gorm:"foreignKey:VehicleID" json:"-"`
	User            User           `gorm:"foreignKey:UserID" json:"-"`
}

func (ServiceReminder) TableName() string {
	return "M_VEHICLE_SERVICE_REMINDER"
}
