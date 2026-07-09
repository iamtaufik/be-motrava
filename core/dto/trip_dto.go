package dto

import "time"

type StartTripRequest struct {
	VehicleID string `json:"vehicle_id"`
}

type TripResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	VehicleID      string     `json:"vehicle_id"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	StartLatitude  float64    `json:"start_latitude"`
	StartLongitude float64    `json:"start_longitude"`
	EndLatitude    *float64   `json:"end_latitude,omitempty"`
	EndLongitude   *float64   `json:"end_longitude,omitempty"`
	StartAddress   *string    `json:"start_address,omitempty"`
	EndAddress     *string    `json:"end_address,omitempty"`
	TotalDistance  float64    `json:"total_distance"`
	Duration       int        `json:"duration"`
	MovingTime     int        `json:"moving_time"`
	IdleTime       int        `json:"idle_time"`
	AverageSpeed   float64    `json:"average_speed"`
	MaximumSpeed   float64    `json:"maximum_speed"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TripHistoryRequest struct {
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	Search   string `json:"search"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

type TripHistoryItem struct {
	ID            string     `json:"id"`
	VehicleName   string     `json:"vehicle_name"`
	PlateNumber   string     `json:"plate_number"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	StartAddress  *string    `json:"start_address,omitempty"`
	EndAddress    *string    `json:"end_address,omitempty"`
	TotalDistance float64    `json:"total_distance"`
	Duration      int        `json:"duration"`
	MovingTime    int        `json:"moving_time"`
	IdleTime      int        `json:"idle_time"`
	AverageSpeed  float64    `json:"average_speed"`
	MaximumSpeed  float64    `json:"maximum_speed"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}
