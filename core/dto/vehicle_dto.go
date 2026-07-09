package dto

import "time"

type CreateVehicleRequest struct {
	VehicleName string `json:"vehicle_name"`
	PlateNumber string `json:"plate_number"`
	Brand       string `json:"brand"`
	Model       string `json:"model"`
	VehicleType string `json:"vehicle_type"`
	Color       string `json:"color"`
	Year        *int   `json:"year,omitempty"`
	Photo       string `json:"photo,omitempty"`
}

type UpdateVehicleRequest struct {
	VehicleName *string `json:"vehicle_name,omitempty"`
	PlateNumber *string `json:"plate_number,omitempty"`
	Brand       *string `json:"brand,omitempty"`
	Model       *string `json:"model,omitempty"`
	VehicleType *string `json:"vehicle_type,omitempty"`
	Color       *string `json:"color,omitempty"`
	Year        *int    `json:"year,omitempty"`
	Photo       *string `json:"photo,omitempty"`
}

type VehicleResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	VehicleName string    `json:"vehicle_name"`
	PlateNumber string    `json:"plate_number"`
	Brand       string    `json:"brand"`
	Model       string    `json:"model"`
	VehicleType string    `json:"vehicle_type"`
	Color       string    `json:"color"`
	Year        *int      `json:"year,omitempty"`
	Photo       *string   `json:"photo,omitempty"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
