package usecase

import "motrava/core/dto"

type VehicleUsecase interface {
	ListVehicles(userID string) ([]dto.VehicleResponse, error)
	GetVehicleByID(id string, userID string) (*dto.VehicleResponse, error)
	CreateVehicle(userID string, input dto.CreateVehicleRequest) (*dto.VehicleResponse, error)
	UpdateVehicle(id string, userID string, input dto.UpdateVehicleRequest) (*dto.VehicleResponse, error)
	DeleteVehicle(id string, userID string) error
	SetDefaultVehicle(id string, userID string) (*dto.VehicleResponse, error)
}
