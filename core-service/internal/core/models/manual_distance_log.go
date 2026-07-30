package models

import (
	"time"

	"github.com/google/uuid"
)

type ManualDistanceLog struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ReminderID uuid.UUID `gorm:"type:uuid;not null;index" json:"reminder_id"`
	DistanceKM float64   `gorm:"not null" json:"distance_km"`
	Note       string    `gorm:"type:text" json:"note,omitempty"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	Reminder   ServiceReminder `gorm:"foreignKey:ReminderID" json:"-"`
}

func (ManualDistanceLog) TableName() string {
	return "M_MANUAL_DISTANCE_LOG"
}
