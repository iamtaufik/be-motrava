package models

import (
	"time"

	"github.com/google/uuid"
)

type TripPoint struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TripID     uuid.UUID `gorm:"type:uuid;not null;index" json:"trip_id"`
	Latitude   float64   `gorm:"not null" json:"latitude"`
	Longitude  float64   `gorm:"not null" json:"longitude"`
	Speed      float64   `gorm:"not null;default:0" json:"speed"`
	Heading    float64   `gorm:"not null;default:0" json:"heading"`
	Accuracy   float64   `gorm:"not null;default:0" json:"accuracy"`
	Altitude   float64   `gorm:"not null;default:0" json:"altitude"`
	Battery    int       `gorm:"not null;default:0" json:"battery"`
	RecordedAt time.Time `gorm:"not null" json:"recorded_at"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	Trip       Trip      `gorm:"foreignKey:TripID" json:"-"`
}

func (TripPoint) TableName() string {
	return "M_TRIP_POINT"
}
