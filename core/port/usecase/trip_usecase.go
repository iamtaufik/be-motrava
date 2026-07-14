package usecase

import (
	"motrava/core/dto"
	"motrava/core/models"
)

type TripUsecase interface {
	StartTrip(userID string, input dto.StartTripRequest) (*dto.TripResponse, error)
	ProcessLocation(userID string, tripID string, point models.TripPoint) error
	EndTrip(userID string, tripID string) (*dto.TripResponse, error)
	GetTripHistory(userID string, req dto.TripHistoryRequest) ([]dto.TripHistoryItem, *dto.PaginationMeta, error)
	GetTripDetail(userID string, tripID string) (*dto.TripDetailResponse, error)
}
