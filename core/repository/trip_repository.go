package repository

import (
	"github.com/google/uuid"

	"motrava/core/models"
)

type TripRepository interface {
	FindByID(id uuid.UUID) (*models.Trip, error)
	FindOngoingByUserID(userID uuid.UUID) (*models.Trip, error)
	FindByUserID(userID uuid.UUID, page, limit int, search, dateFrom, dateTo string) ([]models.Trip, int64, error)
	Create(trip *models.Trip) error
	Save(trip *models.Trip) error
}

type TripPointRepository interface {
	Create(point *models.TripPoint) error
	FindAllByTripID(tripID uuid.UUID) ([]models.TripPoint, error)
	FindLastByTripID(tripID uuid.UUID) (*models.TripPoint, error)
	CountByTripID(tripID uuid.UUID) (int64, error)
}
