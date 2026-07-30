package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TripStatusOngoing   = "ONGOING"
	TripStatusCompleted = "COMPLETED"
)

type Trip struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	VehicleID      uuid.UUID      `gorm:"type:uuid;not null" json:"vehicle_id"`
	StartTime      time.Time      `gorm:"not null" json:"start_time"`
	EndTime        *time.Time     `json:"end_time,omitempty"`
	StartLatitude  float64        `gorm:"not null" json:"start_latitude"`
	StartLongitude float64        `gorm:"not null" json:"start_longitude"`
	EndLatitude    *float64       `json:"end_latitude,omitempty"`
	EndLongitude   *float64       `json:"end_longitude,omitempty"`
	StartAddress   *string        `gorm:"type:text" json:"start_address,omitempty"`
	EndAddress     *string        `gorm:"type:text" json:"end_address,omitempty"`
	TotalDistance  float64        `gorm:"not null;default:0" json:"total_distance"`
	Duration       int            `gorm:"not null;default:0" json:"duration"`
	MovingTime     int            `gorm:"not null;default:0" json:"moving_time"`
	IdleTime       int            `gorm:"not null;default:0" json:"idle_time"`
	AverageSpeed   float64        `gorm:"not null;default:0" json:"average_speed"`
	MaximumSpeed   float64        `gorm:"not null;default:0" json:"maximum_speed"`
	FuelConsumed   *float64       `gorm:"type:double precision" json:"fuel_consumed,omitempty"`
	Status         string         `gorm:"size:20;not null;default:ONGOING" json:"status"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	User           User           `gorm:"foreignKey:UserID" json:"-"`
	Vehicle        Vehicle        `gorm:"foreignKey:VehicleID" json:"-"`
}

func (Trip) TableName() string {
	return "M_TRIP"
}
