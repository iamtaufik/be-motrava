package usecase

import (
	"motrava/core/dto"
	"motrava/core/models"
)

type TripUsecase interface {
	StartTrip(userID string, input dto.StartTripRequest) (*dto.TripResponse, error)
	ProcessLocation(userID string, tripID string, point models.TripPoint, speed float64) error
	EndTrip(userID string, tripID string) (*dto.TripResponse, error)
}
