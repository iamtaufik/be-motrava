package repository

import (
	"github.com/google/uuid"

	"motrava/core-service/internal/core/models"
)

type TripRepository interface {
	FindByID(id uuid.UUID) (*models.Trip, error)
	FindOngoingByUserID(userID uuid.UUID) (*models.Trip, error)
	FindByUserID(userID uuid.UUID, page, limit int, search, dateFrom, dateTo string) ([]models.Trip, int64, error)
	FindByVehicleID(vehicleID uuid.UUID) ([]models.Trip, error)
	Create(trip *models.Trip) error
	Save(trip *models.Trip) error
	Delete(id uuid.UUID) error
	DeleteByVehicleID(vehicleID uuid.UUID) error
	SumDistanceByVehicleID(vehicleID uuid.UUID) (float64, error)
}

type TripPointRepository interface {
	Create(point *models.TripPoint) error
	FindAllByTripID(tripID uuid.UUID) ([]models.TripPoint, error)
	FindLastByTripID(tripID uuid.UUID) (*models.TripPoint, error)
	CountByTripID(tripID uuid.UUID) (int64, error)
	DeleteByTripID(tripID uuid.UUID) error
}
