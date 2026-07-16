package dto

import "time"

type CreateVehicleRequest struct {
	VehicleName             string   `json:"vehicle_name"`
	PlateNumber             string   `json:"plate_number"`
	Brand                   string   `json:"brand"`
	Model                   string   `json:"model"`
	VehicleType             string   `json:"vehicle_type"`
	Color                   string   `json:"color"`
	Year                    *int     `json:"year,omitempty"`
	Photo                   string   `json:"photo,omitempty"`
	FuelEfficiencyKmPerLiter *float64 `json:"fuel_efficiency_km_per_liter,omitempty"`
	InitialKM               *float64 `json:"initial_km,omitempty"`
}

type UpdateVehicleRequest struct {
	VehicleName             *string  `json:"vehicle_name,omitempty"`
	PlateNumber             *string  `json:"plate_number,omitempty"`
	Brand                   *string  `json:"brand,omitempty"`
	Model                   *string  `json:"model,omitempty"`
	VehicleType             *string  `json:"vehicle_type,omitempty"`
	Color                   *string  `json:"color,omitempty"`
	Year                    *int     `json:"year,omitempty"`
	Photo                   *string  `json:"photo,omitempty"`
	FuelEfficiencyKmPerLiter *float64 `json:"fuel_efficiency_km_per_liter,omitempty"`
	InitialKM               *float64 `json:"initial_km,omitempty"`
}

type VehicleResponse struct {
	ID                      string    `json:"id"`
	UserID                  string    `json:"user_id"`
	VehicleName             string    `json:"vehicle_name"`
	PlateNumber             string    `json:"plate_number"`
	Brand                   string    `json:"brand"`
	Model                   string    `json:"model"`
	VehicleType             string    `json:"vehicle_type"`
	Color                   string    `json:"color"`
	Year                    *int      `json:"year,omitempty"`
	Photo                   *string   `json:"photo,omitempty"`
	IsDefault               bool      `json:"is_default"`
	TotalDistanceKM         float64   `json:"total_distance_km"`
	InitialKM               *float64  `json:"initial_km,omitempty"`
	LastRecordedOdometerKM  float64   `json:"last_recorded_odometer_km"`
	FuelEfficiencyKmPerLiter *float64 `json:"fuel_efficiency_km_per_liter,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
