package usecase

import (
	"motrava/core-service/internal/core/dto"
	"motrava/core-service/internal/core/models"
)

type TripUsecase interface {
	StartTrip(userID string, input dto.StartTripRequest) (*dto.TripResponse, error)
	ProcessLocation(userID string, tripID string, point models.TripPoint) error
	BatchLocations(userID string, tripID string, input []dto.BatchLocationRequest) (*dto.BatchLocationResponse, error)
	EndTrip(userID string, tripID string) (*dto.TripResponse, error)
	GetTripHistory(userID string, req dto.TripHistoryRequest) ([]dto.TripHistoryItem, *dto.PaginationMeta, error)
	GetTripDetail(userID string, tripID string) (*dto.TripDetailResponse, error)
	DeleteTrip(userID string, tripID string) error
}
