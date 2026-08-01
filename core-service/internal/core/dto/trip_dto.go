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
	FuelConsumed   *float64   `json:"fuel_consumed,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type BatchLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Speed     float64 `json:"speed"`
	Heading   float64 `json:"heading"`
	Accuracy  float64 `json:"accuracy"`
	Altitude  float64 `json:"altitude"`
	Battery   int     `json:"battery"`
	Timestamp string  `json:"timestamp"`
}

type BatchLocationResponse struct {
	ProcessedCount int    `json:"processed_count"`
	TripID         string `json:"trip_id"`
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
	VehicleType   string     `json:"vehicle_type"`
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
	FuelConsumed  *float64   `json:"fuel_consumed,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}

type TripDetailResponse struct {
	ID             string         `json:"id"`
	UserID         string         `json:"user_id"`
	VehicleID      string         `json:"vehicle_id"`
	VehicleName    string         `json:"vehicle_name"`
	PlateNumber    string         `json:"plate_number"`
	VehicleType    string         `json:"vehicle_type"`
	StartTime      time.Time      `json:"start_time"`
	EndTime        *time.Time     `json:"end_time,omitempty"`
	StartLatitude  float64        `json:"start_latitude"`
	StartLongitude float64        `json:"start_longitude"`
	EndLatitude    *float64       `json:"end_latitude,omitempty"`
	EndLongitude   *float64       `json:"end_longitude,omitempty"`
	StartAddress   *string        `json:"start_address,omitempty"`
	EndAddress     *string        `json:"end_address,omitempty"`
	TotalDistance  float64        `json:"total_distance"`
	Duration       int            `json:"duration"`
	MovingTime     int            `json:"moving_time"`
	IdleTime       int            `json:"idle_time"`
	AverageSpeed   float64        `json:"average_speed"`
	MaximumSpeed   float64        `json:"maximum_speed"`
	FuelConsumed   *float64       `json:"fuel_consumed,omitempty"`
	Status         string         `json:"status"`
	Route          []PointPayload `json:"route"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type PointPayload struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Speed      float64 `json:"speed"`
	Heading    float64 `json:"heading"`
	Accuracy   float64 `json:"accuracy"`
	Altitude   float64 `json:"altitude"`
	Battery    int     `json:"battery"`
	RecordedAt string  `json:"recorded_at"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}
