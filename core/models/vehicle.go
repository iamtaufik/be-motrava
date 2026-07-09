package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleType string

const (
	VehicleTypeCar  VehicleType = "CAR"
	VehicleTypeMotor VehicleType = "MOTOR"
)

type Vehicle struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	VehicleName string         `gorm:"size:255;not null" json:"vehicle_name"`
	PlateNumber string         `gorm:"size:50;not null" json:"plate_number"`
	Brand       string         `gorm:"size:100;not null" json:"brand"`
	Model       string         `gorm:"size:100;not null" json:"model"`
	VehicleType string         `gorm:"size:20;not null;default:CAR" json:"vehicle_type"`
	Color       string         `gorm:"size:50;not null" json:"color"`
	Year        *int           `json:"year,omitempty"`
	Photo       *string        `gorm:"type:text" json:"photo,omitempty"`
	IsDefault   bool           `gorm:"not null;default:false" json:"is_default"`
	CreatedAt   time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
}

func (Vehicle) TableName() string {
	return "M_VEHICLE"
}
