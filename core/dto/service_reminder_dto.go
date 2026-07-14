package dto

import "time"

type CreateServiceReminderRequest struct {
	ServiceName string  `json:"service_name"`
	IntervalKM  float64 `json:"interval_km"`
}

type UpdateServiceReminderRequest struct {
	ServiceName *string  `json:"service_name,omitempty"`
	IntervalKM  *float64 `json:"interval_km,omitempty"`
}

type ManualDistanceRequest struct {
	DistanceKM float64 `json:"distance_km"`
	Note       string  `json:"note,omitempty"`
}

type ServiceReminderResponse struct {
	ID             string     `json:"id"`
	VehicleID      string     `json:"vehicle_id"`
	ServiceName    string     `json:"service_name"`
	IntervalKM     float64    `json:"interval_km"`
	AccumulatedKM  float64    `json:"accumulated_km"`
	ProgressPercent float64   `json:"progress_percent"`
	NeedsService   bool       `json:"needs_service"`
	NotifiedAt     *time.Time `json:"notified_at,omitempty"`
	LastServiceAt  *time.Time `json:"last_service_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ManualDistanceLogResponse struct {
	ID          string    `json:"id"`
	ReminderID  string    `json:"reminder_id"`
	DistanceKM  float64   `json:"distance_km"`
	Note        string    `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
